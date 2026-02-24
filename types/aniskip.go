package types

type AniskipInterval struct {
	StartTime float64 `json:"startTime"`
	EndTime   float64 `json:"endTime"`
}

type AniskipResult struct {
	Interval      AniskipInterval `json:"interval"`
	SkipType      string          `json:"skipType"`
	SkipID        string          `json:"skipId"`
	EpisodeLength float64         `json:"episodeLength"`
}

type AniskipResponse struct {
	Found   bool            `json:"found"`
	Results []AniskipResult `json:"results"`
}
