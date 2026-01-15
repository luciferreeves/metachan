package entities

import (
	"time"

	"gorm.io/gorm"
)

type MappingType string

const (
	m_SPECIAL MappingType = "SPECIAL"
	m_TV      MappingType = "TV"
	m_OVA     MappingType = "OVA"
	m_MOVIE   MappingType = "MOVIE"
	m_ONA     MappingType = "ONA"
	m_UNKNOWN MappingType = "UNKNOWN"
)

type AnimeMapping struct {
	gorm.Model
	AniDB               int
	Anilist             int
	AnimeCountdown      int
	AnimePlanet         string
	AniSearch           int
	IMDB                string
	Kitsu               int
	LiveChart           int
	MAL                 int
	NotifyMoe           string
	Simkl               int
	TMDB                int
	TVDB                int
	Type                MappingType
	MALAnilistComposite *string `gorm:"uniqueIndex"`
}

type Anime struct {
	gorm.Model
	MALID         int `gorm:"uniqueIndex"`
	TitleRomaji   string
	TitleEnglish  string
	TitleJapanese string
	TitleSynonyms string
	Synopsis      string `gorm:"type:text"`
	Type          string
	Source        string
	Airing        bool
	Status        string
	Duration      string
	Color         string
	Season        string
	Year          int
	SubbedCount   int
	DubbedCount   int
	TotalEpisodes int
	AiredEpisodes int
	LastUpdated   time.Time

	Images            *AnimeImages    `gorm:"foreignKey:AnimeID"`
	Logos             *AnimeLogos     `gorm:"foreignKey:AnimeID"`
	Covers            *AnimeCovers    `gorm:"foreignKey:AnimeID"`
	Scores            *AnimeScores    `gorm:"foreignKey:AnimeID"`
	AiringStatus      *AiringStatus   `gorm:"foreignKey:AnimeID"`
	Broadcast         *AnimeBroadcast `gorm:"foreignKey:AnimeID"`
	NextAiringEpisode *NextEpisode    `gorm:"foreignKey:AnimeID"`

	Genres         []AnimeGenre         `gorm:"foreignKey:AnimeID"`
	Producers      []AnimeProducer      `gorm:"foreignKey:AnimeID"`
	Studios        []AnimeStudio        `gorm:"foreignKey:AnimeID"`
	Licensors      []AnimeLicensor      `gorm:"foreignKey:AnimeID"`
	Episodes       []AnimeSingleEpisode `gorm:"foreignKey:AnimeID"`
	Characters     []AnimeCharacter     `gorm:"foreignKey:AnimeID"`
	AiringSchedule []ScheduleEpisode    `gorm:"foreignKey:AnimeID;constraint:OnDelete:CASCADE"`
	Seasons        []AnimeSeason        `gorm:"foreignKey:ParentAnimeID"`
}

type AnimeImages struct {
	gorm.Model
	AnimeID  uint
	Small    string
	Large    string
	Original string
}

type AnimeCovers struct {
	gorm.Model
	AnimeID  uint
	Small    string
	Large    string
	Original string
}

type AnimeLogos struct {
	gorm.Model
	AnimeID  uint
	Small    string
	Medium   string
	Large    string
	XLarge   string
	Original string
}

type AnimeScores struct {
	gorm.Model
	AnimeID    uint
	Score      float64
	ScoredBy   int
	Rank       int
	Popularity int
	Members    int
	Favorites  int
}

type AiringStatusDates struct {
	gorm.Model
	Day    int
	Month  int
	Year   int
	String string
}

type AiringStatus struct {
	gorm.Model
	AnimeID uint
	FromID  *uint
	ToID    *uint
	String  string

	From *AiringStatusDates `gorm:"foreignKey:FromID"`
	To   *AiringStatusDates `gorm:"foreignKey:ToID"`
}

type AnimeBroadcast struct {
	gorm.Model
	AnimeID  uint
	Day      string
	Time     string
	Timezone string
	String   string
}

type AnimeGenre struct {
	gorm.Model
	AnimeID uint
	Name    string
	GenreID int
	URL     string
}

type AnimeProducer struct {
	gorm.Model
	AnimeID    uint
	Name       string
	ProducerID int
	URL        string
}

type AnimeStudio struct {
	gorm.Model
	AnimeID  uint
	Name     string
	StudioID int
	URL      string
}

type AnimeLicensor struct {
	gorm.Model
	AnimeID    uint
	Name       string
	ProducerID int
	URL        string
}

type ScheduleEpisode struct {
	gorm.Model
	AnimeID  uint
	AiringAt int
	Episode  int
	IsNext   bool `gorm:"index"`
}

type EpisodeTitles struct {
	gorm.Model
	EpisodeID uint
	English   string
	Japanese  string
	Romaji    string
}

type AnimeSingleEpisode struct {
	gorm.Model
	EpisodeID    string `gorm:"uniqueIndex;size:32"`
	AnimeID      uint
	TitlesID     uint
	Description  string `gorm:"type:text"`
	Aired        string
	Score        float64
	Filler       bool
	Recap        bool
	ForumURL     string
	URL          string
	ThumbnailURL string

	Titles *EpisodeTitles `gorm:"foreignKey:TitlesID"`
}

type AnimeSeason struct {
	gorm.Model
	ParentAnimeID uint
	MALID         int
	TitleRomaji   string
	TitleEnglish  string
	TitleJapanese string
	TitleSynonyms string
	Synopsis      string `gorm:"type:text"`
	Type          string
	Source        string
	Airing        bool
	Status        string
	Duration      string
	Season        string
	Year          int
	Current       bool

	Images       *AnimeImages  `gorm:"foreignKey:AnimeID"`
	Scores       *AnimeScores  `gorm:"foreignKey:AnimeID"`
	AiringStatus *AiringStatus `gorm:"foreignKey:AnimeID"`
}

type AnimeVoiceActor struct {
	gorm.Model
	CharacterID uint
	MALID       int
	URL         string
	Image       string
	Name        string
	Language    string
}

type AnimeCharacter struct {
	gorm.Model
	AnimeID  uint
	MALID    int
	URL      string
	ImageURL string
	Name     string
	Role     string

	VoiceActors []AnimeVoiceActor `gorm:"foreignKey:CharacterID"`
}

type NextEpisode struct {
	gorm.Model
	AnimeID  uint
	AiringAt int
	Episode  int
}

// CachedAnime represents the main cached anime entity in database
type CachedAnime struct {
	gorm.Model
	MALID         int `gorm:"uniqueIndex"`
	TitleRomaji   string
	TitleEnglish  string
	TitleJapanese string
	TitleSynonyms string // Comma-separated list of synonyms
	Synopsis      string `gorm:"type:text"`
	Type          string
	Source        string
	Airing        bool
	Status        string
	Duration      string
	Color         string
	Season        string
	Year          int
	SubbedCount   int
	DubbedCount   int
	TotalEpisodes int
	AiredEpisodes int
	LastUpdated   time.Time

	// Relationships
	Images            *CachedAnimeImages    `gorm:"foreignKey:AnimeID"`
	Logos             *CachedAnimeLogos     `gorm:"foreignKey:AnimeID"`
	Covers            *CachedAnimeCovers    `gorm:"foreignKey:AnimeID"`
	Scores            *CachedAnimeScores    `gorm:"foreignKey:AnimeID"`
	AiringStatus      *CachedAiringStatus   `gorm:"foreignKey:AnimeID"`
	Broadcast         *CachedAnimeBroadcast `gorm:"foreignKey:AnimeID"`
	NextAiringEpisode *CachedNextEpisode    `gorm:"foreignKey:AnimeID"`

	// One-to-many relationships
	Genres         []CachedAnimeGenre         `gorm:"foreignKey:AnimeID"`
	Producers      []CachedAnimeProducer      `gorm:"foreignKey:AnimeID"`
	Studios        []CachedAnimeStudio        `gorm:"foreignKey:AnimeID"`
	Licensors      []CachedAnimeLicensor      `gorm:"foreignKey:AnimeID"`
	Episodes       []CachedAnimeSingleEpisode `gorm:"foreignKey:AnimeID"`
	Characters     []CachedAnimeCharacter     `gorm:"foreignKey:AnimeID"`
	AiringSchedule []CachedScheduleEpisode    `gorm:"foreignKey:AnimeID;constraint:OnDelete:CASCADE"`
	Seasons        []CachedAnimeSeason        `gorm:"foreignKey:ParentAnimeID"`
}

// CachedAnimeImages for storing anime images
type CachedAnimeImages struct {
	gorm.Model
	AnimeID  uint
	Small    string
	Large    string
	Original string
}

// CachedAnimeCovers for storing anime cover images
type CachedAnimeCovers struct {
	gorm.Model
	AnimeID  uint
	Small    string
	Large    string
	Original string
}

// CachedAnimeLogos for storing anime logos
type CachedAnimeLogos struct {
	gorm.Model
	AnimeID  uint
	Small    string
	Medium   string
	Large    string
	XLarge   string
	Original string
}

// CachedAnimeScores for storing anime scores and popularity data
type CachedAnimeScores struct {
	gorm.Model
	AnimeID    uint
	Score      float64
	ScoredBy   int
	Rank       int
	Popularity int
	Members    int
	Favorites  int
}

// CachedAiringStatusDates for storing airing date information
type CachedAiringStatusDates struct {
	gorm.Model
	Day    int
	Month  int
	Year   int
	String string
}

// CachedAiringStatus for storing anime airing status
type CachedAiringStatus struct {
	gorm.Model
	AnimeID uint
	FromID  *uint
	ToID    *uint
	String  string

	From *CachedAiringStatusDates `gorm:"foreignKey:FromID"`
	To   *CachedAiringStatusDates `gorm:"foreignKey:ToID"`
}

// CachedAnimeBroadcast for storing broadcast information
type CachedAnimeBroadcast struct {
	gorm.Model
	AnimeID  uint
	Day      string
	Time     string
	Timezone string
	String   string
}

// CachedAnimeGenre for storing anime genres
type CachedAnimeGenre struct {
	gorm.Model
	AnimeID uint
	Name    string
	GenreID int
	URL     string
}

// CachedAnimeProducer for storing anime producers
type CachedAnimeProducer struct {
	gorm.Model
	AnimeID    uint
	Name       string
	ProducerID int
	URL        string
}

// CachedAnimeStudio for storing anime studios
type CachedAnimeStudio struct {
	gorm.Model
	AnimeID  uint
	Name     string
	StudioID int
	URL      string
}

// CachedAnimeLicensor for storing anime licensors
type CachedAnimeLicensor struct {
	gorm.Model
	AnimeID    uint
	Name       string
	ProducerID int
	URL        string
}

// CachedAiringEpisode for storing information about airing episodes
type CachedAiringEpisode struct {
	gorm.Model
	AnimeID  uint
	AiringAt int
	Episode  int
	IsNext   bool `gorm:"index"` // true if this is the next airing episode
}

// CachedEpisodeTitles for episode title variants
type CachedEpisodeTitles struct {
	gorm.Model
	EpisodeID uint
	English   string
	Japanese  string
	Romaji    string
}

// CachedAnimeSingleEpisode for storing individual episode details
type CachedAnimeSingleEpisode struct {
	gorm.Model
	EpisodeID    string `gorm:"uniqueIndex;size:32"`
	AnimeID      uint
	TitlesID     uint
	Description  string `gorm:"type:text"`
	Aired        string
	Score        float64
	Filler       bool
	Recap        bool
	ForumURL     string
	URL          string
	ThumbnailURL string

	Titles *CachedEpisodeTitles `gorm:"foreignKey:TitlesID"`
}

// CachedAnimeSeason for storing anime seasons
type CachedAnimeSeason struct {
	gorm.Model
	ParentAnimeID uint `gorm:"index"`
	MALID         int  `gorm:"index"`
	TitleRomaji   string
	TitleEnglish  string
	TitleJapanese string
	TitleSynonyms string // Comma-separated
	Synopsis      string `gorm:"type:text"`
	Type          string
	Source        string
	Airing        bool
	Status        string
	Duration      string
	Season        string
	Year          int
	Current       bool

	// Relationships - fixing the foreign key references
	ImagesID       *uint
	ScoresID       *uint
	AiringStatusID *uint

	// Define proper relationships
	Images       *CachedAnimeImages  `gorm:"foreignKey:ImagesID"`
	Scores       *CachedAnimeScores  `gorm:"foreignKey:ScoresID"`
	AiringStatus *CachedAiringStatus `gorm:"foreignKey:AiringStatusID"`
}

// CachedAnimeCharacter for storing character information
type CachedAnimeCharacter struct {
	gorm.Model
	AnimeID  uint
	MALID    int
	URL      string
	ImageURL string
	Name     string
	Role     string

	// Voice actors
	VoiceActors []CachedAnimeVoiceActor `gorm:"foreignKey:CharacterID"`
}

// CachedAnimeVoiceActor for storing voice actor information
type CachedAnimeVoiceActor struct {
	gorm.Model
	CharacterID uint
	MALID       int
	URL         string
	Image       string
	Name        string
	Language    string
}

// CachedNextEpisode for storing the next airing episode information
type CachedNextEpisode struct {
	gorm.Model
	AnimeID  uint
	AiringAt int
	Episode  int
}

// CachedScheduleEpisode for storing information about scheduled episodes
type CachedScheduleEpisode struct {
	gorm.Model
	AnimeID  uint
	AiringAt int
	Episode  int
}

// EpisodeStreamingSource stores individual streaming sources for episodes
type EpisodeStreamingSource struct {
	gorm.Model
	EpisodeStreamingID uint
	URL                string
	Server             string
	Type               string // M3U8, MP4, or embed
}

// EpisodeStreaming stores streaming data for a specific episode
type EpisodeStreaming struct {
	gorm.Model
	EpisodeID  string                   `gorm:"uniqueIndex:idx_episode_streaming;size:32"`
	AnimeID    uint                     `gorm:"uniqueIndex:idx_episode_streaming"`
	SubSources []EpisodeStreamingSource `gorm:"foreignKey:EpisodeStreamingID;constraint:OnDelete:CASCADE"`
	DubSources []EpisodeStreamingSource `gorm:"foreignKey:EpisodeStreamingID;constraint:OnDelete:CASCADE"`
	LastFetch  time.Time
}
