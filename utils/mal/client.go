package mal

import (
	"fmt"
	"math"
	"metachan/utils/cfbypass"
	"metachan/utils/logger"
	"metachan/utils/ratelimit"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const (
	malBaseURL      = "https://myanimelist.net"
	rateLimitPerSec = 4
	requestTimeout  = 30 * time.Second
	requestJitter   = 250 * time.Millisecond
	maxRetries      = 3
	backoffBase     = 2 * time.Second
)

var (
	rateLimiter      = ratelimit.NewRateLimiter(rateLimitPerSec, time.Second)
	cloudflareClient = cfbypass.NewCloudflareClient(requestTimeout)
)

func StopRateLimiters() {
	rateLimiter.Stop()
}

func makeRequest(targetURL string) (*goquery.Document, error) {
	var retries int

	for retries < maxRetries {
		rateLimiter.Wait()
		time.Sleep(cfbypass.AddJitter(requestJitter))

		request, err := http.NewRequest("GET", targetURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request for %s: %w", targetURL, err)
		}

		for headerName, headerValue := range cloudflareClient.BrowserProfile.Headers {
			if headerName == "Accept-Encoding" {
				continue
			}
			request.Header.Set(headerName, headerValue)
		}
		request.Header.Set("User-Agent", cloudflareClient.BrowserProfile.UserAgent)

		response, err := cloudflareClient.HttpClient.Do(request)
		if err != nil {
			retries++
			if retries >= maxRetries {
				return nil, fmt.Errorf("all retries exhausted for %s: %w", targetURL, err)
			}
			logger.Debugf("MALClient", "Request failed for %s (attempt %d/%d)", targetURL, retries, maxRetries)
			time.Sleep(getBackoffDuration(retries))
			continue
		}

		if response.StatusCode == http.StatusOK {
			document, parseErr := goquery.NewDocumentFromReader(response.Body)
			response.Body.Close()
			if parseErr != nil {
				return nil, fmt.Errorf("failed to parse HTML from %s: %w", targetURL, parseErr)
			}

			pageTitle := document.Find("title").Text()
			logger.Debugf("MALClient", "Page title for %s: %q", targetURL, pageTitle)

			htmlContent, _ := document.Html()
			if len(htmlContent) > 500 {
				htmlContent = htmlContent[:500]
			}
			logger.Debugf("MALClient", "HTML preview for %s: %s", targetURL, htmlContent)

			return document, nil
		}

		response.Body.Close()

		if response.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("resource not found: %s", targetURL)
		}

		if response.StatusCode >= 400 && response.StatusCode < 500 &&
			response.StatusCode != http.StatusTooManyRequests &&
			response.StatusCode != http.StatusForbidden {
			return nil, fmt.Errorf("client error %d for %s", response.StatusCode, targetURL)
		}

		retries++
		if retries >= maxRetries {
			return nil, fmt.Errorf("all retries exhausted for %s (status %d)", targetURL, response.StatusCode)
		}

		logger.Warnf("MALClient", "Status %d for %s (attempt %d/%d)", response.StatusCode, targetURL, retries, maxRetries)
		time.Sleep(getBackoffDuration(retries))
	}

	return nil, fmt.Errorf("all retries exhausted for %s", targetURL)
}

func getBackoffDuration(attempt int) time.Duration {
	exponentialDelay := time.Duration(float64(backoffBase) * math.Pow(2, float64(attempt-1)))
	return cfbypass.AddJitter(exponentialDelay)
}