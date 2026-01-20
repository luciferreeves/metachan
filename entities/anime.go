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
	Count   int `gorm:"default:0"` // Total count from MAL (only set when AnimeID=0 for master genres)
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
