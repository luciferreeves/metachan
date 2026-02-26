package browsers

import (
	tls "github.com/refraction-networking/utls"
)

var ChromeFingerprintsByVersion = []FingerprintMapping{
	{MinVersion: 120, Fingerprint: tls.HelloChrome_120},
	{MinVersion: 114, Fingerprint: tls.HelloChrome_114_Padding_PSK_Shuf},
	{MinVersion: 112, Fingerprint: tls.HelloChrome_112_PSK_Shuf},
	{MinVersion: 106, Fingerprint: tls.HelloChrome_106_Shuffle},
	{MinVersion: 102, Fingerprint: tls.HelloChrome_102},
	{MinVersion: 100, Fingerprint: tls.HelloChrome_100},
	{MinVersion: 96, Fingerprint: tls.HelloChrome_96},
	{MinVersion: 87, Fingerprint: tls.HelloChrome_87},
	{MinVersion: 83, Fingerprint: tls.HelloChrome_83},
	{MinVersion: 72, Fingerprint: tls.HelloChrome_72},
	{MinVersion: 70, Fingerprint: tls.HelloChrome_70},
	{MinVersion: 62, Fingerprint: tls.HelloChrome_62},
	{MinVersion: 0, Fingerprint: tls.HelloChrome_58},
}

var FirefoxFingerprintsByVersion = []FingerprintMapping{
	{MinVersion: 120, Fingerprint: tls.HelloFirefox_120},
	{MinVersion: 105, Fingerprint: tls.HelloFirefox_105},
	{MinVersion: 102, Fingerprint: tls.HelloFirefox_102},
	{MinVersion: 99, Fingerprint: tls.HelloFirefox_99},
	{MinVersion: 65, Fingerprint: tls.HelloFirefox_65},
	{MinVersion: 63, Fingerprint: tls.HelloFirefox_63},
	{MinVersion: 56, Fingerprint: tls.HelloFirefox_56},
	{MinVersion: 0, Fingerprint: tls.HelloFirefox_55},
}