package types

type AnilistTitle struct {
	Romaji        string `json:"romaji"`
	English       string `json:"english"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

type AnilistDate struct {
	Year  int `json:"year"`
	Month int `json:"month"`
	Day   int `json:"day"`
}

type AnilistTrailer struct {
	ID        string `json:"id"`
	Site      string `json:"site"`
	Thumbnail string `json:"thumbnail"`
}

type AnilistCoverImage struct {
	ExtraLarge string `json:"extraLarge"`
	Large      string `json:"large"`
	Medium     string `json:"medium"`
	Color      string `json:"color"`
}

type AnilistImage struct {
	Large  string `json:"large"`
	Medium string `json:"medium"`
}

type AnilistName struct {
	First         string `json:"first"`
	Last          string `json:"last"`
	Middle        string `json:"middle"`
	Full          string `json:"full"`
	Native        string `json:"native"`
	UserPreferred string `json:"userPreferred"`
}

type AnilistTag struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	Category         string `json:"category"`
	Rank             int    `json:"rank"`
	IsGeneralSpoiler bool   `json:"isGeneralSpoiler"`
	IsMediaSpoiler   bool   `json:"isMediaSpoiler"`
	IsAdult          bool   `json:"isAdult"`
}

type AnilistRelationNode struct {
	ID          int               `json:"id"`
	Title       AnilistTitle      `json:"title"`
	Format      string            `json:"format"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	CoverImage  AnilistCoverImage `json:"coverImage"`
	BannerImage string            `json:"bannerImage"`
}

type AnilistRelationEdge struct {
	ID           int                 `json:"id"`
	RelationType string              `json:"relationType"`
	Node         AnilistRelationNode `json:"node"`
}

type AnilistRelations struct {
	Edges []AnilistRelationEdge `json:"edges"`
}

type AnilistCharacterNode struct {
	ID          int          `json:"id"`
	Name        AnilistName  `json:"name"`
	Image       AnilistImage `json:"image"`
	Description string       `json:"description"`
	Age         string       `json:"age"`
}

type AnilistCharacterEdge struct {
	Role string               `json:"role"`
	Node AnilistCharacterNode `json:"node"`
}

type AnilistCharacters struct {
	Edges []AnilistCharacterEdge `json:"edges"`
}

type AnilistStaffNode struct {
	ID                 int          `json:"id"`
	Name               AnilistName  `json:"name"`
	Image              AnilistImage `json:"image"`
	Description        string       `json:"description"`
	PrimaryOccupations []string     `json:"primaryOccupations"`
	Gender             string       `json:"gender"`
	Age                int          `json:"age"`
	LanguageV2         string       `json:"languageV2"`
}

type AnilistStaffEdge struct {
	Role string           `json:"role"`
	Node AnilistStaffNode `json:"node"`
}

type AnilistStaff struct {
	Edges []AnilistStaffEdge `json:"edges"`
}

type AnilistStudioNode struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type AnilistStudioEdge struct {
	IsMain bool              `json:"isMain"`
	Node   AnilistStudioNode `json:"node"`
}

type AnilistStudios struct {
	Edges []AnilistStudioEdge `json:"edges"`
}

type AnilistNextAiringEpisode struct {
	ID              int `json:"id"`
	AiringAt        int `json:"airingAt"`
	TimeUntilAiring int `json:"timeUntilAiring"`
	Episode         int `json:"episode"`
}

type AnilistScheduleNode struct {
	ID              int `json:"id"`
	Episode         int `json:"episode"`
	AiringAt        int `json:"airingAt"`
	TimeUntilAiring int `json:"timeUntilAiring"`
}

type AnilistAiringSchedule struct {
	Nodes []AnilistScheduleNode `json:"nodes"`
}

type AnilistTrendNode struct {
	Date       int `json:"date"`
	Trending   int `json:"trending"`
	Popularity int `json:"popularity"`
	InProgress int `json:"inProgress"`
}

type AnilistTrends struct {
	Nodes []AnilistTrendNode `json:"nodes"`
}

type AnilistExternalLink struct {
	ID   int    `json:"id"`
	URL  string `json:"url"`
	Site string `json:"site"`
}

type AnilistStreamingEpisode struct {
	Title     string `json:"title"`
	Thumbnail string `json:"thumbnail"`
	URL       string `json:"url"`
	Site      string `json:"site"`
}

type AnilistRanking struct {
	ID      int    `json:"id"`
	Rank    int    `json:"rank"`
	Type    string `json:"type"`
	Format  string `json:"format"`
	Year    int    `json:"year"`
	Season  string `json:"season"`
	AllTime bool   `json:"allTime"`
	Context string `json:"context"`
}

type AnilistScoreDistribution struct {
	Score  int `json:"score"`
	Amount int `json:"amount"`
}

type AnilistStatusDistribution struct {
	Status string `json:"status"`
	Amount int    `json:"amount"`
}

type AnilistStats struct {
	ScoreDistribution  []AnilistScoreDistribution  `json:"scoreDistribution"`
	StatusDistribution []AnilistStatusDistribution `json:"statusDistribution"`
}

type AnilistMedia struct {
	ID                int                       `json:"id"`
	MALID             int                       `json:"idMal"`
	Title             AnilistTitle              `json:"title"`
	Type              string                    `json:"type"`
	Format            string                    `json:"format"`
	Status            string                    `json:"status"`
	Description       string                    `json:"description"`
	StartDate         AnilistDate               `json:"startDate"`
	EndDate           AnilistDate               `json:"endDate"`
	Season            string                    `json:"season"`
	SeasonYear        int                       `json:"seasonYear"`
	Episodes          int                       `json:"episodes"`
	Duration          int                       `json:"duration"`
	Chapters          int                       `json:"chapters"`
	Volumes           int                       `json:"volumes"`
	CountryOfOrigin   string                    `json:"countryOfOrigin"`
	IsLicensed        bool                      `json:"isLicensed"`
	Source            string                    `json:"source"`
	Hashtag           string                    `json:"hashtag"`
	Trailer           AnilistTrailer            `json:"trailer"`
	CoverImage        AnilistCoverImage         `json:"coverImage"`
	BannerImage       string                    `json:"bannerImage"`
	Genres            []string                  `json:"genres"`
	Synonyms          []string                  `json:"synonyms"`
	AverageScore      int                       `json:"averageScore"`
	MeanScore         int                       `json:"meanScore"`
	Popularity        int                       `json:"popularity"`
	IsLocked          bool                      `json:"isLocked"`
	Trending          int                       `json:"trending"`
	Favorites         int                       `json:"favorites"`
	Tags              []AnilistTag              `json:"tags"`
	Relations         AnilistRelations          `json:"relations"`
	Characters        AnilistCharacters         `json:"characters"`
	Staff             AnilistStaff              `json:"staff"`
	Studios           AnilistStudios            `json:"studios"`
	IsAdult           bool                      `json:"isAdult"`
	NextAiringEpisode AnilistNextAiringEpisode  `json:"nextAiringEpisode"`
	AiringSchedule    AnilistAiringSchedule     `json:"airingSchedule"`
	Trends            AnilistTrends             `json:"trends"`
	ExternalLinks     []AnilistExternalLink     `json:"externalLinks"`
	StreamingEpisodes []AnilistStreamingEpisode `json:"streamingEpisodes"`
	Rankings          []AnilistRanking          `json:"rankings"`
	Stats             AnilistStats              `json:"stats"`
	SiteURL           string                    `json:"siteUrl"`
}

type AnilistAnimeData struct {
	Media AnilistMedia `json:"media"`
}

type AnilistAnimeResponse struct {
	Data AnilistAnimeData `json:"data"`
}
