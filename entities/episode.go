package entities

import (
	"time"
)

type EpisodeTitle struct {
	English  string `json:"english,omitempty"`
	Japanese string `json:"japanese,omitempty"`
	Romaji   string `json:"romaji,omitempty"`
}

type Episode struct {
	BaseModel
	EpisodeID     string       `gorm:"uniqueIndex;size:32" json:"id"`
	AnimeID       uint         `gorm:"index" json:"-"`
	Description   string       `gorm:"type:text" json:"description,omitempty"`
	Aired         string       `json:"aired,omitempty"`
	Score         float64      `json:"score,omitempty"`
	Filler        bool         `json:"filler,omitempty"`
	Recap         bool         `json:"recap,omitempty"`
	ForumURL      string       `json:"forum_url,omitempty"`
	URL           string       `json:"url,omitempty"`
	ThumbnailURL  string       `json:"thumbnail_url,omitempty"`
	EpisodeNumber int          `json:"episode_number,omitempty"`
	EpisodeLength float64      `json:"episode_length,omitempty"`
	Title         EpisodeTitle `gorm:"embedded;embeddedPrefix:title_" json:"titles"`
	StreamInfo    *StreamInfo       `gorm:"foreignKey:EpisodeID;references:EpisodeID" json:"streaming,omitempty"`
	SkipTimes     []EpisodeSkipTime `gorm:"foreignKey:EpisodeID;references:EpisodeID" json:"skip_times,omitempty"`
}

type EpisodeSkipTime struct {
	BaseModel
	EpisodeID string  `gorm:"index;size:32" json:"-"`
	SkipType  string  `gorm:"index" json:"skip_type"`
	StartTime float64 `json:"start_time"`
	EndTime   float64 `json:"end_time"`
}

type EpisodeSchedule struct {
	BaseModel
	AnimeID  uint `json:"-"`
	AiringAt int  `json:"airing_at,omitempty"`
	Episode  int  `json:"episode,omitempty"`
	IsNext   bool `gorm:"index" json:"is_next,omitempty"`
}

type StreamInfo struct {
	BaseModel
	EpisodeID  string            `gorm:"uniqueIndex:idx_episode_streaming;size:32" json:"-"`
	AnimeID    uint              `gorm:"uniqueIndex:idx_episode_streaming" json:"-"`
	SubSources []StreamingSource `gorm:"foreignKey:StreamInfoID;constraint:OnDelete:CASCADE" json:"sub_sources,omitempty"`
	DubSources []StreamingSource `gorm:"foreignKey:StreamInfoID;constraint:OnDelete:CASCADE" json:"dub_sources,omitempty"`
	LastFetch  time.Time         `json:"last_fetch,omitempty"`
}

type StreamingSource struct {
	BaseModel
	StreamInfoID uint   `json:"-"`
	URL          string `json:"url,omitempty"`
	Server       string `json:"server,omitempty"`
	Type         string `json:"type,omitempty"`
}
