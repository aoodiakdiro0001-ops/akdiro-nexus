// license.go — الترخيص المدفوع: مفتاح Ed25519 موقَّع يُتحقق منه محليًا 100% — بلا خوادم، بلا تسريب.
// نمط المفتاح: NL1.<payload-base64url>.<sig-base64url> — payload JSON {email,tier,issued}.
// الإذن: المفتاح الخاص عند المالك (Baran) فقط؛ المستودع يعلن المفتاح العمومي للتَّحقق.
package nexus

import (
	"crypto/ed25519"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

// IssuerPubKeysHex مفتاح الإصدار العمومي المضمّن — يتحقق من مفاتيح NL1 بلا خوادم.
// (الخاص عند المالك Baran: license-issuer/issuer.key — مرفوع بلا شيء لأي مكان.)
const IssuerPubKeysHex = "bf2f0979a43b8a5c9389fa9b842021c335e3fd344689f5d5758afce546fd6321"

// licensePubKey الترميز السادي المضمّن — lazy.
func licensePubKey() (ed25519.PublicKey, error) {
	return hex.DecodeString(IssuerPubKeysHex)
}

// LicensePayload بيانات الرخصة.
type LicensePayload struct {
	Email  string `json:"email"`
	Tier   string `json:"tier"`   // pro-lifetime | pro-monthly | sec-team
	Issued int64  `json:"issued"` // unix seconds
}

// TierValid يحدد قيمة صحيحة معرفها سابقًا. 0 = بلا انتهاء (مدى الحياة).
func licenseTierValid(tier string) bool {
	switch tier {
	case "pro-lifetime", "pro-monthly", "sec-team":
		return true
	}
	return false
}

var (
	ErrLicenseFormat = errors.New("مفتاح مرفوض: الصيغة NL1.<payload>.<sig> غير صحيحة")
	ErrLicenseSig    = errors.New("مفتاح مرفوض: التوقيع غير صالح أو طُلب من مصدر غير المالك")
	ErrLicenseTier   = errors.New("مفتاح مرفوض: فئة الرخصة غير معروفة")
	ErrLicenseEmail  = errors.New("مفتاح مرفوض: البريد مفقود")
)

// MintLicenseKey يجعل مفتاحًا (يُستدعى بملف المفتاح الخاص من قيادة issue-key فقط).
func MintLicenseKey(email, tier string, issuedAt int64, priv ed25519.PrivateKey) (string, error) {
	if !licenseTierValid(tier) {
		return "", ErrLicenseTier
	}
	if strings.TrimSpace(email) == "" {
		return "", ErrLicenseEmail
	}
	p, err := marshalPayload(LicensePayload{Email: email, Tier: tier, Issued: issuedAt})
	if err != nil {
		return "", err
	}
	sig := ed25519.Sign(priv, p)
	return "NL1." + base64.RawURLEncoding.EncodeToString(p) + "." + base64.RawURLEncoding.EncodeToString(sig), nil
}

func marshalPayload(lp LicensePayload) ([]byte, error) {
	return json.Marshal(lp)
}

// VerifyLicenseKey يتحقق المحلي الكامل — بلا شبكة، بلا تسريب.
func VerifyLicenseKey(key string) (*LicensePayload, error) {
	parts := strings.Split(strings.TrimSpace(key), ".")
	if len(parts) != 3 || parts[0] != "NL1" {
		return nil, ErrLicenseFormat
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrLicenseFormat
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrLicenseFormat
	}
	pub, err := licensePubKey()
	if err != nil || len(pub) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("مفتاح التحقق المضمَّن تالف: %w", err)
	}
	if !ed25519.Verify(pub, payload, sig) {
		return nil, ErrLicenseSig
	}
	var lp LicensePayload
	if err := json.Unmarshal(payload, &lp); err != nil {
		return nil, ErrLicenseFormat
	}
	if !licenseTierValid(lp.Tier) {
		return nil, ErrLicenseTier
	}
	if strings.TrimSpace(lp.Email) == "" {
		return nil, ErrLicenseEmail
	}
	if lp.Issued > time.Now().Unix() {
		return nil, errors.New("مفتاح رخصة من المستقبل — غير مقبول")
	}
	return &lp, nil
}

// LicenseRequired وضع الإجبار للنسخات الموزعة مقابل مال (بيئة التاجر).
func LicenseRequired() bool { return os.Getenv("NEXUS_REQUIRE_LICENSE") == "1" }

// SaveActivated يحفظ الرخصة المفعّلة بذرية.
func SaveActivated(path string, lp *LicensePayload) error {
	b, err := json.Marshal(*lp)
	if err != nil {
		return err
	}
	return writeFileAtomic(path, b)
}

func LoadActivated(path string) (*LicensePayload, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var lp LicensePayload
	if err := json.Unmarshal(b, &lp); err != nil {
		return nil, err
	}
	return &lp, nil
}
