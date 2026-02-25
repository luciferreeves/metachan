package entities

import (
	"time"
)

type AnimeTitle struct {
	English  string   `json:"english,omitempty"`
	Japanese string   `json:"japanese,omitempty"`
	Romaji   string   `json:"romaji,omitempty"`
	Synonyms []string `gorm:"serializer:json" json:"synonyms,omitempty"`
}

type AnimeScores struct {
	Score      float64 `json:"score,omitempty"`
	ScoredBy   int     `json:"scored_by,omitempty"`
	Rank       int     `json:"rank,omitempty"`
	Popularity int     `json:"popularity,omitempty"`
	Members    int     `json:"members,omitempty"`
	Favorites  int     `json:"favorites,omitempty"`
}

type AnimeImages struct {
	Small    string `json:"small,omitempty"`
	Large    string `json:"large,omitempty"`
	Original string `json:"original,omitempty"`
}

type AnimeLogos struct {
	Small    string `json:"small,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Large    string `json:"large,omitempty"`
	XLarge   string `json:"xlarge,omitempty"`
	Original string `json:"original,omitempty"`
}

type AnimeBroadcast struct {
	Day      string `json:"day,omitempty"`
	Time     string `json:"time,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	String   string `json:"string,omitempty"`
}

type AnimeAired struct {
	From   *time.Time `json:"from,omitempty"`
	To     *time.Time `json:"to,omitempty"`
	String string     `json:"string,omitempty"`
}

type AnimeTrailer struct {
	YoutubeID string `json:"youtube_id,omitempty"`
	URL       string `json:"url,omitempty"`
	EmbedURL  string `json:"embed_url,omitempty"`
}

type Anime struct {
	BaseModel
	MALID     int    `gorm:"uniqueIndex" json:"id"`
	MappingID uint   `json:"-"`
	Synopsis  string `gorm:"type:text" json:"synopsis,omitempty"`
	Type      string `json:"type,omitempty"`
	Source    string `json:"source,omitempty"`
	Airing    bool   `json:"airing,omitempty"`
	Status    string `json:"status,omitempty"`
	Duration  string `json:"duration,omitempty"`
	Color     string `json:"color,omitempty"`
	Season    string `json:"season,omitempty"`
	Year      int    `json:"year,omitempty"`

	Rating     string `json:"rating,omitempty"`
	Background string `gorm:"type:text" json:"background,omitempty"`

	SubbedCount       int `json:"subbed_count,omitempty"`
	DubbedCount       int `json:"dubbed_count,omitempty"`
	TotalEpisodes     int `json:"total_episodes,omitempty"`
	AiredEpisodes     int `json:"aired_episodes,omitempty"`
	SeasonNumber      int `json:"season_number,omitempty"`
	NextAiringAt      int `json:"next_airing_at,omitempty"`
	NextAiringEpisode int `json:"next_airing_episode,omitempty"`

	LastUpdated time.Time  `json:"-"`
	EnrichedAt  *time.Time `json:"-"`

	Title     AnimeTitle     `gorm:"embedded;embeddedPrefix:title_" json:"titles"`
	Scores    AnimeScores    `gorm:"embedded;embeddedPrefix:score_" json:"scores"`
	Images    AnimeImages    `gorm:"embedded;embeddedPrefix:image_" json:"images"`
	Covers    AnimeImages    `gorm:"embedded;embeddedPrefix:cover_" json:"covers"`
	Logos     AnimeLogos     `gorm:"embedded;embeddedPrefix:logo_" json:"logos"`
	Broadcast AnimeBroadcast `gorm:"embedded;embeddedPrefix:broadcast_" json:"broadcast"`
	Aired     AnimeAired     `gorm:"embedded;embeddedPrefix:aired_" json:"aired"`
	Trailer   AnimeTrailer   `gorm:"embedded;embeddedPrefix:trailer_" json:"trailer"`

	Mapping      *Mapping          `gorm:"foreignKey:MappingID" json:"mappings,omitempty"`
	Genres       []Genre           `gorm:"many2many:anime_genres;" json:"genres,omitempty"`
	Themes       []Genre           `gorm:"many2many:anime_themes;" json:"themes,omitempty"`
	Demographics []Genre           `gorm:"many2many:anime_demographics;" json:"demographics,omitempty"`
	Seasons      []Season          `gorm:"foreignKey:ParentAnimeID" json:"seasons,omitempty"`
	Producers    []Producer        `gorm:"many2many:anime_producers;" json:"producers,omitempty"`
	Studios      []Producer        `gorm:"many2many:anime_studios;" json:"studios,omitempty"`
	Licensors    []Producer        `gorm:"many2many:anime_licensors;" json:"licensors,omitempty"`
	Episodes     []Episode         `gorm:"foreignKey:AnimeID" json:"episodes,omitempty"`
	Characters   []Character       `gorm:"-" json:"characters,omitempty"`
	Schedule     []EpisodeSchedule `gorm:"foreignKey:AnimeID;constraint:OnDelete:CASCADE" json:"airing_schedule,omitempty"`
}
