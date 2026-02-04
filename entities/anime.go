package entities

import (
	"time"

	"gorm.io/gorm"
)

type Anime struct {
	gorm.Model
	MALID         int               `gorm:"uniqueIndex" json:"id"`
	TitleID       uint              `json:"title_id,omitempty"`
	MappingID     uint              `json:"mapping_id,omitempty"`
	Synopsis      string            `gorm:"type:text" json:"synopsis,omitempty"`
	Type          string            `json:"type,omitempty"`
	Source        string            `json:"source,omitempty"`
	Airing        bool              `json:"airing,omitempty"`
	Status        string            `json:"status,omitempty"`
	Duration      string            `json:"duration,omitempty"`
	Color         string            `json:"color,omitempty"`
	Season        string            `json:"season,omitempty"`
	Year          int               `json:"year,omitempty"`
	SubbedCount   int               `json:"subbed_count,omitempty"`
	DubbedCount   int               `json:"dubbed_count,omitempty"`
	TotalEpisodes int               `json:"total_episodes,omitempty"`
	AiredEpisodes int               `json:"aired_episodes,omitempty"`
	LastUpdated   time.Time         `json:"last_updated,omitempty"`
	Title         *Title            `gorm:"foreignKey:TitleID" json:"titles,omitempty"`
	Mapping       *Mapping          `gorm:"foreignKey:MappingID" json:"mappings,omitempty"`
	Images        *Images           `gorm:"foreignKey:AnimeID" json:"images,omitempty"`
	Covers        *Images           `gorm:"foreignKey:AnimeID" json:"covers,omitempty"`
	Logos         *Logos            `gorm:"foreignKey:AnimeID" json:"logos,omitempty"`
	Scores        *Scores           `gorm:"foreignKey:AnimeID" json:"scores,omitempty"`
	AiringStatus  *AiringStatus     `gorm:"foreignKey:AnimeID" json:"airing_status,omitempty"`
	Broadcast     *Broadcast        `gorm:"foreignKey:AnimeID" json:"broadcast,omitempty"`
	NextAiring    *NextEpisode      `gorm:"foreignKey:AnimeID" json:"next_airing_episode,omitempty"`
	Genres        []Genre           `gorm:"many2many:anime_genres;" json:"genres,omitempty"`
	Producers     []Producer        `gorm:"many2many:anime_producers;" json:"producers,omitempty"`
	Studios       []Producer        `gorm:"many2many:anime_studios;" json:"studios,omitempty"`
	Licensors     []Producer        `gorm:"many2many:anime_licensors;" json:"licensors,omitempty"`
	Episodes      []Episode         `gorm:"foreignKey:AnimeID" json:"episodes,omitempty"`
	Characters    []Character       `gorm:"foreignKey:AnimeID" json:"characters,omitempty"`
	Schedule      []EpisodeSchedule `gorm:"foreignKey:AnimeID;constraint:OnDelete:CASCADE" json:"airing_schedule,omitempty"`
	Seasons       []Season          `gorm:"foreignKey:ParentAnimeID" json:"seasons,omitempty"`
}
