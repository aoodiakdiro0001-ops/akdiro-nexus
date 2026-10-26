// fingerprint_test.go — مصافحة uTLS الحقيقية: عميل Chrome-baseline يشتغل على خادم TLS محلي.
package nexus

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFingerpintDialTLSLocalhost(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("hello-knowledge"))
	}))
	defer srv.Close()

	d := Dialer{SkipVerify: true} // شهادة اختبار موقّعة ذاتيًا فقط
	client, err := d.Client()
	if err != nil {
		t.Fatalf("Client: %v", err)
	}
	resp, err := client.Get(srv.URL + "/")
	if err != nil {
		t.Fatalf("مصافحة uTLS فشلت: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("الحالة %d", resp.StatusCode)
	}
	// صدق: resp.TLS قد تكون nil مع UConn المخصصة (net/http يلتقط حالة TLS لـ *tls.Conn فقط) —
	// الإثبات الفعلي هنا: 200 عبر مصافحة Chrome-hello الكاملة.
}

func TestDialerCapabilitiesHonesty(t *testing.T) {
	c := (Dialer{}).Capabilities()
	if v, ok := c["tls_fingerprint"]; !ok || v != true {
		t.Fatal("بصمة TLS غير معلنة")
	}
	if v, _ := c["h2_fingerprint"]; v != false {
		t.Fatal("يلاحظ h2 — نُعلن غير منفذ")
	}
	if v, _ := c["captcha_solver"]; v != false {
		t.Fatal("حل الكابشا مرفض نهائيًا — لا ادعاء")
	}
	if s, _ := c["honesty_note"].(string); len(s) < 20 {
		t.Fatal("ملاحظة الصدق قصيرة")
	}
}

func TestProxyURLParse(t *testing.T) {
	// بروكسيضرب خاطئ — لا يقع صامتًا
	d := Dialer{ProxyURL: ":%$bad:://"}
	if _, err := d.Client(); err == nil {
		t.Fatal("قُبل نمط بروكسي تالف")
	}
	// نمط صالح يمر
	d2 := Dialer{ProxyURL: "http://user-session-abc:pass@gate.anyip.io:8000"}
	if _, err := d2.Client(); err != nil {
		t.Fatalf("نمط سريعة سليم رفض: %v", err)
	}
}
