package types

type TMDBShowResult struct {
	ID            int      `json:"id"`
	Name          string   `json:"name"`
	FirstAirDate  string   `json:"first_air_date"`
	OriginCountry []string `json:"origin_country"`
	Adult         bool     `json:"adult"`
}

type TMDBSearchResponse struct {
	Page         int              `json:"page"`
	Results      []TMDBShowResult `json:"results"`
	TotalPages   int              `json:"total_pages"`
	TotalResults int              `json:"total_results"`
}

type TMDBEpisode struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	StillPath     string `json:"still_path"`
	AirDate       string `json:"air_date"`
	EpisodeNumber int    `json:"episode_number"`
	SeasonNumber  int    `json:"season_number"`
}

type TMDBSeasonDetails struct {
	ID           int           `json:"id"`
	AirDate      string        `json:"air_date"`
	EpisodeCount int           `json:"episode_count"`
	Name         string        `json:"name"`
	Overview     string        `json:"overview"`
	SeasonNumber int           `json:"season_number"`
	Episodes     []TMDBEpisode `json:"episodes"`
}

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

type TMDBMovieResult struct {
	ID          int    `json:"id"`
	Adult       bool   `json:"adult"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date"`
}

type TMDBMovieSearchResponse struct {
	Page         int               `json:"page"`
	Results      []TMDBMovieResult `json:"results"`
	TotalPages   int               `json:"total_pages"`
	TotalResults int               `json:"total_results"`
}

type TMDBMovieDetails struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Overview     string `json:"overview"`
	BackdropPath string `json:"backdrop_path"`
	PosterPath   string `json:"poster_path"`
	ReleaseDate  string `json:"release_date"`
	Runtime      int    `json:"runtime"`
}
