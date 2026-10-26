// profiler_test.go — التشخيص العادل: عامل بلا دفاع، مدرع بالتحديات، ورفض خارج النصاب يبقى قانونيًا.
package nexus

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestProfilePlainTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "nginx")
		_, _ = w.Write([]byte("<html>plain page</html>"))
	}))
	defer srv.Close()

	rep, err := Profile(context.Background(), srv.Client(), srv.URL+"/")
	if err != nil {
		t.Fatalf("Profile: %v", err)
	}
	if rep.RiskLevel != 1 || rep.RecommendTier != "direct" {
		t.Fatalf("هدف عادي تهلّيد: risk=%d tier=%s", rep.RiskLevel, rep.RecommendTier)
	}
	if len(rep.Markers) != 0 {
		t.Fatalf("مؤشرات وهمية: %v", rep.Markers)
	}
	if rep.HonestyNote == "" {
		t.Fatal("ملاحظة الصدق مفقودة")
	}
}

func TestProfileHardenedTarget(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "cloudflare")
		w.Header().Set("Cf-Ray", "abc123")
		w.Header().Set("Cf-Mitigated", "challenge")
		w.Header().Set("X-Sucuri-Id", "1")
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`<script src="/cdn-cgi/challenge-platform/turnstile.js"></script>`))
	}))
	defer srv.Close()

	rep, err := Profile(context.Background(), srv.Client(), srv.URL+"/protected")
	if err != nil {
		t.Fatalf("Profile: %v", err)
	}
	if rep.RiskLevel < 4 || rep.RecommendTier != "residential" {
		t.Fatalf("هدف مدرع قُيّم برفقية: risk=%d tier=%s markers=%v", rep.RiskLevel, rep.RecommendTier, rep.Markers)
	}
	foundChallenge := false
	for _, m := range rep.Markers {
		if m == "challenge-script" {
			foundChallenge = true
		}
	}
	if !foundChallenge {
		t.Fatal("سكربت التحدي لم يرصد")
	}
}

func TestProfileRateLimited(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Cf-Ray", "anything") // حافة
		w.WriteHeader(429)
	}))
	defer srv.Close()
	rep, err := Profile(context.Background(), srv.Client(), srv.URL+"/")
	if err != nil {
		t.Fatalf("Profile: %v", err)
	}
	if rep.RiskLevel <= 1 {
		t.Fatalf("429 قُبل بلا أثر (risk=%d)", rep.RiskLevel)
	}
}

func TestProfileBadInput(t *testing.T) {
	if _, err := Profile(context.Background(), http.DefaultClient, ""); err == nil {
		t.Fatal("قبلت رابطًا فارغًا")
	}
	if _, err := Profile(context.Background(), http.DefaultClient, "not-a-url"); err == nil {
		t.Fatal("قبلت نصًا غير رابط")
	}
	// خادم الغائب — خطأ شفاف
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	u := srv.URL
	srv.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if _, err := Profile(ctx, http.DefaultClient, u+"/gone"); err == nil {
		t.Fatal("قبلت خادمًا ميتًا")
	}
}

func TestProfileRespectsHonestMarkerNames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Sucuri-Id", "cloudfree") // مؤشر sucuri
		_, _ = w.Write([]byte("denied by waf"))
	}))
	defer srv.Close()
	rep, err := Profile(context.Background(), srv.Client(), srv.URL+"/x")
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range rep.Markers {
		if strings.Contains(m, "sucuri") {
			found = true
		}
	}
	if !found {
		t.Fatalf("مؤشر sucuri مفقود: %v", rep.Markers)
	}
}
