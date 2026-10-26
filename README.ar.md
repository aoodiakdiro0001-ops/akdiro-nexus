# AKDIRO-NEXUS ⚡

### منسّق الاستخبارات المستقل — Target-Aware Profiling · Adaptive Scheduling · Smart Proxy Routing · vulnradar-Core

> البنية = فكرتك حرفيًا: **Profiler → Router → Scheduler → Radar** — بأمانة هندسية كاملة.
> المالك والمطوّر: **Baran** — جميع الحقوق محفوظة © 2026. الترخيص: **Pro — مدى الحياة** (مدفوع).

**المستودع الرسمي:** [github.com/aoodiakdiro0001-ops/akdiro-nexus](https://github.com/aoodiakdiro0001-ops/akdiro-nexus)

---

## طبقات النظام (الأربع — كما صممتَها)

| الطبقة | وحدة Go | الحالة | ملاحظة الصدق |
|---|---|---|---|
| 1️⃣ Target Profiler & Fingerprint Engine | `pkg/nexus/profiler.go` + `fingerprint.go` | ✅ منفذة | طلب استكشافي واحد → مؤشرات CF/Sucuri/WAF + سكربتات التحدي → درجة خطر 1-5. **مؤشرات لا يقين** |
| 2️⃣ Dynamic anyIP Routing & Session Manager | `pkg/nexus/routing.go` | ✅ بنية + قوالب | قوالب بوابة anyIP القياسية (`{session}` إلزامي) بلا اختراع API — URLك أنت من حساب anyIP. جلسات sticky/rotate جاهزة. **الربط الفعلي في v0.2** |
| 3️⃣ Adaptive Rate & Retry Scheduler | `pkg/nexus/scheduler.go` | ✅ منفذة | AIMD لكل مضيف: هدوء يخفف النبض، 429/403 يضاعف فورًا (سقف 30s)، Jitter ±25%، |إخفاقات| ≥ 3 → إشارة Hot-Swap |
| 4️⃣ vulnradar Aggregator | `pkg/radar/store.go` | ✅ منفذة | دفتر JSONL ذري + منع التكرار بهاش + ترتيب الشدة 1-5 — نواة جاهزة ليتصل بها vulnradar |

## ما نقدمه بلا مبالغة (Honesty Matrix)

| الادعاء الشائع | نحن نقول |
|---|---|
| «نسبة نجاح 99.9%» | ❌ لا نمنح نسبًا وعدًا — نجاح كل هدف **يُقاس ويخزن** في سجلك |
| «تجاوز كل defenses» | ❌ بصمة ClientHello تحاكي Chrome عبر **uTLS** (TLS1.2/1.3/HTTP1.1)؛ بصمات **HTTP/2 وQUIC غير منفذة** — مُعلنة |
| «حل Cloudflare» | ❌ لا حلول كابشا ولا تفاعل مع تحديات — موقف ثابت في أي إصدار |
| «بنية لا رفض فيها» | ✅ نعم: 4 disagreedاختبارات + race-clean + بوابة قانونية |

## 🔐 البوابة القانونية (هذا ليس زينة)

لا يشتغل النظام إلا على **نطاقات متعهَّد بها**: `POST /api/scope` مع العبارة الإلزامية
«أتعهد أنني مخوّل بالتحقق من هذا النطاق وأنه لي أو لدي إذن مكتوب».
الملف موقَّع sha256 محليًا — أي تلاعب يُرفض. بدون تعهد: **403 فورًا** (جربها في الدليل أدناه).

## 🧑‍🔬 دليل تشغيل 30 ثانية

```bash
go run ./cmd/nexusd            # http://127.0.0.1:48213
# من دون نصاب — الاختبار القانوني يرفض:
curl -s "localhost:48213/api/profile?url=https://app.example.com"
# إضافة نطاقك بتعهد:
curl -s -X POST localhost:48213/api/scope -H 'Content-Type: application/json' \
  -d '{"domains":["app.example.com"],"phrase":"أتعهد أنني مخوّل بالتحقق من هذا النطاق وأنه لي أو لدي إذن مكتوب"}'
# الآن التشخيص يعمل:
curl -s "localhost:48213/api/profile?url=https://app.example.com"
# دفع ثغرة من vulnradar:
curl -s -X POST localhost:48213/api/findings -H 'Content-Type: application/json' \
  -d '{"target":"app.example.com","kind":"waf","severity":4,"detail":"cv-detail","source":"vulnradar"}'
```

## Capabilities APIs

`/api/health` · `/api/version` · `/api/license` (Pro مفعّل للمالك) · `/api/capabilities` (TLS ✓ h2 ✗ كابشا ✗ — صراحة) · `/api/honesty` · `/api/scope` · `/api/profile?url=` · `/api/findings` (GET/POST) · `/api/egress` (IP الخارجي)

## حدود v0.1 — مكتوبة لا مخفية

- تشخيص بصمة Chrome باتجاه مباشر؛ التوجيه عبر أيIP يُركب بأي قالب بوّابة لكن آليته الآلية بعد v0.2 (المنطق والخلايا جاهزة: `ForTier/Expand/TestEgress` + اختبارات).
- صور h2/QUIC غير منفذة؛ الجدولة تعمل في المكتبة واختباراتها — الوصلة أوتوماتيكيًا بمهمة متعددة الطلبات في v0.2.
- المراقبة لكل-مضيف بذاكرة لكن بدون استمرارية restart (دفتر الثغرات فقط يُحفظ).

---

## الحقوق والدعم

**AKDIRO-NEXUS** © 2026 — المالك: **Baran** — جميع الحقوق محفوظة. الترخيص **Pro مدفوع — مدى الحياة**.
📧 [baranavisat@gmail.com](mailto:baranavisat@gmail.com) · 📞 [+90 531 326 37 22](tel:+905313263722) · 💬 [wa.me/905313263722](https://wa.me/905313263722)
