// session_test.go — الجلسة المربوطة: تفويض، جدولة فعلية، ترقية الطبقة بعد المنع، قياس.
package nexus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func newTestScope(t *testing.T, hosts ...string) string {
	t.Helper()
	s, err := NewScope(hosts, AttestationPhrase)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "consent.scope.json")
	if err := s.Save(path); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestSessionFetchHappyPath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Server", "testd")
		_, _ = w.Write([]byte("ok"))
	}))
	s, err := NewSession(SessionConfig{ScopePath: newTestScope(t, "127.0.0.1"), DefaultGovernorBaseMS: 200})
	if err != nil {
		t.Fatal(err)
	}
	rep, err := s.Fetch(context.Background(), srv.URL+"/", "")
	if err != nil {
		t.Fatalf("Fetch: %v", err)
	}
	if rep.Status != 200 || rep.Tier != "direct" || rep.Rotated {
		t.Fatalf("الجلب فسد: %+v", rep)
	}
	if s.Metrics().RateOks() != 1 {
		t.Fatalf("نسبة الجلسة: %v", s.Metrics().RateOks())
	}
}

func TestSessionConsentInjectionEnforced(t *testing.T) {
	s, _ := NewSession(SessionConfig{ScopePath: newTestScope(t, "own.co")})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	if _, err := s.Fetch(context.Background(), srv.URL+"/x", ""); err != ErrOutOfScope {
		t.Fatalf("بلا تعهد — يجب الرفض القانوني: %v", err)
	}
}

func TestSessionEscalatesTierAfterBlocks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(429)
	}))
	s, _ := NewSession(SessionConfig{
		ScopePath:             newTestScope(t, "127.0.0.1"),
		DatacenterURL:         "http://u-{session}:p@gate.dc.test:8000",
		ResidentialURL:        "http://u-{session}:p@gate.res.test:9000",
		DefaultGovernorBaseMS: 200,
	})
	// ثلاث جلبات محملة — سيصعد للطبقةDC مع أول منع ثم residentialθ عند Hot-Swap
	tiers := map[string]bool{}
	for i := 0; i < 3; i++ {
		rep, err := s.Fetch(context.Background(), srv.URL+"/b", "")
		if err != nil {
			t.Fatalf("Fetch %d: %v", i, err)
		}
		tiers[rep.Tier] = true
	}
	if !tiers["datacenter"] {
		t.Fatalf("لم ترق للـDC بعد أول منع والبروكسي مضبوط: %v", tiers)
	}
	if s.Metrics().Snapshot()["blocked"].(int64) != 3 {
		t.Fatal("منع غير مؤould")
	}
}

func TestSessionGovernorRateLocalt(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	s, _ := NewSession(SessionConfig{ScopePath: newTestScope(t, "127.0.0.1"), DefaultGovernorBaseMS: 150})
	// طلاءان متتاليان يجب أن يتبعهما فاصل ≥ الأساس (جدولة حقيقية لا صفارة)
	start := time.Now()
	for i := 0; i < 2; i++ {
		if _, err := s.Fetch(context.Background(), srv.URL+"/", ""); err != nil {
			t.Fatal(err)
		}
	}
	elapsed := time.Since(start)
	if elapsed < 100*time.Millisecond {
		t.Fatalf("الجدولة تضافرت (مرور فوري): %v", elapsed)
	}
}

func TestSessionScopeFileMissing(t *testing.T) {
	dir := t.TempDir()
	s, err := NewSession(SessionConfig{ScopePath: dir + "/missing.json"})
	if err != nil {
		t.Fatalf("جلسة بلا ملف نصاب تُنشأ — لكن لا تعمل: %v", err)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	if _, err := s.Fetch(context.Background(), srv.URL+"/", ""); err == nil {
		t.Fatal("قبلت بدون نصاب")
	}
	if _, err := os.Stat(dir + "/missing.json"); !os.IsNotExist(err) {
		t.Fatal("لم يُنشئ ملفًا — التصميم: لا تعهد لا ملف")
	}
}
