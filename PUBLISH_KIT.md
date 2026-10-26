# AKDIRO-NEXUS — حزمة النشر (PUBLISH_KIT) — Baran © 2026

> النصوص جاهزة للنسخ. النشر الاجتماعي بيدك (X/Reddit/HN/LinkedIn).

---

## 1) X (منشور)

**AKDIRO-NEXUS ⚡** — منسّق كشط ذكي Target-Aware بواجهة عربية:
– تشخيص الدفاعات (WAF/Cloudflare مؤشرات → درجة خطر 1-5)
– توجيه بروكسي ذكي: datacenter للأهداف الرخوة، residential للأPatrick المدرعة
– جدولة AIMD تكيفية + بصمة Chrome عبر uTLS
– دفتر ثغرات ضد التكرار — جاهز لـ vulnradar
كل هذا **بلا نسب نجاح وهمية**: النجاح يُقاس لكل هدف. والبنية ترفض أي نطاق بلا تَعُدّ أنت مخوّلًا.
الحقوق: © 2026 Baran — github.com/aoodiakdiro0001-ops/akdiro-nexus

---

## 2) Reddit (r/golang)

**العنوان:** I built AKDIRO-NEXUS: consent-gated, target-aware scraping orchestrator in Go — uTLS Chrome fingerprints, AIMD scheduler per host, honest success metrics (no "99.9%")  
**النص:**  
AKDIRO-NEXUS is a 4-layer local orchestrator: (1) a one-request target profiler scoring WAF indicators into risk 1-5, (2) proxy routing tiers (datacenter for easy targets, residential templates with sticky sessions for hardened ones — works with any gateway provider like anyIP), (3) a per-host AIMD governor (interval decays 20% on 2xx, doubles on 429/403 capped at 30s, ±25% jitter, signals IP hot-swap after 3 blocks), (4) a dedup-by-hash vulnerability findings ledger designed to plug into a vulnradar scanner.

Non-negotiable design choices:
- TLS ClientHello mimics Chrome via uTLS (HTTP/1.1) — HTTP/2 and QUIC fingerprints are NOT implemented and the API says so.
- No CAPTCHA solving, ever.
- A signed local consent gate: nothing scans a domain unless the operator attests authorization; tampered files fail validation and the API returns 403 otherwise.

Repo: https://github.com/aoodiakdiro0001-ops/akdiro-nexus (14 tests, race-clean).

---

## 3) Hacker News (Show HN)

**العنوان:** Show HN: AKDIRO-NEXUS – a consent-gated target-aware scraping orchestrator (Go)  
**النص التعليقي:**  
It sends one probe, classifies defenses (WAF headers, challenge scripts) into a 1-5 risk score, then routes: cheap datacenter IPs for easy targets, residential gateways only where needed. A per-host AIMD governor adapts rate (halve-ish on success, double on blocks, jitter ±25%) and flags proxy rotation after 3 consecutive blocks. Findings land in a hash-deduped ledger. TLS is Chrome-mimicking via uTLS; the honesty matrix in /api/honesty documents exactly what is NOT implemented (h2/QUIC fingerprints, CAPTCHA solving — refused by policy). Everything sits behind a signed consent file: no authorization attestation, no scan — the API returns 403 instead of pretending. 14 tests incl. race, single binary ~13MB with embedded RTL UI.

---

## 4) النشر المدفوع (Pro)

الترخيص: **AKDIRO-NEXUS Pro — مدى الحياة** للمالك Baran (مفعّل من المصدر، مفتاح محلي موقَّع sha256).
قنوات البيع المقترحة (اتخذه بيدك: حساب + شاشة + تسعير اختياري 69–149 دولار):
- Gumroad / BuyMeACoffee المنتج: "AKDIRO-NEXUS Pro Lifetime Key"
- النسخة المجانية: كود مستودع open-source. النسخة المدفوعة: مفتاح برو + جسر anyIP المفصل + دعم واتساب مباشر.
- ملاحظة صدق تستحق قولها في صفحة البيع: «الأداة تعمل على أهدافك المخوّلة فقط — القيود قانونية وهندسية مكتوبة، والوعد رقم 99.9% غير موجود فيها.»

---

## 5) بطاقة الدعم (ثابتة في كل منشور)

باران (Baran)  
📧 baranavisat@gmail.com  
📞 هاتف/واتساب: +90 531 326 37 22 — https://wa.me/905313263722
