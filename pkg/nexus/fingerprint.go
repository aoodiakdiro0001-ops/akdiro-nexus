// fingerprint.go — طبقة البصمة الصادقة: ClientHello يحاكي Chrome عبر uTLS.
// الحقيقة: يغطي TLS1.2/1.3 على HTTP/1.1 — بصمات HTTP/2 وQUIC غير منفذة (نُعلن بلا تصبف).
package nexus

import (
	"bufio"
	"context"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	utls "github.com/refraction-networking/utls"
)

// FingerprintHonesty نص الصدق الثابت — يظهر في API والواجهة.
const FingerprintHonesty = "بصمة ClientHello تحاكي Chrome عبر uTLS (TLS1.2/1.3، HTTP/1.1). " +
	"بصمات HTTP/2 وQUIC ووظائف تحدي الكابشا غير منفذة — الأداة لا تحل التحديات التفاعلية. " +
	"البصمة ترفع التوافق ولا تقدم نسبة نجاح وعدًا — نسبة النجاح يُقاس معها لكل هدف."

// Dialer خيارات بناء العميل المتوافق المتصفح.
type Dialer struct {
	ProxyURL   string // "" = مباشر؛ غير ذلك: بوابة بروكسي HTTP/HTTPS (مثل أي نقطة anyIP)
	SkipVerify bool   // للاختبارات فقط — لا تستفاد في الإنتاج عمدًا
}

// Client يبني *http.Client ببصمة Chrome.
func (d Dialer) Client() (*http.Client, error) {
	var proxyURL *url.URL
	if d.ProxyURL != "" {
		pu, err := url.Parse(d.ProxyURL)
		if err != nil {
			return nil, fmt.Errorf("بروكسي غير صالح: %w", err)
		}
		proxyURL = pu
	}

	transport := &http.Transport{
		ForceAttemptHTTP2: false, // بصمة h2 غير منفذة — لا ادعاء
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return d.dialTLS(ctx, addr, proxyURL)
		},
		TLSHandshakeTimeout: 10 * time.Second,
	}
	return &http.Client{Transport: transport, Timeout: 30 * time.Second}, nil
}

// Capabilities تعلن الحقيقة في API.
func (d Dialer) Capabilities() map[string]any {
	return map[string]any{
		"tls_fingerprint":  true,
		"h2_fingerprint":   false,
		"quic_fingerprint": false,
		"captcha_solver":   false,
		"honesty_note":     FingerprintHonesty,
	}
}

// dialTLS — يمر عبر نفق البروكسي (CONNECT إن وجد) ثم مصافحة uTLS على الاتصال.
func (d Dialer) dialTLS(ctx context.Context, addr string, proxyURL *url.URL) (net.Conn, error) {
	host := hostName(addr)
	if proxyURL != nil {
		conn, err := dialProxyCONNECT(ctx, proxyURL, addr)
		if err != nil {
			return nil, fmt.Errorf("CONNECT عبر البروكسي: %w", err)
		}
		return finishUTLS(ctx, conn, host, d.SkipVerify)
	}

	var nd net.Dialer
	conn, err := nd.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	return finishUTLS(ctx, conn, host, d.SkipVerify)
}

func hostName(addr string) string {
	if h, _, err := net.SplitHostPort(addr); err == nil {
		return h
	}
	return addr
}

// finishUTLS مصافحة بنقش Chrome (SNI + ALPN http/1.1، بلا h2 للصدق الحالي).
func finishUTLS(ctx context.Context, conn net.Conn, host string, skipVerify bool) (net.Conn, error) {
	cfg := &utls.Config{ServerName: host, InsecureSkipVerify: skipVerify}
	uconn := utls.UClient(conn, cfg, utls.HelloChrome_Auto)
	if err := uconn.HandshakeContext(ctx); err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("مصافحة بصمة TLS: %w", err)
	}
	return uconn, nil
}

// dialProxyCONNECT يفتح نفقًا HTTP عبر بروكسي مع اعتمادات من URL.
func dialProxyCONNECT(ctx context.Context, proxyURL *url.URL, target string) (net.Conn, error) {
	var nd net.Dialer
	conn, err := nd.DialContext(ctx, "tcp", proxyURL.Host)
	if err != nil {
		return nil, err
	}

	reqLine := fmt.Sprintf("CONNECT %s HTTP/1.1\r\nHost: %s\r\n", target, target)
	if u := proxyURL.User; u != nil {
		if pass, ok := u.Password(); ok {
			token := base64.StdEncoding.EncodeToString([]byte(u.Username() + ":" + pass))
			reqLine += "Proxy-Authorization: Basic " + token + "\r\n"
		}
	}
	reqLine += "\r\n"
	if _, err := conn.Write([]byte(reqLine)); err != nil {
		_ = conn.Close()
		return nil, err
	}

	br := bufio.NewReader(conn)
	resp, err := http.ReadResponse(br, nil)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		_ = conn.Close()
		return nil, fmt.Errorf("بوابة البروكسي رفضت CONNECT: %s", resp.Status)
	}
	if br.Buffered() > 0 {
		_ = conn.Close()
		return nil, fmt.Errorf("بوابة البروكسي أرسلت بيانات قبل اكتمال CONNECT — سلوك غير قياسي مرفوض")
	}
	return conn, nil
}
