package types

type AniskipInterval struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

type AniskipResult struct {
	Interval      AniskipInterval `json:"interval"`
	SkipType      string          `json:"skip_type"`
	SkipID        string          `json:"skip_id"`
	EpisodeLength float64         `json:"episode_length"`
}

type AniskipResponse struct {
	Found   bool            `json:"found"`
	Results []AniskipResult `json:"results"`
}
