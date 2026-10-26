// consent.go — بوابة التفويض: النظام لا يعمل على أي نطاق إلا بتعهّد محلي مُوقَّع.
// هذا ليس زينة — القانون هو الحد. Baran © 2026
package nexus

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

// AttestationPhrase عبارة التعهد الإلزامية (يجب تقديمها حرفيًا).
const AttestationPhrase = "أتعهد أنني مخوّل بالتحقق من هذا النطاق وأنه لي أو لدي إذن مكتوب"

var (
	ErrNoAttestation = errors.New(" التفويض مرفوض: التعهد الإلزامي غير مقدَّم حرفيًا (الأداة تعمل على نطاقاتك/النطاقات المخوّلة فقط)")
	ErrOutOfScope    = errors.New("النطاق خارج نصاب التفويض — أضفه بتعهد صريح عبر /api/scope (قانونًا: لا نعمل على ما نملكه أو نُذن)")
)

// Scope نصاب تفويض موقَّع محليًا (sha256) — بلا خوادم.
type Scope struct {
	Domains  []string  `json:"domains"`
	Phrase   string    `json:"phrase"`
	SignedAt time.Time `json:"signed_at"`
	Sign     string    `json:"sign"`
}

func signScope(domains []string, phrase string) string {
	sorted := append([]string(nil), domains...)
	sort.Strings(sorted)
	sum := sha256.Sum256([]byte("nexus-consent-v1|" + strings.Join(sorted, ",") + string(phrase)))
	return hex.EncodeToString(sum[:16])
}

// NewScope يبني نصابًا — يقبل التعهد الكتابي الحرفي فقط.
func NewScope(domains []string, phrase string) (*Scope, error) {
	clean := make([]string, 0, len(domains))
	for _, d := range domains {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" {
			clean = append(clean, d)
		}
	}
	if len(clean) == 0 {
		return nil, errors.New("لا نطاقات في الطلب")
	}
	if phrase != AttestationPhrase {
		return nil, ErrNoAttestation
	}
	return &Scope{
		Domains:  clean,
		Phrase:   phrase,
		SignedAt: time.Now().UTC(),
		Sign:     signScope(clean, phrase),
	}, nil
}

// Resign يعيد توقيع نصاب (يُستدعى فقط بعد تعهد جديد صحيح — تغيير التعهد يعيد توقيع كل النطاقات).
func Resign(domains []string, phrase string) string {
	return signScope(domains, phrase)
}

// Validate يرفض أي تلاعب بالملف (التوقيع الذاتي).
func (s *Scope) Validate() error {
	if s == nil {
		return errors.New("no scope")
	}
	if s.Domains == nil || s.Phrase != AttestationPhrase {
		return ErrNoAttestation
	}
	if signScope(s.Domains, s.Phrase) != s.Sign {
		return errors.New("توقيع نصاب التفويض مخالف — ملف متلاعب به، يجب التعهد من جديد")
	}
	return nil
}

// Allows يطابق النطاق ونطاقاته الفرعية (نطاق أخي = أسرته).
func (s *Scope) Allows(host string) bool {
	if s == nil || s.Validate() != nil {
		return false
	}
	h := strings.ToLower(strings.TrimSpace(host))
	h = strings.TrimSuffix(strings.TrimPrefix(h, "www."), ".")
	for _, d := range s.Domains {
		if h == d || strings.HasSuffix(h, "."+d) {
			return true
		}
	}
	return false
}

// Gate ترفض العمليات خارج النصاب.
func (s *Scope) Gate(host string) error {
	if s.Allows(host) {
		return nil
	}
	return ErrOutOfScope
}

// Save/LoadScope حفظ/تحميل الذري عبر tmp+rename.
func (s *Scope) Save(path string) error {
	b, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := writeFileAtomic(path, b); err != nil {
		return err
	}
	// التحقق السابر (self-check) بعد الكتابة
	var reloaded Scope
	return readJSONInto(path, &reloaded)
}

func LoadScope(path string) (*Scope, error) {
	var s Scope
	b, err := readFile(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return nil, fmt.Errorf("نصاب تلف الـ JSON: %w", err)
	}
	return &s, nil
}
