// profiler.go — المحرك الأول: طلب استكشافي واحد + تقييم عادل للدفاعات (مؤشرات لا يقين).
package nexus

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// DefenseReport نتيجة التشخيص (الحقيقة كاملة).
type DefenseReport struct {
	URL           string   `json:"url"`
	Host          string   `json:"host"`
	Status        int      `json:"status"`
	Server        string   `json:"server"`
	Markers       []string `json:"markers"`
	TLSCipher     string   `json:"tls_cipher"`
	TLSVersion    string   `json:"tls_version"`
	LatencyMS     int64    `json:"latency_ms"`
	RiskLevel     int      `json:"risk_level"`     // 1..5
	RecommendTier string   `json:"recommend_tier"` // direct|datacenter|residential
	HonestyNote   string   `json:"honesty_note"`
}

const probeHonesty = "مؤشرات إحصائية لا يقين: التقييم ينفذ من خصائص الرد المرصودة فقط — لا بديلًا عن اختبارك اليدوي"

// WAFHeaders مرويسات دفاعية شائعة (هرم المعرفة العامة).
var wafHeaders = map[string]string{
	"Cf-Mitigated":             "cloudflare",
	"Cf-Ray":                   "cloudflare-edge",
	"X-Sucuri-Id":              "sucuri",
	"X-Waf-Status":             "generic-waf",
	"X-Akamai":                 "akamai",
	"X-Fastly-Error":           "fastly-waf",
	"X-Protected-By":           "generic-shield",
	"X-Mod-Security-Exclusion": "modsecurity",
}

var challengeRe = regexp.MustCompile(`(?i)(cdn-cgi/challenge-platform|__cf_chl_jschl|turnstile|recaptcha/api|h-captcha|denied by waf|access denied)`)

// Profile يرسل طلبًا استكشافيًا واحدًا ويشخصه.
func Profile(ctx context.Context, client *http.Client, urlStr string) (*DefenseReport, error) {
	raw := strings.TrimSpace(urlStr)
	if raw == "" {
		return nil, fmt.Errorf("رابط فارغ")
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("رابط غير صالح")
	}
	if client == nil {
		client = http.DefaultClient
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, raw, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AkdiroNexus/0.1 (+target-aware-probe; owner=Baran)")

	t0 := time.Now()
	resp, err := client.Do(req)
	lat := time.Since(t0).Milliseconds()
	if err != nil {
		return nil, fmt.Errorf("الفحص فشل في الاتصال: %w", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 256*1024))

	r := &DefenseReport{
		URL: raw, Host: u.Hostname(), Status: resp.StatusCode,
		Server: resp.Header.Get("Server"), LatencyMS: lat,
		Markers: []string{}, HonestyNote: probeHonesty,
	}

	for h, tag := range wafHeaders {
		if v := resp.Header.Get(h); v != "" {
			r.Markers = append(r.Markers, tag)
		}
	}
	if m := challengeRe.Find(body); m != nil {
		r.Markers = append(r.Markers, "challenge-script")
	}
	r.TLSCipher, r.TLSVersion = "", "plain"
	if resp.TLS != nil {
		r.TLSCipher = tls.CipherSuiteName(resp.TLS.CipherSuite)
		r.TLSVersion = tlsVersionName(resp.TLS.Version)
	}

	// التقييم العادل — بلا سحر
	risk := 1
	for _, m := range r.Markers {
		switch m {
		case "challenge-script", "turnstile":
			risk += 2
		default:
			risk++
		}
	}
	if resp.StatusCode == 429 || resp.StatusCode == 403 {
		risk++
	}
	if resp.StatusCode >= 500 {
		risk++
	}
	if risk > 5 {
		risk = 5
	}
	r.RiskLevel = risk
	switch {
	case risk <= 2:
		r.RecommendTier = "direct"
	case risk == 3:
		r.RecommendTier = "datacenter"
	default:
		r.RecommendTier = "residential"
	}
	return r, nil
}

func tlsOf(resp *http.Response) string {
	if resp.TLS != nil {
		return tlsVersionName(resp.TLS.Version)
	}
	return "plain"
}

func tlsVersionName(v uint16) string {
	switch v {
	case tls.VersionTLS10:
		return "TLS1.0"
	case tls.VersionTLS11:
		return "TLS1.1"
	case tls.VersionTLS12:
		return "TLS1.2"
	case tls.VersionTLS13:
		return "TLS1.3"
	}
	return fmt.Sprintf("tls-0x%x", v)
}
