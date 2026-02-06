package tvdb

import (
	"net/http"
	"time"
)

type client struct {
	httpClient  *http.Client
	token       string
	tokenExpiry time.Time
}
