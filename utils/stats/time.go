package stats

import "time"

func GetCurrentTimestamp() string {
	// Get the current UTC time and format it as RFC3339
	return time.Now().UTC().Format(time.RFC3339)
}
