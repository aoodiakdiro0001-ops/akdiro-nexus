// metrics_test.go — القياس: العد الذري، دقة الحالات، وإخراج Prometheus سليم الشكل.
package nexus

import (
	"strings"
	"testing"
)

func TestMetricsCountingAndRate(t *testing.T) {
	m := NewMetrics()
	m.IncProbe()
	m.Result("a.local", 200, 12)
	m.Result("a.local", 200, 40)
	m.Result("a.local", 429, 80)
	m.Result("b.local", 403, 100)
	m.IncRotate("a.local")
	m.IncFinding()
	m.IncScan(3)

	s := m.Snapshot()
	if s["probes"].(int64) != 1 || s["oks"].(int64) != 2 || s["blocked"].(int64) != 2 {
		t.Fatalf("عدد فاسد: %v", s)
	}
	// نسبة النجاح العامة: 2/4
	if got := m.RateOks(); got < 0.49 || got > 0.51 {
		t.Fatalf("نسبة النجاح: %v", got)
	}
	ph := m.PerHost()
	if len(ph) != 2 {
		t.Fatalf("مضيفان مطلوبان: %d", len(ph))
	}
	// الأعلى منعًا أولًا
	if ph[0].Host != "a.local" || ph[0].Block != 1 {
		t.Fatalf("ترتيب / حساب المنع فسد: %+v", ph[0])
	}
}

func TestPrometheusShape(t *testing.T) {
	m := NewMetrics()
	m.IncProbe()
	m.Result("a.local", 200, 5)
	txt := m.Prometheus()
	for _, want := range []string{
		"# TYPE akdiro_probes_total counter",
		"akdiro_probes_total 1",
		"akdiro_ok_total 1",
		"akdiro_success_rate 1",
		`akdiro_host_blocked_total{host="a.local"} 0`,
	} {
		if !strings.Contains(txt, want) {
			t.Fatalf("Prometheus بلا [%s]:\n%s", want, txt)
		}
	}
}

func TestLatencyBuckets(t *testing.T) {
	m := NewMetrics()
	for _, ms := range []int64{5, 30, 700, 99999} {
		m.observeLatency(ms)
	}
	// خانات: ≤10 → 1 ، ≤50 → 1، ≤1000 → 1، +Inf → 1
	if m.lbLatency[0].Load() != 1 || m.lbLatency[2].Load() != 1 || m.lbLatency[6].Load() != 1 {
		t.Fatal("خانات أقلية فاسدة")
	}
	if m.lbLatency[LatencySlots-1].Load() != 1 {
		t.Fatal("خانة +Inf فاسدة")
	}
}
