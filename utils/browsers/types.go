package browsers

import (
	tls "github.com/refraction-networking/utls"
)

type BrowserProfile struct {
	UserAgent      string
	Headers        map[string]string
	TLSFingerprint tls.ClientHelloID
	BrowserFamily  string
}

type FingerprintMapping struct {
	MinVersion  int
	Fingerprint tls.ClientHelloID
}