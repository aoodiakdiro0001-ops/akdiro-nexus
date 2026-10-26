// consent_test.go — بوابة التفويض: التعهد إلزامي، التلاعب يُرفض، النطاقات الفرعية تدخل.
package nexus

import (
	"strings"
	"testing"
)

func TestScopeRequiresAttestation(t *testing.T) {
	if _, err := NewScope([]string{"example.com"}, "موافق"); err != ErrNoAttestation {
		t.Fatalf("قبلت تعهدًا غير حرفي: %v", err)
	}
	s, err := NewScope([]string{"example.com"}, AttestationPhrase)
	if err != nil {
		t.Fatalf("رفض التعهد الحرفي: %v", err)
	}
	if s.Validate() != nil {
		t.Fatal("نصاب صحيح فشل في التحقق")
	}
}

func TestScopeAllowsSubdomains(t *testing.T) {
	s, _ := NewScope([]string{"example.com"}, AttestationPhrase)
	cases := map[string]bool{
		"example.com":     true,
		"www.example.com": true,
		"a.b.example.com": true,
		"badexample.com":  false, // لاحقة نقطية كاملة
		"other.io":        false,
	}
	for h, want := range cases {
		if got := s.Allows(h); got != want {
			t.Errorf("Allows(%q)=%v want %v", h, got, want)
		}
	}
	if err := s.Gate("other.io"); err != ErrOutOfScope {
		t.Fatalf("بوابة لم ترفض خارج النصاب: %v", err)
	}
}

func TestScopeTamperRejected(t *testing.T) {
	s, _ := NewScope([]string{"example.com"}, AttestationPhrase)
	// تلاعب: إضافة نطاق بعد التوقيع
	tampered := *s
	tampered.Domains = append(tampered.Domains, "evil-target.net")
	if err := tampered.Validate(); err == nil || !strings.Contains(err.Error(), "تلاعب") {
		t.Fatalf("التلاعب نجح: %v", err)
	}
	// تلاعب بالعبارة
	t2 := *s
	t2.Phrase = " його تعهد مزروع"
	if t2.Validate() == nil {
		t.Fatal("عبارة مزوَّعة قُبلت")
	}
}

func TestScopeRoundtripPersistence(t *testing.T) {
	s, _ := NewScope([]string{"example.com", "own.co"}, AttestationPhrase)
	dir := t.TempDir()
	path := dir + "/consent.scope.json"
	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}
	l, err := LoadScope(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if l.Validate() != nil || !l.Allows("own.co") {
		t.Fatalf("النصاب المنقولة فسدت: %v", l.Validate())
	}
}
