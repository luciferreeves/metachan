package cfbypass

import (
	"metachan/utils/browsers"
	"net/http"
)

type CloudflareClient struct {
	HttpClient     *http.Client
	BrowserProfile browsers.BrowserProfile
}