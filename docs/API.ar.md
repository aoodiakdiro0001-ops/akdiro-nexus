# مرجع واجهات AKDIRO-NEXUS (API Reference) — v1.0.0 — Baran © 2026

المنفذ الافتراضي: `http://127.0.0.1:48213` — كل الاستجابات JSON (باستثناء `/metrics` بتنسيق Prometheus).

> القاعدة: أي فحص/جلب/ثغرة على نطاق **غير مفوَّض** → `403` مع النص القانوني. الفكرة: الإذن ملف موقَّع لا وعد شفهي.

---

## 0) التشغيل السريع المجدول بـ 3 أسطر

**Go (SDK داخل العملية):**
```go
s, _ := nexus.NewSession(nexus.SessionConfig{ScopePath: "./consent.scope.json"})
rep, _ := s.Fetch(context.Background(), "https://allowed.example.com/api", "GET")
fmt.Println(rep.Status, rep.Tier, rep.LatencyMS) // الحالة، الطبقة، الزمن، وقد Rotated?: true
```
*(أمثلة كاملة: `examples/go/main.go`)*

**Python (عبر daemon REST):**
```python
import requests
s = requests.Session()
print(s.get("http://127.0.0.1:48213/api/fetch", params={"url":"https://allowed.example.com/"}).json())
```
*(أمثلة كاملة: `examples/python/quickstart.py`)*

**curl (سطر واحد):**
```bash
curl -s "localhost:48213/api/fetch?url=https://allowed.example.com/"
```

---

## 1) النظام والحالة

| المسار | النوع | الوصف |
|---|---|---|
| `/api/health` | GET | `{ok:true, service:"nexusd", version:"1.0.0"}` |
| `/api/version` | GET | المنتج والإصدار |
| `/api/capabilities` | GET | ما المنفذ وما غير المنفذ — **بصراحة مفهومة** (`h2_fingerprint:false`, `captcha_solver:false`) |
| `/api/honesty` | GET | نصوص مصفوفة الصدق الكاملة |

## 2) النصاب القانوني (Consent)

| المسار | النوع | الجسد | ملاحظات |
|---|---|---|---|
| `/api/scope` | GET | — | النطاقات المتعهَّد بها + `phrase_required` + حالة التوقيع |
| `/api/scope` | POST | `{"domains":["x.com"],"phrase":"أتعهد أنني مخوّل بالتحقق من هذا النطاق وأنه لي أو لدي إذن مكتوب"}` | العبارة **حرفيًا**؛ التلاعب بالملف لاحقًا = رفض Validate |

## 3) التشخيص والجلب

| المسار | النوع | معاملات | الاستجابة |
|---|---|---|---|
| `/api/profile?url=` | GET | `url` مفوَّض | `{"risk_level":5,"recommend_tier":"residential","markers":["cloudflare","challenge-script"],...}` |
| `/api/fetch?url=` | GET | `url` (بمخطط) | `{"status":200,"latency_ms":41,"tier":"direct","rotated":false,...}` — عبر الجلسة الكاملة (جدولة+توجيه+قياس) |
| `/api/egress` | GET | — | IP الخارجي الفعلي (فحص توصيل، هدفه `api.ipify.org`) |

## 4) محرك المهام المتزامن (B2B)

| المسار | النوع | الجسد | ملاحظة الصدق |
|---|---|---|---|
| `/api/engine/scan` | POST | `{"domains":["a.com","b.com"],"paths":["/","/api"],"workers":8}` | خيوط موازِنة **عبر** المضيفين (حتى 32)؛ داخل مضيف واحد الجدولة نَبض واحد في وقته — حدود صادقة: نطاقات 1..8، مسارات ≤16/نطاق |

الاستجابة: `{"jobs":8,"results":[{"url":"…","status":200,"latency_ms":9,"tier":"direct"},...],"honesty":"…"}`

## 5) الثغرات (vulnradar-core)

| المسار | النوع | الجسد |
|---|---|---|
| `/api/findings` | GET | `?target=a.com&min=4` (تقليل) — `{findings:[...], stats:{critical:n,…}}` |
| `/api/findings` | POST | `{"target":"a.com","kind":"xss","severity":4,"detail":"…","source":"vulnradar"}` — التكرار بالهاش: تُعيد `created:false` |

## 6) المراقبة الحية (B2B — المحلي، بلا سحابة)

| المسار | النوع | الوصف |
|---|---|---|
| `/api/metrics` | GET | JSON كامل: `{probes, oks, blocked, rotations, findings, success_rate, per_host:[{host,ok,block,rotate,success_rate}]}` |
| `/metrics` | GET | Prometheus text v0.0.4 — وصلها مباشرة بـ **grafana-agent/thanos** المحلي أو أي سكرابر برومetheus |

سلسلة المترائات المسخرة: `akdiro_probes_total`, `akdiro_ok_total`, `akdiro_blocked_total`, `akdiro_rotations_total`, `akdiro_findings_total`, `akdiro_scan_runs_total`, `akdiro_success_rate`, `akdiro_host_{ok,blocked,rotations}_total{host=…}`, `akdiro_latency_bucket_{10..5000,_inf}`.

## 7) الترخيص المدفوع (Ed25519 — محلي 100%)

| المسار | النوع | الجسد | الوصف |
|---|---|---|---|
| `/api/license` | GET | — | `{product,tier,owner,required,activated}` |
| `/api/license/activate` | POST | `{"key":"NL1.…"}` | تحقق توقيع محلي — بلا شبكة بلا خوادم؛ `400` مع رسالة الصدق عند الفشل |

- مفاتيح NL1 الموقَّعة من مفتاح المالك (خاص). المفتاح العمومي مضمّن بالثنائي.
- النسخة المفتوحة: تعمل بحرية وأنانية بلا حاجات. **التوزيع المدفوع**: شغّل `NEXUS_REQUIRE_LICENSE=1` → كل نقاط النهاية العاملة ترفض `402` إلا تفعيل/ترخيص/نظَم.

## 8) الترقيات الحاملة (B2B Scalability)

- **إعادة استعمال اتصالات TLS** عبر تخزين عملاء الجلسة (per-(proxy,session)) — لن يفتح مصافحة Chrome كل طلب (ضيرة سرعة حقيقية).
- **توازن عبر المضيفين** + **نبض قلب مُهَدَّد داخل المضيف** — أخلاقيات الكشط قيد الميزان.
- **الخطوة القادمة (v1.1)**: واجهات قائد قوائم Redis/Postgres بلا نسخ السحابة — التصميم مسموح بموديول (`Session`/`Governor` مستقلين).
