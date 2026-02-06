package types

// type EpisodeTitles struct {
// 	English  string `json:"english"`
// 	Japanese string `json:"japanese"`
// 	Romaji   string `json:"romaji"`
// }

// type AnimeStreamingSource struct {
// 	URL    string `json:"url"`
// 	Server string `json:"server"`
// 	Type   string `json:"type"`
// }

// type AnimeStreaming struct {
// 	Sub []AnimeStreamingSource `json:"sub"`
// 	Dub []AnimeStreamingSource `json:"dub"`
// }

// type AnimeSingleEpisode struct {
// 	ID           string          `json:"id"`
// 	Titles       EpisodeTitles   `json:"titles"`
// 	Description  string          `json:"description"`
// 	Aired        string          `json:"aired"`
// 	Score        float64         `json:"score"`
// 	Filler       bool            `json:"filler"`
// 	Recap        bool            `json:"recap"`
// 	ForumURL     string          `json:"forum_url"`
// 	URL          string          `json:"url"`
// 	ThumbnailURL string          `json:"thumbnail_url"`
// 	Streaming    *AnimeStreaming `json:"streaming,omitempty"`
// }

// // AnimeTitles contains different naming variants for an anime
// type AnimeTitles struct {
// 	English  string   `json:"english"`
// 	Japanese string   `json:"japanese"`
// 	Romaji   string   `json:"romaji"`
// 	Synonyms []string `json:"synonyms"`
// }

// // AnimeMappings contains IDs from various anime databases/services
// type AnimeMappings struct {
// 	AniDB          int    `json:"anidb,omitempty"`
// 	Anilist        int    `json:"anilist,omitempty"`
// 	AnimeCountdown int    `json:"anime_countdown,omitempty"`
// 	AnimePlanet    string `json:"anime_planet,omitempty"`
// 	AniSearch      int    `json:"ani_search,omitempty"`
// 	IMDB           string `json:"imdb,omitempty"`
// 	Kitsu          int    `json:"kitsu,omitempty"`
// 	LiveChart      int    `json:"live_chart,omitempty"`
// 	NotifyMoe      string `json:"notify_moe,omitempty"`
// 	Simkl          int    `json:"simkl,omitempty"`
// 	TMDB           int    `json:"tmdb,omitempty"`
// 	TVDB           int    `json:"tvdb,omitempty"`
// }

type AniSyncType string

const (
	AniSyncTypeTV      AniSyncType = "TV"
	AniSyncTypeMovie   AniSyncType = "MOVIE"
	AniSyncTypeOVA     AniSyncType = "OVA"
	AniSyncTypeONA     AniSyncType = "ONA"
	AniSyncTypeSpecial AniSyncType = "SPECIAL"
	AniSyncTypeMusic   AniSyncType = "MUSIC"
)

// //
// // Episode Related Types
// //

// // EpisodeTitles contains different naming variants for an episode
// type EpisodeTitles struct {
// 	English  string `json:"english"`
// 	Japanese string `json:"japanese"`
// 	Romaji   string `json:"romaji"`
// }

// // AnimeStreamingSource represents a single streaming source
// type AnimeStreamingSource struct {
// 	URL    string `json:"url"`
// 	Server string `json:"server"`
// 	Type   string `json:"type"` // direct or embed
// }

// // AnimeStreaming represents all available streaming sources
// type AnimeStreaming struct {
// 	Sub []AnimeStreamingSource `json:"sub"`
// 	Dub []AnimeStreamingSource `json:"dub"`
// }

// // AnimeSingleEpisode contains information about a single anime episode
// type AnimeSingleEpisode struct {
// 	ID           string          `json:"id"`
// 	Titles       EpisodeTitles   `json:"titles"`
// 	Description  string          `json:"description"`
// 	Aired        string          `json:"aired"`
// 	Score        float64         `json:"score"`
// 	Filler       bool            `json:"filler"`
// 	Recap        bool            `json:"recap"`
// 	ForumURL     string          `json:"forum_url"`
// 	URL          string          `json:"url"`
// 	ThumbnailURL string          `json:"thumbnail_url"`
// 	Streaming    *AnimeStreaming `json:"streaming,omitempty"`
// }

// // AnimeEpisodes contains information about all episodes of an anime
// type AnimeEpisodes struct {
// 	Total    int                  `json:"total"`
// 	Aired    int                  `json:"aired"`
// 	Subbed   int                  `json:"subbed"` // Count of subbed episodes
// 	Dubbed   int                  `json:"dubbed"` // Count of dubbed episodes
// 	Episodes []AnimeSingleEpisode `json:"episodes,omitempty"`
// }

// //
// // Visual Media Types
// //

// // AnimeLogos contains logo images in various sizes
// type AnimeLogos struct {
// 	Small    string `json:"small,omitempty"`
// 	Medium   string `json:"medium,omitempty"`
// 	Large    string `json:"large,omitempty"`
// 	XLarge   string `json:"xlarge,omitempty"`
// 	Original string `json:"original,omitempty"`
// }

// // AnimeImages contains general images in various sizes
// type AnimeImages struct {
// 	Small    string `json:"small,omitempty"`
// 	Large    string `json:"large,omitempty"`
// 	Original string `json:"original,omitempty"`
// }

// //
// // Production and Content Types
// //

// // AnimeGenres contains genre information
// type AnimeGenres struct {
// 	Name    string `json:"name"`
// 	GenreID int    `json:"genre_id"`
// 	URL     string `json:"url"`
// }

// // AnimeProducer contains full producer/studio/licensor information
// type AnimeProducer struct {
// 	MALID       int                   `json:"mal_id"`
// 	URL         string                `json:"url"`
// 	Name        string                `json:"name,omitempty"`
// 	Titles      []ProducerTitle       `json:"titles,omitempty"`
// 	Images      *ProducerImages       `json:"images,omitempty"`
// 	Favorites   int                   `json:"favorites,omitempty"`
// 	Count       int                   `json:"count,omitempty"`
// 	Established string                `json:"established,omitempty"`
// 	About       string                `json:"about,omitempty"`
// 	External    []ProducerExternalURL `json:"external,omitempty"`
// }

// // ProducerTitle represents a producer title variant
// type ProducerTitle struct {
// 	Type  string `json:"type"`
// 	Title string `json:"title"`
// }

// // ProducerImages represents producer image URLs
// type ProducerImages struct {
// 	JPG ProducerJPGImage `json:"jpg"`
// }

// // ProducerJPGImage represents JPG image variant
// type ProducerJPGImage struct {
// 	ImageURL string `json:"image_url"`
// }

// // ProducerExternalURL represents an external URL for a producer
// type ProducerExternalURL struct {
// 	Name string `json:"name"`
// 	URL  string `json:"url"`
// }

// //
// // Airing and Schedule Types
// //

// // AiringStatusDates contains date information for airing
// type AiringStatusDates struct {
// 	Day    int    `json:"day"`
// 	Month  int    `json:"month"`
// 	Year   int    `json:"year"`
// 	String string `json:"string"`
// }

// // AiringStatus contains full airing status information
// type AiringStatus struct {
// 	From   AiringStatusDates `json:"from"`
// 	To     AiringStatusDates `json:"to"`
// 	String string            `json:"string"`
// }

// // AnimeAiringEpisode contains information about a single upcoming episode
// type AnimeAiringEpisode struct {
// 	AiringAt int `json:"airing_at"`
// 	Episode  int `json:"episode"`
// }

// // AnimeBroadcast contains broadcast schedule information
// type AnimeBroadcast struct {
// 	Day      string `json:"day"`
// 	Time     string `json:"time"`
// 	Timezone string `json:"timezone"`
// 	String   string `json:"string"`
// }

// //
// // Community and Metrics Types
// //

// // AnimeScores contains popularity and rating information
// type AnimeScores struct {
// 	Score      float64 `json:"score"`
// 	ScoredBy   int     `json:"scored_by"`
// 	Rank       int     `json:"rank"`
// 	Popularity int     `json:"popularity"`
// 	Members    int     `json:"members"`
// 	Favorites  int     `json:"favorites"`
// }

// //
// // Character Related Types
// //

// // AnimeVoiceActor contains voice actor information
// type AnimeVoiceActor struct {
// 	MALID    int    `json:"mal_id"`
// 	URL      string `json:"url"`
// 	Image    string `json:"image_url"`
// 	Name     string `json:"name"`
// 	Language string `json:"language"`
// }

// // AnimeCharacter contains character information
// type AnimeCharacter struct {
// 	MALID       int               `json:"mal_id"`
// 	URL         string            `json:"url"`
// 	ImageURL    string            `json:"image_url"`
// 	Name        string            `json:"name"`
// 	Role        string            `json:"role"`
// 	VoiceActors []AnimeVoiceActor `json:"voice_actors"`
// }

// //
// // Season Related Types
// //

// // AnimeSeason contains information about a single anime season
// type AnimeSeason struct {
// 	MALID        int          `json:"id"`
// 	Titles       AnimeTitles  `json:"titles"`
// 	Synopsis     string       `json:"synopsis"`
// 	Type         AniSyncType  `json:"type"`
// 	Source       string       `json:"source"`
// 	Airing       bool         `json:"airing"`
// 	Status       string       `json:"status"`
// 	AiringStatus AiringStatus `json:"airing_status"`
// 	Duration     string       `json:"duration"`
// 	Images       AnimeImages  `json:"images"`
// 	Scores       AnimeScores  `json:"scores"`
// 	Season       string       `json:"season"`
// 	Year         int          `json:"year"`
// 	Current      bool         `json:"current"`
// }

// //
// // Main Anime Type
// //

// // Anime is the main structure containing all anime information
// type Anime struct {
// 	MALID             int                  `json:"id"`
// 	Titles            AnimeTitles          `json:"titles"`
// 	Synopsis          string               `json:"synopsis"`
// 	Type              AniSyncType          `json:"type"`
// 	Source            string               `json:"source"`
// 	Airing            bool                 `json:"airing"`
// 	Status            string               `json:"status"`
// 	AiringStatus      AiringStatus         `json:"airing_status"`
// 	Duration          string               `json:"duration,omitempty"`
// 	Images            AnimeImages          `json:"images"`
// 	Logos             AnimeLogos           `json:"logos,omitempty"`
// 	Covers            AnimeImages          `json:"covers,omitempty"`
// 	Color             string               `json:"color,omitempty"`
// 	Genres            []AnimeGenres        `json:"genres"`
// 	Scores            AnimeScores          `json:"scores"`
// 	Season            string               `json:"season"`
// 	Year              int                  `json:"year"`
// 	Broadcast         AnimeBroadcast       `json:"broadcast"`
// 	Producers         []AnimeProducer      `json:"producers"`
// 	Studios           []AnimeProducer      `json:"studios"`
// 	Licensors         []AnimeProducer      `json:"licensors"`
// 	Seasons           []AnimeSeason        `json:"seasons,omitempty"`
// 	Episodes          AnimeEpisodes        `json:"episodes"`
// 	NextAiringEpisode AnimeAiringEpisode   `json:"next_airing_episode,omitempty"`
// 	AiringSchedule    []AnimeAiringEpisode `json:"airing_schedule,omitempty"`
// 	Characters        []AnimeCharacter     `json:"characters,omitempty"`
// 	Mappings          AnimeMappings        `json:"mappings,omitempty"`
// }
