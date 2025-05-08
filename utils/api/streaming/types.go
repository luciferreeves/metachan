package streaming

import "net/http"

// AllAnimeClient provides methods for interacting with the AllAnime API
type AllAnimeClient struct {
	client  *http.Client
	headers http.Header
}

// AnimeStreamingSource represents a single streaming source for an episode
type AnimeStreamingSource struct {
	URL    string `json:"url"`
	Server string `json:"server"`
	Type   string `json:"type"` // direct or embed
}

// AnimeStreaming represents all available streaming sources for an episode
type AnimeStreaming struct {
	Sub []AnimeStreamingSource `json:"sub"`
	Dub []AnimeStreamingSource `json:"dub"`
}

// StreamingSearchResult represents a search result from streaming providers
type StreamingSearchResult struct {
	ID          string  `json:"_id"`
	Name        string  `json:"name"`
	SubEpisodes int     `json:"sub_episodes"`
	DubEpisodes int     `json:"dub_episodes"`
	Similarity  float64 `json:"similarity"`
}

// EpisodeStreamingResult contains streaming sources for a specific episode
// Used for parallel streaming source fetching
type EpisodeStreamingResult struct {
	EpisodeNumber int
	Streaming     *AnimeStreaming
}
