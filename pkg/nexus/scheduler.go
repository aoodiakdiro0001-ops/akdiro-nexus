// scheduler.go — الجدولة التكيفية: AIMD + Jitter، ونبض لكل مضيف مستقل.
// ينزل تقديره مع الهدوء (200)، يرفع فورًا عند 429/403، ويطلق إشارة التيم (Hot-Swap).
package nexus

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// Governor منظم معدل لكل مضيف.
type Governor struct {
	mu       sync.Mutex
	host     string
	base     time.Duration
	interval time.Duration
	fails    int
	// المرصد الكلي للنظام
	OKs, Blocks int
}

const (
	MinInterval      = 300 * time.Millisecond
	MaxInterval      = 30 * time.Second
	RotateThreshold  = 3
	IntervalShrink   = 0.8 // انخفاض عادل بعد 200
	IntervalGrow     = 2.0 // رفع صارم بعد 429/403
	IntervalSoftGrow = 1.1 // أخطاء أخرى
)

func NewGovernor(host string, baseMS int) *Governor {
	if baseMS <= 0 {
		baseMS = 500
	}
	base := time.Duration(baseMS) * time.Millisecond
	if base < MinInterval {
		base = MinInterval
	}
	return &Governor{host: host, base: base, interval: base}
}

// Wait يلتزم بجدولة نصيب التوقيت مع ±25% Jitter (يجارِ تقلب الزمن الحقيقي).
func (g *Governor) Wait(ctx context.Context) error {
	g.mu.Lock()
	iv := g.interval
	jf := 0.75 + rand.Float64()*0.5 // 0.75..1.25
	g.mu.Unlock()

	it := time.Duration(float64(iv) * jf)
	if it < MinInterval {
		it = MinInterval
	}
	select {
	case <-time.After(it):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// Result يعدّل السرعة حسب كود الرد — قلب AIMD.
func (g *Governor) Result(code int) {
	g.mu.Lock()
	defer g.mu.Unlock()
	switch {
	case code >= 200 && code < 300:
		g.OKs++
		g.fails = 0
		g.interval = time.Duration(float64(g.interval) * IntervalShrink)
		if g.interval < g.base {
			g.interval = g.base
		}
	case code == 429 || code == 403 || code == 503:
		g.Blocks++
		g.fails++
		g.interval = time.Duration(float64(g.interval) * IntervalGrow)
		if g.interval > MaxInterval {
			g.interval = MaxInterval
		}
	case code >= 400:
		g.fails++
		g.interval = time.Duration(float64(g.interval) * IntervalSoftGrow)
		if g.interval > MaxInterval {
			g.interval = MaxInterval
		}
	}
}

// ShouldRotate إشارة الـ Hot-Swap — بدّل IP/بصمة الآن.
func (g *Governor) ShouldRotate() bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.fails >= RotateThreshold
}

// Reset يصف المحاولات بعد السايح/التبديل.
func (g *Governor) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.fails = 0
	g.interval = g.base
}

// BlocksCount عدد المنع المرصود (يدخل في الترقية التلقائية للطبقة).
func (g *Governor) BlocksCount() int {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.Blocks
}

type GovState struct {
	Host       string        `json:"host"`
	IntervalMS int64         `json:"interval_ms"`
	BaseMS     int64         `json:"base_ms"`
	Fails      int           `json:"fails"`
	OKs        int           `json:"oks"`
	Blocks     int           `json:"blocks"`
	NextWait   time.Duration `json:"-"`
}

func (g *Governor) Snapshot() GovState {
	g.mu.Lock()
	defer g.mu.Unlock()
	return GovState{
		Host:       g.host,
		IntervalMS: g.interval.Milliseconds(),
		BaseMS:     g.base.Milliseconds(),
		Fails:      g.fails,
		OKs:        g.OKs,
		Blocks:     g.Blocks,
	}
}

// AtSTring طباعة صادقة للعرض.
func (s GovState) String() string {
	return fmt.Sprintf("%s ± %dms/2xx (أساس %dms, إخفاقات %d)", s.Host, s.IntervalMS, s.BaseMS, s.Fails)
}
