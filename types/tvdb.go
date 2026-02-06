package types

type TMDBAuthData struct {
	Token string `json:"token"`
}

type TVDBAuthResponse struct {
	Status string       `json:"status"`
	Data   TMDBAuthData `json:"data"`
}

type TVDBEpisode struct {
	ID                   int      `json:"id"`
	SeriesID             int      `json:"seriesId"`
	Name                 string   `json:"name"`
	Aired                string   `json:"aired"`
	Runtime              int      `json:"runtime"`
	NameTranslations     []string `json:"nameTranslations"`
	Overview             string   `json:"overview"`
	OverviewTranslations []string `json:"overviewTranslations"`
	Image                string   `json:"image"`
	ImageType            int      `json:"imageType"`
	IsMovie              int      `json:"isMovie"`
	Number               int      `json:"number"`
	AbsoluteNumber       int      `json:"absoluteNumber"`
	SeasonNumber         int      `json:"seasonNumber"`
	LastUpdated          string   `json:"lastUpdated"`
	FinaleType           *string  `json:"finaleType"`
	AirsBeforeSeason     int      `json:"airsBeforeSeason"`
	AirsBeforeEpisode    int      `json:"airsBeforeEpisode"`
	Year                 string   `json:"year"`
}

type TVDBEpisodesData struct {
	Episodes []TVDBEpisode `json:"episodes"`
}

type TVDBEpisodesResponse struct {
	Status string           `json:"status"`
	Data   TVDBEpisodesData `json:"data"`
}
