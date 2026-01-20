package tvdb

// TVDBAuthResponse represents the authentication response from TVDB
type TVDBAuthResponse struct {
	Status string `json:"status"`
	Data   struct {
		Token string `json:"token"`
	} `json:"data"`
}

// TVDBEpisode represents an episode from TVDB API v4
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

// TVDBEpisodesData represents the data container for episodes
type TVDBEpisodesData struct {
	Episodes []TVDBEpisode `json:"episodes"`
}

// TVDBEpisodesResponse represents the episodes response from TVDB API v4
type TVDBEpisodesResponse struct {
	Status string           `json:"status"`
	Data   TVDBEpisodesData `json:"data"`
}
