package streaming

import (
	"net/http"
	"time"
)

type client struct {
	httpClient *http.Client
	headers    http.Header
	maxRetries int
	backoff    time.Duration
}
