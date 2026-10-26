// license_test.go — التحقق المحلي 100%: مفتاح المالك الحقيقي يمر، التلاعب والتنسيق يُرفضان.
package nexus

import (
	"strings"
	"testing"
)

// مفتاح المالك الحقيقي (يُسك مرة، يتحقق محليًا دائمًا — بروكسي/خوادم لا وجود).
const ownerKey = "NL1.eyJlbWFpbCI6ImJhcmFuYXZpc2F0QGdtYWlsLmNvbSIsInRpZXIiOiJwcm8tbGlmZXRpbWUiLCJpc3N1ZWQiOjE3OTMwMzIwNjh9.SxnN6jMZpqzsfouJbX-0MACg2BgO-FyzntPN06rlP1DwlTnjc1z7YWOZGTfPvlJP6GIyY_CkpSFx8E_6WQDXDg"

func TestOwnerKeyVerifies(t *testing.T) {
	lp, err := VerifyLicenseKey(ownerKey)
	if err != nil {
		t.Fatalf("مفتاح المالك رُفض: %v", err)
	}
	if lp.Email != "baranavisat@gmail.com" || lp.Tier != "pro-lifetime" {
		t.Fatalf("بيانات الرخصة فاسد: %+v", lp)
	}
}

func TestLicenseTamperRejected(t *testing.T) {
	// تلاعب في البايت الأخير من التوقيع
	tk := ownerKey[:len(ownerKey)-2] + "AA"
	if _, err := VerifyLicenseKey(tk); err != ErrLicenseSig {
		t.Fatalf("توقيع مزوَّر قُبل: %v", err)
	}
	// تنسيقات فاسدة
	for _, bad := range []string{"", "LOST.aaa.bbb", "NL1.bad!sig", "NL1.cGF5bG9hZA.sig-with-dashes"} {
		if _, err := VerifyLicenseKey(bad); err == nil {
			t.Fatalf("قُبل تنسيق فاسد: %q", bad)
		}
	}
}

func TestLicenseTierValidation(t *testing.T) {
	if !licenseTierValid("pro-lifetime") || !licenseTierValid("sec-team") {
		t.Fatal("فئات صحيحة رفضت")
	}
	if licenseTierValid("free-forever") {
		t.Fatal("فئة مزوَّرة سُمح")
	}
}

func TestLicenseRequiredEnvHonesty(t *testing.T) {
	// النسخة المفتوحة: لا الإجبار — النسخة الموزعة تجارًا: NEXUS_REQUIRE_LICENSE=1
	if strings.Contains(IssuerPubKeysHex, "PLACEHOLDER") {
		t.Fatal("المفتاح العمومي المضمن لم يوضع")
	}
}
