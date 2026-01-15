package types

// AnimeTitles contains different naming variants for an anime
type AnimeTitles struct {
	English  string   `json:"english"`
	Japanese string   `json:"japanese"`
	Romaji   string   `json:"romaji"`
	Synonyms []string `json:"synonyms"`
}

// AnimeMappings contains IDs from various anime databases/services
type AnimeMappings struct {
	AniDB          int    `json:"anidb"`
	Anilist        int    `json:"anilist"`
	AnimeCountdown int    `json:"anime_countdown"`
	AnimePlanet    string `json:"anime_planet"`
	AniSearch      int    `json:"ani_search"`
	IMDB           string `json:"imdb"`
	Kitsu          int    `json:"kitsu"`
	LiveChart      int    `json:"live_chart"`
	NotifyMoe      string `json:"notify_moe"`
	Simkl          int    `json:"simkl"`
	TMDB           int    `json:"tmdb"`
	TVDB           int    `json:"tvdb"`
}

//
// Episode Related Types
//

// EpisodeTitles contains different naming variants for an episode
type EpisodeTitles struct {
	English  string `json:"english"`
	Japanese string `json:"japanese"`
	Romaji   string `json:"romaji"`
}

// AnimeStreamingSource represents a single streaming source
type AnimeStreamingSource struct {
	URL    string `json:"url"`
	Server string `json:"server"`
	Type   string `json:"type"` // direct or embed
}

// AnimeStreaming represents all available streaming sources
type AnimeStreaming struct {
	Sub []AnimeStreamingSource `json:"sub"`
	Dub []AnimeStreamingSource `json:"dub"`
}

// AnimeSingleEpisode contains information about a single anime episode
type AnimeSingleEpisode struct {
	ID           string          `json:"id"`
	Titles       EpisodeTitles   `json:"titles"`
	Description  string          `json:"description"`
	Aired        string          `json:"aired"`
	Score        float64         `json:"score"`
	Filler       bool            `json:"filler"`
	Recap        bool            `json:"recap"`
	ForumURL     string          `json:"forum_url"`
	URL          string          `json:"url"`
	ThumbnailURL string          `json:"thumbnail_url"`
	Streaming    *AnimeStreaming `json:"streaming,omitempty"`
}

// AnimeEpisodes contains information about all episodes of an anime
type AnimeEpisodes struct {
	Total    int                  `json:"total"`
	Aired    int                  `json:"aired"`
	Subbed   int                  `json:"subbed"` // Count of subbed episodes
	Dubbed   int                  `json:"dubbed"` // Count of dubbed episodes
	Episodes []AnimeSingleEpisode `json:"episodes"`
}

//
// Visual Media Types
//

// AnimeLogos contains logo images in various sizes
type AnimeLogos struct {
	Small    string `json:"small,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Large    string `json:"large,omitempty"`
	XLarge   string `json:"xlarge,omitempty"`
	Original string `json:"original,omitempty"`
}

// AnimeImages contains general images in various sizes
type AnimeImages struct {
	Small    string `json:"small,omitempty"`
	Large    string `json:"large,omitempty"`
	Original string `json:"original,omitempty"`
}

//
// Production and Content Types
//

// AnimeGenres contains genre information
type AnimeGenres struct {
	Name    string `json:"name"`
	GenreID int    `json:"genre_id"`
	URL     string `json:"url"`
}

// AnimeProducer contains producer information
type AnimeProducer struct {
	Name       string `json:"name"`
	ProducerID int    `json:"producer_id"`
	URL        string `json:"url"`
}

// AnimeLicensor contains licensor information
type AnimeLicensor struct {
	Name       string `json:"name"`
	ProducerID int    `json:"producer_id"`
	URL        string `json:"url"`
}

// AnimeStudio contains studio information
type AnimeStudio struct {
	Name     string `json:"name"`
	StudioID int    `json:"studio_id"`
	URL      string `json:"url"`
}

//
// Airing and Schedule Types
//

// AiringStatusDates contains date information for airing
type AiringStatusDates struct {
	Day    int    `json:"day"`
	Month  int    `json:"month"`
	Year   int    `json:"year"`
	String string `json:"string"`
}

// AiringStatus contains full airing status information
type AiringStatus struct {
	From   AiringStatusDates `json:"from"`
	To     AiringStatusDates `json:"to"`
	String string            `json:"string"`
}

// AnimeAiringEpisode contains information about a single upcoming episode
type AnimeAiringEpisode struct {
	AiringAt int `json:"airing_at"`
	Episode  int `json:"episode"`
}

// AnimeBroadcast contains broadcast schedule information
type AnimeBroadcast struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	Timezone string `json:"timezone"`
	String   string `json:"string"`
}

//
// Community and Metrics Types
//

// AnimeScores contains popularity and rating information
type AnimeScores struct {
	Score      float64 `json:"score"`
	ScoredBy   int     `json:"scored_by"`
	Rank       int     `json:"rank"`
	Popularity int     `json:"popularity"`
	Members    int     `json:"members"`
	Favorites  int     `json:"favorites"`
}

//
// Character Related Types
//

// AnimeVoiceActor contains voice actor information
type AnimeVoiceActor struct {
	MALID    int    `json:"mal_id"`
	URL      string `json:"url"`
	Image    string `json:"image_url"`
	Name     string `json:"name"`
	Language string `json:"language"`
}

// AnimeCharacter contains character information
type AnimeCharacter struct {
	MALID       int               `json:"mal_id"`
	URL         string            `json:"url"`
	ImageURL    string            `json:"image_url"`
	Name        string            `json:"name"`
	Role        string            `json:"role"`
	VoiceActors []AnimeVoiceActor `json:"voice_actors"`
}

//
// Season Related Types
//

// AnimeSeason contains information about a single anime season
type AnimeSeason struct {
	MALID        int          `json:"id"`
	Titles       AnimeTitles  `json:"titles"`
	Synopsis     string       `json:"synopsis"`
	Type         AniSyncType  `json:"type"`
	Source       string       `json:"source"`
	Airing       bool         `json:"airing"`
	Status       string       `json:"status"`
	AiringStatus AiringStatus `json:"airing_status"`
	Duration     string       `json:"duration"`
	Images       AnimeImages  `json:"images"`
	Scores       AnimeScores  `json:"scores"`
	Season       string       `json:"season"`
	Year         int          `json:"year"`
	Current      bool         `json:"current"`
}

//
// Main Anime Type
//

// Anime is the main structure containing all anime information
type Anime struct {
	MALID             int                  `json:"id"`
	Titles            AnimeTitles          `json:"titles"`
	Synopsis          string               `json:"synopsis"`
	Type              AniSyncType          `json:"type"`
	Source            string               `json:"source"`
	Airing            bool                 `json:"airing"`
	Status            string               `json:"status"`
	AiringStatus      AiringStatus         `json:"airing_status"`
	Duration          string               `json:"duration"`
	Images            AnimeImages          `json:"images"`
	Logos             AnimeLogos           `json:"logos"`
	Covers            AnimeImages          `json:"covers"`
	Color             string               `json:"color"`
	Genres            []AnimeGenres        `json:"genres"`
	Scores            AnimeScores          `json:"scores"`
	Season            string               `json:"season"`
	Year              int                  `json:"year"`
	Broadcast         AnimeBroadcast       `json:"broadcast"`
	Producers         []AnimeProducer      `json:"producers"`
	Studios           []AnimeStudio        `json:"studios"`
	Licensors         []AnimeLicensor      `json:"licensors"`
	Seasons           []AnimeSeason        `json:"seasons"`
	Episodes          AnimeEpisodes        `json:"episodes"`
	NextAiringEpisode AnimeAiringEpisode   `json:"next_airing_episode"`
	AiringSchedule    []AnimeAiringEpisode `json:"airing_schedule"`
	Characters        []AnimeCharacter     `json:"characters"`
	Mappings          AnimeMappings        `json:"mappings"`
}
