package cfbypass

import (
	"context"
	"fmt"
	"math/rand"
	"metachan/utils/browsers"
	"net"
	"net/http"
	"net/http/cookiejar"
	"time"

	utls "github.com/refraction-networking/utls"
)

func NewCloudflareClient(timeout time.Duration) *CloudflareClient {
	selectedProfile := browsers.SelectRandomProfile()

	cookieJar, _ := cookiejar.New(nil)

	transport := &http.Transport{
		DialTLSContext: func(ctx context.Context, network, address string) (net.Conn, error) {
			return dialTLSWithFingerprint(ctx, network, address, selectedProfile.TLSFingerprint)
		},
		DisableKeepAlives: false,
		MaxIdleConns:      10,
		IdleConnTimeout:   90 * time.Second,
	}

	return &CloudflareClient{
		HttpClient: &http.Client{
			Timeout:   timeout,
			Transport: transport,
			Jar:       cookieJar,
		},
		BrowserProfile: selectedProfile,
	}
}

func dialTLSWithFingerprint(ctx context.Context, network string, address string, fingerprint utls.ClientHelloID) (net.Conn, error) {
	dialer := &net.Dialer{}
	rawConnection, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to dial %s: %w", address, err)
	}

	host, _, err := net.SplitHostPort(address)
	if err != nil {
		rawConnection.Close()
		return nil, fmt.Errorf("failed to split host and port from %s: %w", address, err)
	}

	clientHelloSpec, specErr := utls.UTLSIdToSpec(fingerprint)
	if specErr != nil {
		rawConnection.Close()
		return nil, fmt.Errorf("failed to convert fingerprint to spec: %w", specErr)
	}

	for _, extension := range clientHelloSpec.Extensions {
		if alpnExtension, isALPN := extension.(*utls.ALPNExtension); isALPN {
			alpnExtension.AlpnProtocols = []string{"http/1.1"}
			break
		}
	}

	utlsConnection := utls.UClient(rawConnection, &utls.Config{
		ServerName: host,
	}, utls.HelloCustom)

	if applyErr := utlsConnection.ApplyPreset(&clientHelloSpec); applyErr != nil {
		rawConnection.Close()
		return nil, fmt.Errorf("failed to apply TLS preset for %s: %w", address, applyErr)
	}

	if handshakeErr := utlsConnection.HandshakeContext(ctx); handshakeErr != nil {
		rawConnection.Close()
		return nil, fmt.Errorf("TLS handshake failed for %s: %w", address, handshakeErr)
	}

	return utlsConnection, nil
}

func AddJitter(baseDelay time.Duration) time.Duration {
	jitterRange := float64(baseDelay) * 0.4
	jitterOffset := rand.Float64()*jitterRange - jitterRange/2
	return baseDelay + time.Duration(jitterOffset)
}
