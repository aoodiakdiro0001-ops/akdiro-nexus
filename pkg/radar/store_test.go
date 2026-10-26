// radar_test.go — دفتر الثغرات: تشوف مكرر، ترتيب الشدّة، تُستعاد بعد الإغلاق.
package radar

import (
	"path/filepath"
	"testing"
)

func TestLedgerAddDedupe(t *testing.T) {
	dir := t.TempDir()
	l, err := OpenLedger(filepath.Join(dir, "f.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	f1, created, err := l.Add("demo.local", "xss", 4, "المعامل موسوم", "vulnradar")
	if err != nil || !created {
		t.Fatalf("Add أولى: created=%v err=%v", created, err)
	}
	f2, created, err := l.Add("demo.local", "xss", 4, "المعامل موسوم", "vulnradar")
	if err != nil {
		t.Fatal(err)
	}
	if created || f2.Hash != f1.Hash {
		t.Fatal("الكشف المكرر تكرر في الدفتر")
	}
	if got := len(l.List("", 0)); got != 1 {
		t.Fatalf("عدد ≤ الكشف: %d", got)
	}
}

func TestLedgerSeverityFilterAndSort(t *testing.T) {
	dir := t.TempDir()
	l, _ := OpenLedger(filepath.Join(dir, "f.jsonl"))
	_, _, _ = l.Add("demo.local", "info-leak", 2, "وسوم مستقدمة", "audit")
	_, _, _ = l.Add("demo.local", "rce", 5, "تنفيذ أومار", "manual")
	_, _, _ = l.Add("demo.local", "sqli", 4, "تصيب SQL", "vulnradar")

	all := l.List("", 0)
	if len(all) != 3 {
		t.Fatalf("عدد = 3, got %d", len(all))
	}
	if all[0].Severity != 5 || all[1].Severity != 4 {
		t.Fatalf("ترتيب الشدّة فشل: %d,%d", all[0].Severity, all[1].Severity)
	}
	hi := l.List("", 4)
	if len(hi) != 2 {
		t.Fatalf("مرشح الشدّة: %d", len(hi))
	}
	st := l.Stats()
	if st["total"] != 3 || st["critical"] != 1 || st["high"] != 1 {
		t.Fatalf("إحصائيات قابلة للتشهية: %v", st)
	}
}

func TestLedgerPersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.jsonl")
	l1, _ := OpenLedger(path)
	_, _, err := l1.Add("app.local", "ssrf", 3, "فالنمور داخلية", "vulnradar")
	if err != nil {
		t.Fatal(err)
	}
	l2, err := OpenLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(l2.List("", 0)); got != 1 {
		t.Fatalf("البيانات لم تستعداد بعد openclose: %d", got)
	}
	// الخدمة: ملف التالف لا يُسقط الدفتر
	if _, err := OpenLedger(filepath.Join(dir, "missing.jsonl")); err != nil {
		t.Fatal("ملف الغائب يسبب خطأ")
	}
}

func TestLedgerBoundaries(t *testing.T) {
	dir := t.TempDir()
	l, _ := OpenLedger(filepath.Join(dir, "f.jsonl"))
	if _, _, err := l.Add("", "xss", 1, "-", "s"); err == nil {
		t.Fatal("قبلت هدفًا فارغًا")
	}
	if _, _, err := l.Add("app.local", "", 3, "-", "s"); err == nil {
		t.Fatal("قبلت نوعًا فارغًا")
	}
	// شدّة خارج الحدود تُلحد صحيحًا
	f, _, err := l.Add("app.local", "dos", 99, "مبالغ", "test")
	if err != nil {
		t.Fatal(err)
	}
	if f.Severity != 5 {
		t.Fatalf("الشدّة لم تُلحد: %d", f.Severity)
	}
}
