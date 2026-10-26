#!/usr/bin/env python3
# بروتوكول المجدول بـ 3 أسطر — daemon REST — AKDIRO-NEXUS v1.0.0 (Baran © 2026)
# الشرط القانوني: النطاق مفوَّض بالتعهد (POST /api/scope أولًا).
import requests

NEXUS = "http://127.0.0.1:48213"
ATTEST = "أتعهد أنني مخوّل بالتحقق من هذا النطاق وأنه لي أو لدي إذن مكتوب"

# (0) مرة واحدة: تسجيل نصاب التفويض (مثال: 127.0.0.1 — جهازك أنت)
s = requests.Session()
r = s.post(f"{NEXUS}/api/scope", json={"domains": ["127.0.0.1"], "phrase": ATTEST})
assert r.ok, r.text

# (1) تشخيص الهدف: درجة خطر + طبقة موصى بها
probe = s.get(f"{NEXUS}/api/profile", params={"url": "http://127.0.0.1:48214/"}).json()
print("risk:", probe["risk_level"], "/5 → tier:", probe["recommend_tier"])

# (2) جلب مربوط عبر الجلسة (جدولة+توجيه+بصمة+قياس)
fetch = s.get(f"{NEXUS}/api/fetch", params={"url": "http://127.0.0.1:48214/"}).json()
print("status:", fetch["status"], "| tier:", fetch["tier"], "| latency:", fetch["latency_ms"], "ms", "| rotated:", fetch["rotated"])

# (3) المراقبة الحية: نسبة النجاح المقاسة — لا وعود
metrics = s.get(f"{NEXUS}/api/metrics").json()
print("probes:", metrics["probes"], "| blocked:", metrics["blocked"], "| success_rate:", round(metrics["success_rate"]*100, 1), "%")

# (بونص) محرك المهام المتزامن عبر مضيفين (B2B)
scan = s.post(f"{NEXUS}/api/engine/scan", json={
    "domains": ["127.0.0.1"], "paths": ["/", "/x"], "workers": 4}).json()
print("scan jobs:", scan["jobs"], "| done:", len(scan["results"]))

# (بونص) دفع ثغرة من vulnradar (منع تكرار تلقائي)
f = s.post(f"{NEXUS}/api/findings", json={
    "target": "127.0.0.1", "kind": "waf", "severity": 3, "detail": "log-header", "source": "vulnradar"}).json()
print("finding created:", f["created"])
