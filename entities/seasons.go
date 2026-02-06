package entities

import "gorm.io/gorm"

type Season struct {
	gorm.Model
	ParentAnimeID  uint          `json:"parent_anime_id,omitempty"`
	MALID          int           `json:"mal_id,omitempty"`
	TitleID        uint          `json:"title_id,omitempty"`
	ImagesID       *uint         `json:"images_id,omitempty"`
	ScoresID       *uint         `json:"scores_id,omitempty"`
	AiringStatusID *uint         `json:"airing_status_id,omitempty"`
	Synopsis       string        `gorm:"type:text" json:"synopsis,omitempty"`
	Type           string        `json:"type,omitempty"`
	Source         string        `json:"source,omitempty"`
	Airing         bool          `json:"airing,omitempty"`
	Status         string        `json:"status,omitempty"`
	Duration       string        `json:"duration,omitempty"`
	Season         string        `json:"season,omitempty"`
	Year           int           `json:"year,omitempty"`
	Current        bool          `json:"current,omitempty"`
	Title          *Title        `gorm:"foreignKey:TitleID" json:"titles,omitempty"`
	Images         *Images       `gorm:"foreignKey:ImagesID" json:"images,omitempty"`
	Scores         *Scores       `gorm:"foreignKey:ScoresID" json:"scores,omitempty"`
	AiringStatus   *AiringStatus `gorm:"foreignKey:AiringStatusID" json:"airing_status,omitempty"`
}
