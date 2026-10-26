# PARTNER_PITCH.md — رسالة الشراكة (anyIP / Isaac) — Baran © 2026

> الهدف النصي من نصيحتك: إطلاق **v1.0.0** بدليل مادي (CI أخضر) مقابل رصيد بروكسي مجاني ودعم تجاري.
> النسخة الإنجليزية للإرسال الرسمي، والعربية للتفاوض. البريد المرسل: baranavisat@gmail.com.

---

## ✉️ الإنجليزية (أرسلها كما هي)

**Subject: AKDIRO-NEXUS v1.0.0 launched — target-aware orchestration ready to route through anyIP**

Hi Isaac,

I'm Baran — I've just shipped **AKDIRO-NEXUS v1.0.0**, an open-first, consent-gated scraping orchestration layer in Go (MIT-like codebase, commercial Pro tier). Today it has:

- **Target-aware profiler**: one probe request → WAF/challenge indicators → risk 1-5 → proxy tier decision (datacenter by default, residential templates with sticky sessions only for hardened targets). Cost efficiency is a design goal, not marketing.
- **Adaptive scheduling**: per-host AIMD governor (success relaxes the pulse, 429/403 doubles it, capped jitter ±25%), with automatic IP/session hot-swap signaling after repeated blocks.
- **Real TLS realism**: Chrome ClientHello via uTLS; HTTP/2 & QUIC gaps declared openly.
- **Live, honest metrics**: measured success rate per host — Prometheus `/metrics` + JSON — no invented "99.9%" promises anywhere in my product copy.
- **Proof**: 30+ tests incl. race-clean engine tests, CI green, single 13MB binary, embedded dashboard.

Why anyIP:
Your residential+datacenter rotation and sticky sessions are exactly the routing tiers my engine selects between. I'd like to build a first-class anyIP integration (docs + `examples/anyip/` quickstart + your badge in the README) — and evaluate your service as the default gateway in the Pro distribution I sell to small security teams and infra shops.

What I'm asking for:
1. Trial credits for residential + datacenter bandwidth (to publish real integration numbers).
2. A 30-minute call with whoever owns partner/API docs, so my gate-level integration uses your canonical session endpoints (I'm deliberately not guessing them — my docs require the `{session}` template from the provider of record).

My repo & dashboard: https://github.com/aoodiakdiro0001-ops/akdiro-nexus
My terms (honesty matrix): https://github.com/aoodiakdiro0001-ops/akdiro-nexus#-ما-نقدمه-بلا-مبالغة

If the numbers look good on both sides, I'm ready to make anyIP the recommended provider in the product docs and to route my paid customers' templates to your gateways.

— Baran
baranavisat@gmail.com · +90 531 326 37 22 (WhatsApp ok)

---

## ✉️ العربية (نفس المضمون للتفاوض)

عيسى (Isaac) — أرسل لك اليوم إطلاق **AKDIRO-NEXUS v1.0.0**: طبقة تنسيق كشط مبنية بواجهة عربية ودفتر مصفوفة صدق كامل (نِسب النجاح تقاس ولا تُعَد وعدًا). والمنفذ جاهز ببنية توجيه ثلاثية (direct/datacenter/residential) وقوالب جلسة sticky — وأريد أن تكون anyIP هي البوابة الافتراضية في النسخة المدفوعة.

ما أطلبه: (1) رصيد تجريبي لعرض أرقام التكامل الحقيقية في توثيقي، (2) 30 دقيقة مع صاحب التوثيق لربط نقاط الجلسة الرسمية بصح (لا أخترع API). الدليل المادي: CI أخضر بـ30+ اختبارًا، دوائر سباق نظيفة، وثنائي 13MB مع لوحة مراقبة مدمجة.

المستودع: github.com/aoodiakdiro0001-ops/akdiro-nexus
نصيحة محِيطة: رسالتك — إطلاق v1.0.0 — هي خير وسط دليل مادي: النسخة شغّالة محليًا على جهازك بالأرقام أسفلها.

باران — baranavisat@gmail.com · wa.me/905313263722

---

## 📈 التقديرات لحساب العائد (بصدق — بلا آراء وهمية)

| السوق (B2B مصغّرة) | قناة الوصول | السعر المحدد | ملاحظة |
|---|---|---|---|
| فرق أمن سيبراني صغيرة | Reddit r/netsec / LinkedIn | $149/سنة (Pro-Team) | أقوى قيمة: دفتر الثغرات + القياس الصادق |
| مطورو بيانات مستقلون | Gumroad + X | $19 (كما حددت) | واجهة SDK بثلاثة أسطر تواجِبهم |
| بنى مراقبة/شِبكات | HN Show / Site بدليل CI | $49 (Pro) | خِدمة الجسد الوسيط أيIP الأعمدة |

**الثروة تنبع من الاستمرارية لا من العرض:** كل عميل Pro يستخدم قوالب anyIP = يرفع اعتراف شركتك بالجودة، وكل إصدار يرسل رسالة إطلاقة مع CI الأخضر — لمسة الشهادة التي يريد الشريك رؤيتها قبل منح الرصيد.
