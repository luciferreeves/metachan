package malsync

import (
	"net/http"
	"time"
)

type client struct {
	httpClient *http.Client
	maxRetries int
	backoff    time.Duration
}
