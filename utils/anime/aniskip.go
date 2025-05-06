package anime

import (
	"encoding/json"
	"fmt"
	"metachan/types"
	"net/http"
	"time"
)

type AniSkipInterval struct {
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

type AniSkipResult struct {
	Interval struct {
		StartTime float64 `json:"start_time"`
		EndTime   float64 `json:"end_time"`
	} `json:"interval"`
	SkipType      string  `json:"skip_type"`
	SkipID        string  `json:"skip_id"`
	EpisodeLength float64 `json:"episode_length"`
}

type AniSkipResponse struct {
	Found   bool            `json:"found"`
	Results []AniSkipResult `json:"results"`
}

func getAnimeEpisodeSkipTimes(malID int, episodeNumber int) ([]types.AnimeSkipTimes, error) {
	client := http.Client{
		Timeout: 10 * time.Second,
	}

	url := fmt.Sprintf("https://api.aniskip.com/v1/skip-times/%d/%d?types=op&types=ed", malID, episodeNumber)

	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to request skip times: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get skip times: status code %d", resp.StatusCode)
	}

	var skipResp AniSkipResponse
	if err := json.NewDecoder(resp.Body).Decode(&skipResp); err != nil {
		return nil, fmt.Errorf("failed to parse skip times: %w", err)
	}

	if !skipResp.Found || len(skipResp.Results) == 0 {
		return nil, nil
	}

	skipTimes := make([]types.AnimeSkipTimes, len(skipResp.Results))
	for i, result := range skipResp.Results {
		skipTimes[i] = types.AnimeSkipTimes{
			SkipType:      result.SkipType,
			StartTime:     result.Interval.StartTime,
			EndTime:       result.Interval.EndTime,
			EpisodeLength: result.EpisodeLength,
		}
	}

	return skipTimes, nil
}
