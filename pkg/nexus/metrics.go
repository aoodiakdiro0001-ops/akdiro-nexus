// metrics.go — قياس حي كامل: بلا خوادم خارجية، يعمل محليًا ويعرض Prometheus text + JSON.
// نسبة النجاح هنا ليست وعودًا للتسويق — إنها أداة قياس فعلية لكل هدف.
package nexus

import (
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Metrics مسجّل ذري آمن للخيوط.
type Metrics struct {
	startedAt time.Time

	probes    atomic.Int64
	oks       atomic.Int64
	blocks    atomic.Int64
	errors    atomic.Int64
	rotations atomic.Int64
	findings  atomic.Int64
	scans     atomic.Int64

	// بخزين زمن الاستجابة (ملّي ثانية)
	lbLatency [LatencySlots]atomic.Int64

	perHostMu sync.Mutex
	perHost   map[string]*hostCounters
}

type hostCounters struct {
	ok, block, rotate, errs atomic.Int64
}

// الحدود (مللي ثانية، 9 حدود) — عدد الخلايا الذرية 10 (9 + ما فوق الخانة الأخيرة)
const LatencyMSBucketCount = 9
const LatencySlots = LatencyMSBucketCount + 1

// LatencyBucketsMS الحدود الفعلية (شيك وقت التشغيل أدناه)
var LatencyBucketsMS = []int64{10, 25, 50, 100, 250, 500, 1000, 2500, 5000}

func init() {
	if len(LatencyBucketsMS) != LatencyMSBucketCount {
		panic("LatencyBucketsMS غير متطابق مع LatencyMSBucketCount")
	}
}

func NewMetrics() *Metrics {
	return &Metrics{startedAt: time.Now().UTC(), perHost: make(map[string]*hostCounters)}
}

func (m *Metrics) Uptime() time.Duration { return time.Since(m.startedAt) }

// IncProbe يسمي استدعاء أي طلب استكشاف/جلب.
func (m *Metrics) IncProbe() { m.probes.Add(1) }

// Result يسجل كود الرد + زمنه، ويصنف على مستوى المضيف أيضًا.
func (m *Metrics) Result(host string, code int, latencyMS int64) {
	if code >= 200 && code < 300 {
		m.oks.Add(1)
		m.hostOf(host).ok.Add(1)
	} else if code == 429 || code == 403 || code == 503 {
		m.blocks.Add(1)
		m.hostOf(host).block.Add(1)
	} else {
		m.errors.Add(1)
		m.hostOf(host).errs.Add(1)
	}
	m.observeLatency(latencyMS)
}

// IncRotate إشارة تبديل IP/جلسة (Hot-Swap).
func (m *Metrics) IncRotate(host string) {
	m.rotations.Add(1)
	m.hostOf(host).rotate.Add(1)
}

func (m *Metrics) IncFinding()     { m.findings.Add(1) }
func (m *Metrics) IncScan(n int64) { m.scans.Add(n) }

func (m *Metrics) hostOf(host string) *hostCounters {
	m.perHostMu.Lock()
	defer m.perHostMu.Unlock()
	if h, ok := m.perHost[host]; ok {
		return h
	}
	h := &hostCounters{}
	m.perHost[host] = h
	return h
}

func (m *Metrics) observeLatency(ms int64) {
	idx := LatencySlots - 1 // +Inf افتراضي
	for i, th := range LatencyBucketsMS {
		if ms <= th {
			idx = i
			break
		}
	}
	m.lbLatency[idx].Add(1)
}

// HostStat حالات عمّال التوجيه.
type HostStat struct {
	Host    string  `json:"host"`
	OK      int64   `json:"ok"`
	Block   int64   `json:"block"`
	Errs    int64   `json:"errors"`
	Rotate  int64   `json:"rotate"`
	Success float64 `json:"success_rate"` // المقياس الفعلي — لا وعود
}

func (m *Metrics) PerHost() []HostStat {
	m.perHostMu.Lock()
	defer m.perHostMu.Unlock()
	out := make([]HostStat, 0, len(m.perHost))
	for h, c := range m.perHost {
		total := c.ok.Load() + c.block.Load() + c.errs.Load()
		sr := 0.0
		if total > 0 {
			sr = float64(c.ok.Load()) / float64(total)
		}
		out = append(out, HostStat{Host: h, OK: c.ok.Load(), Block: c.block.Load(), Errs: c.errs.Load(), Rotate: c.rotate.Load(), Success: sr})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Block > out[j].Block })
	return out
}

// Snapshot مجلدة JSON للواجهة — النسبة المستعملة RateOks (مقاسة على النتائج كلها).
func (m *Metrics) Snapshot() map[string]any {
	return map[string]any{
		"uptime_s":     int64(m.Uptime().Seconds()),
		"probes":       m.probes.Load(),
		"oks":          m.oks.Load(),
		"blocked":      m.blocks.Load(),
		"errors":       m.errors.Load(),
		"rotations":    m.rotations.Load(),
		"findings":     m.findings.Load(),
		"scans":        m.scans.Load(),
		"success_rate": m.RateOks(),
		"per_host":     m.PerHost(),
	}
}

// Prometheus ينتج الصيغة القياسية — تُقرأ مباشرة بأي سكرابر/Grafana محليًا.
func (m *Metrics) Prometheus() string {
	var b strings.Builder
	w := func(line string) { b.WriteString(line + "\n") }

	w("# HELP akdiro_probes_total عدد الطلبات المنفذة")
	w("# TYPE akdiro_probes_total counter")
	w("akdiro_probes_total " + itoa(m.probes.Load()))
	w("# HELP akdiro_ok_total الطلبات 2xx")
	w("# TYPE akdiro_ok_total counter")
	w("akdiro_ok_total " + itoa(m.oks.Load()))
	w("# HELP akdiro_blocked_total العملة 429/403/503")
	w("# TYPE akdiro_blocked_total counter")
	w("akdiro_blocked_total " + itoa(m.blocks.Load()))
	w("# HELP akdiro_rotations_total تبديلات IP/SESSION")
	w("# TYPE akdiro_rotations_total counter")
	w("akdiro_rotations_total " + itoa(m.rotations.Load()))
	w("# HELP akdiro_errors_total أخطاء اتصال/ردود غير الـ2xx وغير المنع")
	w("# TYPE akdiro_errors_total counter")
	w("akdiro_errors_total " + itoa(m.errors.Load()))
	w("# HELP akdiro_findings_total الثغرات المسجلة")
	w("# TYPE akdiro_findings_total counter")
	w("akdiro_findings_total " + itoa(m.findings.Load()))
	w("# HELP akdiro_scan_runs_total عدد مسارات محرك المهام")
	w("# TYPE akdiro_scan_runs_total counter")
	w("akdiro_scan_runs_total " + itoa(m.scans.Load()))
	w("# HELP akdiro_success_rate نسبة النجاح المقاسة — لا وعد تسويقي")
	w("# TYPE akdiro_success_rate gauge")
	w("akdiro_success_rate " + ftoa(m.RateOks()))

	// الخانات الزمنية
	for i, th := range LatencyBucketsMS {
		w("# HELP akdiro_latency_bucket_" + itoa(th) + " الطلبات ≤ " + itoa(th) + "ms")
		w("# TYPE akdiro_latency_bucket_" + itoa(th) + " counter")
		w("akdiro_latency_bucket_" + itoa(th) + " " + itoa(m.lbLatency[i].Load()))
	}
	w("# HELP akdiro_latency_bucket_inf")
	w("# TYPE akdiro_latency_bucket_inf counter")
	w("akdiro_latency_bucket_inf " + itoa(m.lbLatency[LatencySlots-1].Load()))

	// لكل مضيف (المقاييس الفعلية)
	for _, h := range m.PerHost() {
		w("akdiro_host_ok_total{host=\"" + h.Host + "\"} " + itoa(h.OK))
		w("akdiro_host_blocked_total{host=\"" + h.Host + "\"} " + itoa(h.Block))
		w("akdiro_host_errors_total{host=\"" + h.Host + "\"} " + itoa(h.Errs))
		w("akdiro_host_rotations_total{host=\"" + h.Host + "\"} " + itoa(h.Rotate))
	}
	return b.String()
}

// RateOks نسبة النجاح العامة المحسوبة: نجاح ÷ (نجاح + منع + أخطاء) — كل النتائج تدخل.
func (m *Metrics) RateOks() float64 {
	total := m.oks.Load() + m.blocks.Load() + m.errors.Load()
	if total == 0 {
		return 0
	}
	return float64(m.oks.Load()) / float64(total)
}

func itoa(v int64) string {
	return string(strconv.AppendInt(nil, v, 10))
}

func ftoa(f float64) string {
	return strconv.FormatFloat(f, 'g', 6, 64)
}
