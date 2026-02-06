package entities

import "gorm.io/gorm"

type Genre struct {
	gorm.Model
	Name    string  `json:"name,omitempty"`
	GenreID int     `gorm:"uniqueIndex" json:"genre_id,omitempty"`
	URL     string  `json:"url,omitempty"`
	Count   int     `gorm:"default:0" json:"count,omitempty"`
	Anime   []Anime `gorm:"many2many:anime_genres;" json:"anime,omitempty"`
}
