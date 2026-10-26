// scheduler_test.go — قلب AIMD: هدوء يعصر النبض، 429 فجر يبطّئ، 3 إخفاقات تُ alarmed على التبديل.
package nexus

import (
	"context"
	"testing"
	"time"
)

func TestGovernorRelaxesOn200(t *testing.T) {
	g := NewGovernor("test-host", 500)
	// ارفع أولًا بأخطاء ثم خفف بـ 200
	g.Result(429)
	g.Result(429)
	mid := g.Snapshot()
	if mid.IntervalMS == mid.BaseMS {
		t.Fatal("429 لم يرفع الفاصل")
	}
	for i := 0; i < 30; i++ {
		g.Result(200)
		g.Result(200)
	}
	final := g.Snapshot()
	if final.IntervalMS != final.BaseMS {
		t.Fatalf("لم يعُد للأساس بعد النجاح: %dms != %dms", final.IntervalMS, final.BaseMS)
	}
}

func TestGovernorGrowsAndCapsOnBlocks(t *testing.T) {
	g := NewGovernor("host2", 300)
	for i := 0; i < 30; i++ {
		g.Result(429)
	}
	s := g.Snapshot()
	if s.IntervalMS != MaxInterval.Milliseconds() {
		t.Fatalf("السقف انكسر: %dms", s.IntervalMS)
	}
	if s.Fails != 30 {
		t.Fatalf("العد الحسابات: %d", s.Fails)
	}
}

func TestGovernorRotateSignalAndReset(t *testing.T) {
	g := NewGovernor("host3", 300)
	if g.ShouldRotate() {
		t.Fatal("بلا إخفاقات فُصح تحويل IP")
	}
	g.Result(403)
	g.Result(403)
	if g.ShouldRotate() {
		t.Fatal("تحويل جاهز قبل العتبة (3)")
	}
	g.Result(403)
	if !g.ShouldRotate() {
		t.Fatal("عند 3 إخفاقات يجب إطلاق إشارة التبديل")
	}
	g.Reset()
	if g.ShouldRotate() {
		t.Fatal("AfterReset ما زال يصرخ")
	}
	s := g.Snapshot()
	if s.IntervalMS != s.BaseMS {
		t.Fatal("Reset لم يرجع الفاصل للأساس")
	}
}

func TestGovernorWaitRespectsCancellation(t *testing.T) {
	g := NewGovernor("host4", 5000) // انتظار طويل
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := g.Wait(ctx); err == nil {
		t.Fatal("الإلغاء لم يشق الانتظار")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatal("الانتظار تجاوز مدة الإلغاء")
	}
}

func TestGovernorJitterBounds(t *testing.T) {
	g := NewGovernor("host5", 500)
	// لا اختبار دفعة الزمن هنا — نتحقق أن الفاصل في النطاق تُختار من حيث التصنيف
	s := g.Snapshot()
	if s.IntervalMS < MinInterval.Milliseconds() || s.IntervalMS < 500 {
		t.Fatalf("الفاصل تحت الحد الأدنى: %d", s.IntervalMS)
	}
}
