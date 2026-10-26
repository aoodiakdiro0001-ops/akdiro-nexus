// session.go — SDK داخل العملية: جلسة كاملة بـ 3 أسطر (بصمة + توجيه + جدولة + تفويض + قياس).
// الحقيقة المعلنة: المكونات الثلاثة جاهزة ومُختبرة — هذه الطبقة تربطها في مهمة واحدة.
package nexus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// SessionConfig تكوين الجلسة (من هم ما تملكه anyIP يمر هنا).
type SessionConfig struct {
	ScopePath             string // ملف النصاب الموقَّع
	DatacenterURL         string // قالب مدخل DC (يحتاج {session})
	ResidentialURL        string // قالب مدخل الريزيدنتيال (يحتاج {session})
	MaxParallelClients    int    // سقف تخزين العملاء المُعاد استعمالها (B2B)
	DefaultGovernorBaseMS int
	Metrics               *Metrics // مشاركة القياس الحي مع daemon (nil = قياس منفصل للجلسة)
}

// Session حمولة النظام الكاملة قابلة للاستدعاء البرمجي.
type Session struct {
	cfg     SessionConfig
	scope   *Scope
	router  ProxyConfig
	gmu     sync.Mutex
	govs    map[string]*governorEntry
	clients sync.Mutex
	clientM map[string]*http.Client // إعادة استعمال الاتصالات — سرعة الانتاج
	met     *Metrics
}

type governorEntry struct {
	g    *Governor
	sess int64
}

// NewSession يفتح جلسة: تحميل النصاب (إن وجد) وحاشية العملاء.
func NewSession(cfg SessionConfig) (*Session, error) {
	if cfg.MaxParallelClients <= 0 {
		cfg.MaxParallelClients = 64
	}
	s := &Session{
		cfg:     cfg,
		router:  ProxyConfig{DatacenterURL: cfg.DatacenterURL, ResidentialURL: cfg.ResidentialURL},
		govs:    make(map[string]*governorEntry),
		clientM: make(map[string]*http.Client),
		met:     NewMetrics(),
	}
	if cfg.Metrics != nil {
		s.met = cfg.Metrics
	}
	if cfg.ScopePath != "" {
		if sc, err := LoadScope(cfg.ScopePath); err == nil && sc.Validate() == nil {
			s.scope = sc
		}
	}
	return s, nil
}

// Metrics قياسات الجلسة الحية.
func (s *Session) Metrics() *Metrics { return s.met }

// Gate بوابة التفويض: نصاب الجلسة (يُنشأ بتعهد عبر daemon أو ملف).
func (s *Session) Gate(urlStr string) error {
	host, err := sessionHost(urlStr)
	if err != nil {
		return err
	}
	return s.scope.Gate(host)
}

func sessionHost(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", errors.New("هدف فارغ")
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil || u.Hostname() == "" {
		return "", errors.New("رابط غير صالح")
	}
	return strings.ToLower(u.Hostname()), nil
}

// Fetch طلاء كامل: تفويض → جدولة → طبقة → بصمة → قياس → Hot-Swap.
func (s *Session) Fetch(ctx context.Context, urlStr, method string) (*FetchReport, error) {
	if method == "" {
		method = http.MethodGet
	}
	host, err := sessionHost(urlStr)
	if err != nil {
		return nil, err
	}
	if err := s.scope.Gate(host); err != nil {
		return nil, err
	}

	s.met.IncProbe()
	e := s.governorFor(host, time.Now().Unix())

	if err := e.g.Wait(ctx); err != nil {
		return nil, err
	}

	tier := s.pickTier(e)
	proxyURL, err := s.router.ForTier(tier, s.sessionToken(host, e))
	if err != nil {
		return nil, err
	}
	client, err := s.clientFor(proxyURL)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, urlStr, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "AkdiroNexus/1.0 (+session; owner=Baran)")

	t0 := time.Now()
	resp, err := client.Do(req)
	lat := time.Since(t0).Milliseconds()
	if err != nil {
		e.g.Result(500)
		s.met.Result(host, 500, lat)
		return nil, fmt.Errorf("الجلب فشل: %w", err)
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	e.g.Result(code)
	s.met.Result(host, code, lat)

	rotated := false
	if e.g.ShouldRotate() {
		rotated = true
		e.g.Reset()
		e.sess++
		s.met.IncRotate(host)
	}

	return &FetchReport{
		URL: urlStr, Host: host, Status: code, LatencyMS: lat,
		Tier: string(tier), Rotated: rotated,
		ContentType: resp.Header.Get("Content-Type"),
		Server:      resp.Header.Get("Server"),
	}, nil
}

// FetchReport ملخص الجلب — خفيف وقابل للتسلسل.
type FetchReport struct {
	URL         string `json:"url"`
	Host        string `json:"host"`
	Status      int    `json:"status"`
	LatencyMS   int64  `json:"latency_ms"`
	Tier        string `json:"tier"`
	Rotated     bool   `json:"rotated"`
	ContentType string `json:"content_type"`
	Server      string `json:"server"`
}

func (s *Session) governorFor(host string, now int64) *governorEntry {
	s.gmu.Lock()
	defer s.gmu.Unlock()
	e, ok := s.govs[host]
	if !ok {
		e = &governorEntry{g: NewGovernor(host, s.cfg.DefaultGovernorBaseMS), sess: now}
		s.govs[host] = e
	}
	return e
}

// pickTier: بعد أي منع ⇒ نرفع إلى بروكسي (اقتصادية أولاً — residential عند البنية الثقيلة).
func (s *Session) pickTier(e *governorEntry) Tier {
	if e.g.ShouldRotate() {
		return TierResidential
	}
	if e.g.BlocksCount() >= 1 {
		return TierDatacenter
	}
	return TierDirect
}

// clientFor يبني/يعيد عميل — إعادة الاستعمال = سرعة B2B (لا جديد TLS للكل طلب).
func (s *Session) clientFor(proxyURL string) (*http.Client, error) {
	s.clients.Lock()
	defer s.clients.Unlock()
	if c, ok := s.clientM[proxyURL]; ok {
		return c, nil
	}
	c, err := (Dialer{ProxyURL: proxyURL}).Client()
	if err != nil {
		return nil, err
	}
	if len(s.clientM) < s.cfg.MaxParallelClients {
		s.clientM[proxyURL] = c
	}
	return c, nil
}

func (s *Session) sessionToken(host string, e *governorEntry) string {
	return fmt.Sprintf("sess-%s-%d-%d",
		strings.ReplaceAll(strings.ToLower(host), ".", "-"),
		e.sess, time.Now().UnixNano()%1e6)
}
