package entities

type Character struct {
	BaseModel
	AnimeID     uint         `json:"-"`
	MALID       int          `json:"mal_id,omitempty"`
	URL         string       `json:"url,omitempty"`
	ImageURL    string       `json:"image_url,omitempty"`
	Name        string       `json:"name,omitempty"`
	Role        string       `json:"role,omitempty"`
	VoiceActors []VoiceActor `gorm:"foreignKey:CharacterID" json:"voice_actors,omitempty"`
}

type VoiceActor struct {
	BaseModel
	CharacterID uint   `json:"-"`
	MALID       int    `json:"mal_id,omitempty"`
	URL         string `json:"url,omitempty"`
	Image       string `json:"image_url,omitempty"`
	Name        string `json:"name,omitempty"`
	Language    string `json:"language,omitempty"`
}
