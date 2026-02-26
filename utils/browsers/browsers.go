package browsers

import (
	"math/rand"
	"regexp"
	"strconv"

	tls "github.com/refraction-networking/utls"
)

var chromeVersionPattern = regexp.MustCompile(`Chrome/(\d+)`)
var firefoxVersionPattern = regexp.MustCompile(`Firefox/(\d+)`)

var allBrowserProfiles []BrowserProfile

func init() {
	for _, userAgent := range ChromeUserAgents {
		version := extractChromeVersion(userAgent)
		fingerprint := matchChromeFingerprint(version)
		allBrowserProfiles = append(allBrowserProfiles, BrowserProfile{
			UserAgent:      userAgent,
			Headers:        ChromeHeaders,
			TLSFingerprint: fingerprint,
			BrowserFamily:  "chrome",
		})
	}

	for _, userAgent := range FirefoxUserAgents {
		version := extractFirefoxVersion(userAgent)
		fingerprint := matchFirefoxFingerprint(version)
		allBrowserProfiles = append(allBrowserProfiles, BrowserProfile{
			UserAgent:      userAgent,
			Headers:        FirefoxHeaders,
			TLSFingerprint: fingerprint,
			BrowserFamily:  "firefox",
		})
	}
}

func extractChromeVersion(userAgent string) int {
	matches := chromeVersionPattern.FindStringSubmatch(userAgent)
	if len(matches) < 2 {
		return 0
	}
	version, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return version
}

func extractFirefoxVersion(userAgent string) int {
	matches := firefoxVersionPattern.FindStringSubmatch(userAgent)
	if len(matches) < 2 {
		return 0
	}
	version, err := strconv.Atoi(matches[1])
	if err != nil {
		return 0
	}
	return version
}

func matchChromeFingerprint(version int) tls.ClientHelloID {
	for _, mapping := range ChromeFingerprintsByVersion {
		if version >= mapping.MinVersion {
			return mapping.Fingerprint
		}
	}
	return tls.HelloChrome_58
}

func matchFirefoxFingerprint(version int) tls.ClientHelloID {
	for _, mapping := range FirefoxFingerprintsByVersion {
		if version >= mapping.MinVersion {
			return mapping.Fingerprint
		}
	}
	return tls.HelloFirefox_55
}

func SelectRandomProfile() BrowserProfile {
	return allBrowserProfiles[rand.Intn(len(allBrowserProfiles))]
}