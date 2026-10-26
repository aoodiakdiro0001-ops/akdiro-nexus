// examples/go/main.go — المجدول بـ 3 أسطر + لوحة القياس — AKDIRO-NEXUS v1.0.0
// الشرط: نطاق الهدف مفوَّض بالتعهد في ملف consent.scope.json.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/baran/akdiro-nexus/pkg/nexus"
)

func main() {
	// (1) جلسة: نصاب موقَّع + قوالب البروكسي (أعي عمودك أيIP هنا)
	sess, err := nexus.NewSession(nexus.SessionConfig{
		ScopePath:      "./consent.scope.json",
		DatacenterURL:  "http://USER-session-{session}:PASS@gate.anyip-provider.io:8000",
		ResidentialURL: "http://USER-session-{session}:PASS@gate.anyip-provider.io:9000",
	})
	if err != nil {
		panic(err)
	}

	// (2) جلب مربوط: تفويض → جدولة AIMD → طبقة ذكية → بصمة Chrome → قياس → Hot-Swap
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	rep, err := sess.Fetch(ctx, "https://allowed.example.com/api/data", "GET")
	if err != nil {
		panic(err) // رفض قانوني 403؟ أضف النطاق بتعهد عبر nexus.NewScope(...)
	}

	// (3) القياس الحي — نسبة نجاح مقاسة لا وعود
	fmt.Printf("الحالة %d — الطبقة %s — زمن %dms — تبديل IP: %v\n",
		rep.Status, rep.Tier, rep.LatencyMS, rep.Rotated)
	m := sess.Metrics().Snapshot()
	fmt.Printf("القياس الحي: probes=%d oks=%d blocked=%d success=%v\n",
		m["probes"], m["oks"], m["blocked"], m["success_rate"])
}
