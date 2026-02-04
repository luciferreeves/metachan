package entities

import (
	"time"

	"gorm.io/gorm"
)

type Episode struct {
	gorm.Model
	EpisodeID    string      `gorm:"uniqueIndex;size:32" json:"id"`
	AnimeID      uint        `json:"anime_id,omitempty"`
	TitleID      uint        `json:"title_id,omitempty"`
	Description  string      `gorm:"type:text" json:"description,omitempty"`
	Aired        string      `json:"aired,omitempty"`
	Score        float64     `json:"score,omitempty"`
	Filler       bool        `json:"filler,omitempty"`
	Recap        bool        `json:"recap,omitempty"`
	ForumURL     string      `json:"forum_url,omitempty"`
	URL          string      `json:"url,omitempty"`
	ThumbnailURL string      `json:"thumbnail_url,omitempty"`
	Title        *Title      `gorm:"foreignKey:TitleID" json:"titles,omitempty"`
	StreamInfo   *StreamInfo `gorm:"foreignKey:EpisodeID;references:EpisodeID" json:"streaming,omitempty"`
}

type EpisodeSchedule struct {
	gorm.Model
	AnimeID  uint `json:"anime_id,omitempty"`
	AiringAt int  `json:"airing_at,omitempty"`
	Episode  int  `json:"episode,omitempty"`
	IsNext   bool `gorm:"index" json:"is_next,omitempty"`
}

type NextEpisode struct {
	gorm.Model
	AnimeID  uint `json:"anime_id,omitempty"`
	AiringAt int  `json:"airing_at,omitempty"`
	Episode  int  `json:"episode,omitempty"`
}

type StreamInfo struct {
	gorm.Model
	EpisodeID  string            `gorm:"uniqueIndex:idx_episode_streaming;size:32" json:"episode_id"`
	AnimeID    uint              `gorm:"uniqueIndex:idx_episode_streaming" json:"anime_id,omitempty"`
	SubSources []StreamingSource `gorm:"foreignKey:StreamInfoID;constraint:OnDelete:CASCADE" json:"sub_sources,omitempty"`
	DubSources []StreamingSource `gorm:"foreignKey:StreamInfoID;constraint:OnDelete:CASCADE" json:"dub_sources,omitempty"`
	LastFetch  time.Time         `json:"last_fetch,omitempty"`
}

type StreamingSource struct {
	gorm.Model
	StreamInfoID uint   `json:"stream_info_id,omitempty"`
	URL          string `json:"url,omitempty"`
	Server       string `json:"server,omitempty"`
	Type         string `json:"type,omitempty"`
}
