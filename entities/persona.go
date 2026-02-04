package entities

import "gorm.io/gorm"

type Character struct {
	gorm.Model
	AnimeID     uint         `json:"anime_id,omitempty"`
	MALID       int          `json:"mal_id,omitempty"`
	URL         string       `json:"url,omitempty"`
	ImageURL    string       `json:"image_url,omitempty"`
	Name        string       `json:"name,omitempty"`
	Role        string       `json:"role,omitempty"`
	VoiceActors []VoiceActor `gorm:"foreignKey:CharacterID" json:"voice_actors,omitempty"`
}

type VoiceActor struct {
	gorm.Model
	CharacterID uint   `json:"character_id,omitempty"`
	MALID       int    `json:"mal_id,omitempty"`
	URL         string `json:"url,omitempty"`
	Image       string `json:"image_url,omitempty"`
	Name        string `json:"name,omitempty"`
	Language    string `json:"language,omitempty"`
}
