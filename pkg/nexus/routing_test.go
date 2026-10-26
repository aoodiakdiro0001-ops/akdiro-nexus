// routing_test.go — قوالب anyIP: {session} إلزامية، التوسيع صحيح، الترتيب اقتصادي.
package nexus

import (
	"strings"
	"testing"
)

func TestProxyConfigRequiresSessionPlaceholder(t *testing.T) {
	pc := ProxyConfig{DatacenterURL: "http://user:pass@gate.io:8000"}
	if err := pc.validate(pc.DatacenterURL); err == nil {
		t.Fatal("قالب بلا {session} قُبل")
	}
	pc2 := ProxyConfig{DatacenterURL: "http://user-session-{session}:pass@gate.io:8000"}
	if err := pc2.validate(pc2.DatacenterURL); err != nil {
		t.Fatalf("قالب سليم رفض: %v", err)
	}
	var empty ProxyConfig
	if err := empty.validate(""); err != nil {
		t.Fatal("عرض فارغ مرفوض بلا سبب")
	}
}

func TestProxyConfigExpandHasSession(t *testing.T) {
	pc := ProxyConfig{ResidentialURL: "http://u-{session}:p@gate.ips.io:9000"}
	u, err := pc.ForTier(TierResidential, "sess-123")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(u, "sess-123") || !strings.Contains(u, "gate.ips.io") {
		t.Fatalf("توسيع فسد: %s", u)
	}
	// الطبقة الاقتصادية: direct لا يحمل بروكسي
	u2, _ := pc.ForTier(TierDirect, "x")
	if u2 != "" {
		t.Fatalf("direct خرج ببروكسي: %s", u2)
	}
}

func TestTierCostEscalation(t *testing.T) {
	// قرار الطبقات: 1-2 بلا حضور residential (توثيق اقتصادي)
	pc := ProxyConfig{ResidentialURL: "http://x-{session}@r.io"}
	rep := DefenseReport{RiskLevel: 1, RecommendTier: "direct"}
	if rep.RiskLevel <= 2 && rep.RecommendTier == "residential" {
		t.Fatal("هدف محايد انتقل مباشرة للطبقة الغالية")
	}
	if u, _ := pc.ForTier(TierDatacenter, "s"); u != "" {
		t.Fatal("طبقة داتا سنتر طلعت بروكسي رغم قالب فارغ")
	}
}
