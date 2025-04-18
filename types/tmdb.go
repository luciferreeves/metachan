package types

// TMDBShowResult represents a TV show result from TMDB search
type TMDBShowResult struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	FirstAirDate  string   `json:"first_air_date"`
	OriginCountry []string `json:"origin_country"`
	Adult         bool     `json:"adult"`
}

// TMDBSearchResponse represents the response from TMDB search API
type TMDBSearchResponse struct {
	Page         int              `json:"page"`
	Results      []TMDBShowResult `json:"results"`
	TotalPages   int              `json:"total_pages"`
	TotalResults int              `json:"total_results"`
}

// TMDBEpisode represents a TV episode from TMDB
type TMDBEpisode struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	StillPath     string `json:"still_path"`
	AirDate       string `json:"air_date"`
	EpisodeNumber int    `json:"episode_number"`
	SeasonNumber  int    `json:"season_number"`
}

// TMDBSeasonDetails represents a TV season from TMDB
type TMDBSeasonDetails struct {
	ID           int           `json:"id"`
	AirDate      string        `json:"air_date"`
	EpisodeCount int           `json:"episode_count"`
	Name         string        `json:"name"`
	Overview     string        `json:"overview"`
	SeasonNumber int           `json:"season_number"`
	Episodes     []TMDBEpisode `json:"episodes"`
}

// TMDBShowDetails represents a TV show from TMDB
type TMDBShowDetails struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Overview string `json:"overview"`
	Seasons  []struct {
		ID           int    `json:"id"`
		Name         string `json:"name"`
		SeasonNumber int    `json:"season_number"`
		EpisodeCount int    `json:"episode_count"`
		AirDate      string `json:"air_date"`
	} `json:"seasons"`
}
