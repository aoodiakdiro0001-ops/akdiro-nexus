# AKDIRO-NEXUS ⚡

### منسّق الاستخبارات المستقل — Target-Aware Profiling · Adaptive Scheduling · Smart Proxy Routing · vulnradar-Core

> بنية = فكرتك حرفيًا: **Profiler → Router → Scheduler → Radar** — بأمانة هندسية كاملة.
> المالك والمطوّر: **Baran** — جميع الحقوق محفوظة © 2026. الترخيص: **Pro — مدى الحياة** (مدفوع $19 مع مفتاح).

**المستودع الرسمي:** [github.com/aoodiakdiro0001-ops/akdiro-nexus](https://github.com/aoodiakdiro0001-ops/akdiro-nexus)

---

## 🚀 التشغيل السريع — المجدول بـ 3 أسطر

**Go:**
```go
s, _ := nexus.NewSession(nexus.SessionConfig{ScopePath: "./consent.scope.json"})
rep, _ := s.Fetch(context.Background(), "https://allowed.example.com/api", "GET")
fmt.Println(rep.Status, rep.Tier, rep.LatencyMS, rep.Rotated)
```

**Python (REST):** `requests.get("http://127.0.0.1:48213/api/fetch", params={"url": ...})` — النموذج: [examples/python/quickstart.py](examples/python/quickstart.py)
**curl:** `curl -s "localhost:48213/api/fetch?url=…"`
**مرجع كامل لكل نقطة نهاية:** [docs/API.ar.md](docs/API.ar.md)

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
| «نسبة نجاح 99.9%» | ❌ لا نمنح نسبًا وعدًا — نجاح كل هدف **يُقاس ويُعرض حيًا** في `/metrics` ولوحة المراقبة |
| «تجاوز كل defenses» | ❌ بصمة ClientHello تحاكي Chrome عبر **uTLS** (TLS1.2/1.3/HTTP1.1)؛ بصمات **HTTP/2 وQUIC غير منفذة** — مُعلنة |
| «حل Cloudflare» | ❌ لا حلول كابشا ولا تفاعل مع تحديات — موقف ثابت في أي إصدار |
| «بنية لا رفض فيها» | ✅ نعم: 30+ اختبارًا + race-clean + بوابة قانونية موقَّعة |

## 📊 المراقبة الحية (جديد v1.0.0)

- `GET /metrics` — **Prometheus صيغ قياسية** (بلا سحابة): `akdiro_probes_total`, `akdiro_success_rate`, `akdiro_host_blocked_total{host=…}`…
- `GET /api/metrics` — JSON: عدادات + نسبة النجاح **المقاسة** لكل مضيف.
- اللوحة المدمجة تعرض: طلبات/نجاح/منع/تبديلات IP + جدول لكل مضيف بنسبة حية.

## 🔒 الترخيص المدفوع (جديد v1.0.0 — Ed25519 محلي 100%)

فكرة بلا خوادم: المفتاح `NL1.<payload>.<sig>` موقَّع من مفتاحك الخاص (المالك)، والتحقق عبر المفتاح العمومي المضمّن في الثنائي:

```bash
# إصدار مفتاح للعميل (من حسابك):
go run ./cmd/issue-key init          # مرة واحدة — license-issuer/issuer.key (سري)
go run ./cmd/issue-key mint -email client@corp.com -tier pro-lifetime
# تفعيل عند العميل:
curl -s -X POST localhost:48213/api/license/activate -d '{"key":"NL1.eyJ…"}'
# النسخة الموزعة المدفوعة تُجبر التفعيل:
NEXUS_REQUIRE_LICENSE=1 ./nexusd     # 402 لكل نقاط التشغيل إلى التفعيل
```

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

`/api/health` · `/api/version` · `/api/license` + `/api/license/activate` (Ed25519) · `/api/capabilities` · `/api/honesty` · `/api/scope` · `/api/profile?url=` · `/api/fetch?url=` (الجلسة المربوطة) · `/api/engine/scan` (المحرك المتزامن) · `/api/findings` (GET/POST) · `/api/egress` · `/api/metrics` + `/metrics` (Prometheus) — **المرجع الكامل بلمستجيب:** [docs/API.ar.md](docs/API.ar.md)

## حدود v1.0 — مكتوبة لا مخفية

- التوجيه عبر anyIP يُركب بأي قالب بوّابة؛ الترقية الذاتية داخل `Session.Fetch` (direct → datacenter عند المنع → residential عند Hot-Swap) مُراجعة بالاختبارات — **الربط بأعمدة anyIP الرسمية في v1.1** (لا نخترع API — نُعلن ذلك).
- صور h2/QUIC غير منفذة (معلنة في `/api/honesty` والواجهة).
- داخل المضيف الواحد: نَبض مُهَدَّد (واحد في وقته) — التوازن الحقيقي عبر مضيفين مختلفين (`/api/engine/scan` حتى 32 خيطًا).
- المراقبة لكل-مضيف بذاكرة (بدون استمرارية restart) — دفتر الثغرات والنصاب والترخيص فقط تُحفظ ذريًا.

---

## الحقوق والدعم

**AKDIRO-NEXUS** © 2026 — المالك: **Baran** — جميع الحقوق محفوظة. الترخيص **Pro مدفوع — مدى الحياة**.
📧 [baranavisat@gmail.com](mailto:baranavisat@gmail.com) · 📞 [+90 531 326 37 22](tel:+905313263722) · 💬 [wa.me/905313263722](https://wa.me/905313263722)
