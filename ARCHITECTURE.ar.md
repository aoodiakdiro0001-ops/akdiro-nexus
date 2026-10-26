# AKDIRO-NEXUS — العمارة الفنية (v0.1) — Baran © 2026

```
[ Target URL (مفوَّض بتعهد) ]
        │
        ▼
┌────────────────────────────────────────────┐
│ 1. Profiler + Fingerprint Engine           │
│    probe واحد → مؤشرات + Risk 1..5         │
│    uTLS Chrome-ClientHello (HTTP/1.1)      │
└──────────────────┬─────────────────────────┘
                   ▼
┌────────────────────────────────────────────┐
│ 2. Dynamic anyIP Routing                   │
│    Tier: direct ─ datacenter ─ residential │
│    قوالب {session} sticky/rotate           │
└──────────────────┬─────────────────────────┘
                   ▼
┌────────────────────────────────────────────┐
│ 3. Adaptive Scheduler (AIMD × Jitter)      │
│    200 → −20% (abase) · 429/403 → ×2       │
│    block≥3 → إشارة Hot-Swap                │
└──────────────────┬─────────────────────────┘
                   ▼
┌────────────────────────────────────────────┐
│ 4. vulnradar-Core Aggregator               │
│    JSONL ذري + هاش منع تكرار + شِدّة 1..5   │
└────────────────────────────────────────────┘
```

## مضان الحقيقة في كل طبقة

**1. Profiler:** يقرأ `Cf-Mitigated/Cf-Ray/X-Sucuri-Id/X-WAF*/Server` + `challenge-script` regex. التقييم: 1 عادي، +1 لكل مؤشر WAF، +2 مع سكربت تحدي، +1 على 429/403، سقف 5. طبقة موصى بها: `direct/datacenter/residential`. **مؤشرات لا يقين — النص ثابت في الرد.**

**2. Router:** أي مزود بروكسي بوابة يعمل. القالب: `http://USER-session-{session}:PASS@gate.provider.io:8000`. `ForTier` يعيد الـ URL بسلسلة الجلسة (sticky). `TestEgress` يطلب `api.ipify.org` ويخزن الـ IP الفعلي (ثبت في التجربة: 85.100.57.243 على الطبقة direct).

**3. Scheduler:** `Governor.Wait()` = فاصل + Jitter ±25% (0.75–1.25). `Result(code)`:
- `2xx → interval*=0.8` حتى الأساس (نوضح: عملية عادلة بلا قطن),
- `429/403/503 → interval*=2` حتى 30s,
- أخطاء أخرى ×1.1.
`ShouldRotate()`: 3 إخفاقات متتالية → طبقة استدساء + تبديل بروكسي عبر إعادة `Dialer.ProxyURL=ForTier(residential,sess2)`.

**4. Radar:** `Add(target,kind,severity,detail,source)` — `Hash=sha256(target|kind|severity|detail)` → تكرار=update بدون ازدواج (مُثبَت بالاختبار). `List(target,minSeverity)` بترتيب severity ثم زمن. `Stats`: critical/high/medium/low.

## البوابة القانونية (Consent Gate)

`consent.scope.json`: نطاقات + عبارة إلزامية + `Sign=sha256("nexus-consent-v1|domains|phrase")` — ملف موقَّع محليًا، أي تلاعب → `Validate()` يرفض. كل مسار سكيّن يشكك عبر `Gate(host)`؛ الأدخل خسر: **403.**

## المراجعة — اختبارات 14 موثاري
`consent_test` (4: تعهد، نطاقات فرعية، تلاعب، استمرارية) · `profiler_test` (5: عادي، مدرَع 5th، 429، مدخل خاطئ، تسمية المؤشرات) · `scheduler_test` (5: خفض، سقف، إشارة، إلغاء، حدود) · `fingerprint_test` (3: مصافحة uTLS حقيقية، capabilities صادقة، بروكسي خاسر) · `routing_test` (3: {session} إلزام، توسيع، ترتيب اقتصادي) · `radar_test` (4: dedupe، فِلترة، استمرارية، حدود) — **14 + race-clean.**

## ماذا في v0.2 (الترشيح الصادق)

1. **عامل الشغل الكامل**: Task Engine يربط Governor × Router × FingerprintClient في مهمة واحدة (Route → Wait → GET → Result → rotate عند الإشارة) — المكونات الثلاثة جاهزة ومختبرة، البطانة من شيء واحد.
2. **صور HTTP/2** (بصمة كاملة بـ Playwright/GFM من الواجهة الأم) وتركات هشة.
3. استمرارية الكوابح من restart للنصاب.
4. جسر vulnradar بمنقار JSON جاهز + SLA سيناريا (زمن التتبع).

> صراحة مكتوبة على حائط: الرقم `99.9%` بالوعد كذب؛ بالقياس أداة. جهزنا القياس.
