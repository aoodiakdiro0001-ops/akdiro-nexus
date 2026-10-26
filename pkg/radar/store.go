// store.go — دفتر الثغرات: قلب لتحرك vulnradar (יותر JSONL ذري — وسع أس القوة الجديدة).
package radar

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
)

// Finding ثغرة سجلة من أي مصدر (vulnradar أو أداؤك يدويًا).
type Finding struct {
	ID       string    `json:"id"`
	Target   string    `json:"target"`
	Kind     string    `json:"kind"`
	Severity int       `json:"severity"` // 1..5
	Detail   string    `json:"detail"`
	Source   string    `json:"source"`
	Ts       time.Time `json:"ts"`
	Hash     string    `json:"hash"` // هاش الكشف — يمنع التكرار
}

// Ledger دفتر الثغرات (سجل عملي ومثبت "مظامية").
type Ledger struct {
	mu    sync.Mutex
	path  string
	items []Finding
}

// OpenLedger يفتح بستر (تحميل من JSONL).
func OpenLedger(path string) (*Ledger, error) {
	if path == "" {
		path = "findings.jsonl"
	}
	l := &Ledger{path: path}
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return l, nil
		}
		return nil, err
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var f Finding
		if json.Unmarshal([]byte(line), &f) == nil && f.ID != "" {
			l.items = append(l.items, f)
		}
	}
	return l, nil
}

func findingHash(target, kind string, severity int, detail string) string {
	sum := sha256.Sum256([]byte(target + "|" + kind + "|" + fmt.Sprint(severity) + "|" + detail))
	return hex.EncodeToString(sum[:12])
}

// Add يسجل كشفًا جديدًا — يعيد الحالي إن تكرار (لا ازدواج).
func (l *Ledger) Add(target, kind string, severity int, detail, source string) (*Finding, bool, error) {
	target = strings.TrimSpace(target)
	detail = strings.TrimSpace(detail)
	if target == "" || kind == "" {
		return nil, false, fmt.Errorf("الهدف والنوع مطلوبان")
	}
	if severity < 1 {
		severity = 1
	}
	if severity > 5 {
		severity = 5
	}
	h := findingHash(target, kind, severity, detail)
	l.mu.Lock()
	defer l.mu.Unlock()
	for i := range l.items {
		if l.items[i].Hash == h {
			return &l.items[i], false, nil
		}
	}
	now := time.Now().UTC()
	f := Finding{
		ID:     fmt.Sprintf("f%s%04x", now.Format("20060102150405"), now.UnixMicro()&0xFFFF),
		Target: target, Kind: kind, Severity: severity,
		Detail: detail, Source: source, Ts: now, Hash: h,
	}
	l.items = append(l.items, f)
	if err := l.saveLocked(); err != nil {
		return nil, false, err
	}
	return &f, true, nil
}

// List يستخرج تجميعة من المكتوس (dokepo إذا target فارغ).
func (l *Ledger) List(target string, minSeverity int) []Finding {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Finding, 0, len(l.items))
	for _, f := range l.items {
		if target != "" && f.Target != target {
			continue
		}
		if minSeverity > 1 && f.Severity < minSeverity {
			continue
		}
		out = append(out, f)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Severity != out[j].Severity {
			return out[i].Severity > out[j].Severity
		}
		return out[i].Ts.After(out[j].Ts)
	})
	return out
}

// Stats ملخص عام للواجهة.
func (l *Ledger) Stats() map[string]int {
	l.mu.Lock()
	defer l.mu.Unlock()
	st := map[string]int{"total": 0, "critical": 0, "high": 0, "medium": 0, "low": 0}
	st["total"] = len(l.items)
	for _, f := range l.items {
		switch f.Severity {
		case 5:
			st["critical"]++
		case 4:
			st["high"]++
		case 3:
			st["medium"]++
		default:
			st["low"]++
		}
	}
	return st
}

func (l *Ledger) saveLocked() error {
	var b strings.Builder
	for _, f := range l.items {
		j, err := json.Marshal(f)
		if err != nil {
			return err
		}
		b.Write(j)
		b.WriteByte('\n')
	}
	tmp := l.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(b.String()), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, l.path)
}
