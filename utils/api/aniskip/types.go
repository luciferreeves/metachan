package aniskip

// AnimeSkipTimes represents skip time intervals for anime episodes
type AnimeSkipTimes struct {
	SkipType      string  `json:"skip_type"`
	StartTime     float64 `json:"start_time"`
	EndTime       float64 `json:"end_time"`
	EpisodeLength float64 `json:"episode_length"`
}

// EpisodeSkipResult contains skip times for a specific episode
type EpisodeSkipResult struct {
	EpisodeNumber int
	SkipTimes     []AnimeSkipTimes
}
