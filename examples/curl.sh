# examples/curl.sh — دورة كاملة بأسطر قابلة للنسخ — Baran © 2026
# 1) تشغيل nexo daemon: go run ./cmd/nexusd   (المنفذ 48213)
# 2) تسجيل نطاقك بتعهد:
curl -s -X POST localhost:48213/api/scope -H 'Content-Type: application/json' \
  -d '{"domains":["my-domain.com"],"phrase":"أتعهد أنني مخوّل بالتحقق من هذا النطاق وأنه لي أو لدي إذن مكتوب"}'

# 3) المجدول المربوط بجلب واحد:
curl -s "localhost:48213/api/fetch?url=https://my-domain.com/api" | head -c 400
# ÷ مثال استجابة الملخص:
# {"url":"https://my-domain.com/api","status":200,"latency_ms":88,"tier":"datacenter","rotated":false,...}
