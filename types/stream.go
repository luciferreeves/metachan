package types

type StreamAnimeStreamingSource struct {
	URL    string `json:"url"`
	Server string `json:"server"`
	Type   string `json:"type"`
}

type StreamAnimeStreaming struct {
	Sub []StreamAnimeStreamingSource `json:"sub"`
	Dub []StreamAnimeStreamingSource `json:"dub"`
}

type StreamSearchResult struct {
	ID          string  `json:"_id"`
	Name        string  `json:"name"`
	SubEpisodes int     `json:"sub_episodes"`
	DubEpisodes int     `json:"dub_episodes"`
	Similarity  float64 `json:"similarity"`
}

type StreamEpisodeStreamingResult struct {
	EpisodeNumber int
	Streaming     *StreamAnimeStreaming
}
