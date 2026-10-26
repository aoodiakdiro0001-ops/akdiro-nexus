// routing.go — توجيه anyIP الذكي: طبقة من الموارد Academic — من قالب دفاعرك وفق درجة المخاطرة.
// يعمل مع أي مزود بروكسي يقدم Gateways بروكي HTTP(S) — بينهم anyIP — عبر نماذج سجسون.
package nexus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Tier مستويات البروكسي.
type Tier string

const (
	TierDirect      Tier = "direct"
	TierDatacenter  Tier = "datacenter"
	TierResidential Tier = "residential"
)

// ProxyConfig نماذج anyIP الحقيقية — المستخدم يحدد نقاط الدخول من حسابه (بدون ما نخترع API).
type ProxyConfig struct {
	// DatacenterURL/ResidentialURL قالب بوابة بروكسي، مثل:
	// http://USER-session-{session}:PASS@gate.provider.io:8000
	DatacenterURL  string `json:"datacenter_url"`
	ResidentialURL string `json:"residential_url"`
	RotateAfter    int    `json:"rotate_after"`
}

// validate {session} شرط إلزامي (لجلسات دوّارة) — والفحص عبر استبدالها بعينة آمنة.
func (pc ProxyConfig) validate(u string) error {
	if u == "" {
		return nil
	}
	if !strings.Contains(u, "{session}") {
		return errors.New("قالب البروكسي يجب أن يحتوي {session} لإدارة جلسات دوّارة")
	}
	probe := strings.ReplaceAll(u, "{session}", "probesample")
	parsed, err := url.Parse(probe)
	if err != nil {
		return fmt.Errorf("قالب غير صحيح: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return errors.New("قالب البروكسي يجب أن يكون http/https")
	}
	if parsed.Hostname() == "" {
		return errors.New("قالب البروكسي بلا مضيف")
	}
	return nil
}

// Expand يحول القالب إلى URL حقيقي بسلسلة محددة (sticky session).
func (pc ProxyConfig) Expand(u, session string) (string, error) {
	if u == "" {
		return "", nil
	}
	return strings.ReplaceAll(u, "{session}", url.PathEscape(session)), pc.validate(u)
}

// ForTier يختار البروكسي المناسب من الدرجة (wrf التكلفة صادقة: costly only عند صعوبة الهدف).
func (pc ProxyConfig) ForTier(t Tier, session string) (string, error) {
	switch t {
	case TierResidential:
		return pc.Expand(pc.ResidentialURL, session)
	case TierDatacenter:
		return pc.Expand(pc.DatacenterURL, session)
	default:
		return "", nil
	}
}

// ExitInfo نتيجة اختبار حيار (من أي IP خرجنا؟).
type ExitInfo struct {
	IP        string `json:"ip"`
	ViaProxy  bool   `json:"via_proxy"`
	Tier      string `json:"tier"`
	LatencyMS int64  `json:"latency_ms"`
	Err       string `json:"error,omitempty"`
}

// TestEgress فحص IP الخارجي عبر بروكسي — إلى api.ipify.org (هدف مشاع عام للفحص).
func TestEgress(ctx context.Context, client *http.Client, tier Tier) ExitInfo {
	if client == nil {
		client = http.DefaultClient
	}
	t0 := time.Now()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.ipify.org?format=json", nil)
	if err != nil {
		return ExitInfo{Err: err.Error(), Tier: string(tier)}
	}
	resp, err := client.Do(req)
	if err != nil {
		return ExitInfo{Err: err.Error(), Tier: string(tier), ViaProxy: tier != TierDirect}
	}
	defer resp.Body.Close()
	var out struct {
		IP string `json:"ip"`
	}
	_ = json.NewDecoder(io.LimitReader(resp.Body, 256)).Decode(&out)
	return ExitInfo{
		IP: out.IP, Tier: string(tier), LatencyMS: time.Since(t0).Milliseconds(),
		ViaProxy: tier != TierDirect,
	}
}
