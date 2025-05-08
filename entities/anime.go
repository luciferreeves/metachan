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
	LastUpdated   time.Time

	// Relationships
	Images            *CachedAnimeImages    `gorm:"foreignKey:AnimeID"`
	Logos             *CachedAnimeLogos     `gorm:"foreignKey:AnimeID"`
	Covers            *CachedAnimeCovers    `gorm:"foreignKey:AnimeID"`
	Scores            *CachedAnimeScores    `gorm:"foreignKey:AnimeID"`
	AiringStatus      *CachedAiringStatus   `gorm:"foreignKey:AnimeID"`
	Broadcast         *CachedAnimeBroadcast `gorm:"foreignKey:AnimeID"`
	NextAiringEpisode *CachedAiringEpisode  `gorm:"foreignKey:AnimeID"`

	// One-to-many relationships
	Genres         []CachedAnimeGenre         `gorm:"foreignKey:AnimeID"`
	Producers      []CachedAnimeProducer      `gorm:"foreignKey:AnimeID"`
	Studios        []CachedAnimeStudio        `gorm:"foreignKey:AnimeID"`
	Licensors      []CachedAnimeLicensor      `gorm:"foreignKey:AnimeID"`
	Episodes       []CachedAnimeSingleEpisode `gorm:"foreignKey:AnimeID"`
	Characters     []CachedAnimeCharacter     `gorm:"foreignKey:AnimeID"`
	AiringSchedule []CachedAiringEpisode      `gorm:"foreignKey:AnimeID;references:ID;foreignKey:AnimeID;constraint:OnDelete:CASCADE"`
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
	AiringStatusID uint
	Day            int
	Month          int
	Year           int
	String         string
}

// CachedAiringStatus for storing anime airing status
type CachedAiringStatus struct {
	gorm.Model
	AnimeID uint
	FromID  uint
	ToID    uint
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
	AnimeID         uint
	AiringAt        int
	TimeUntilAiring int
	Episode         int
	IsNext          bool `gorm:"index"` // true if this is the next airing episode
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
	ImagesID       uint
	ScoresID       uint
	AiringStatusID uint

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
