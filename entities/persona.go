package entities

type Character struct {
	BaseModel
	MALID       int                   `gorm:"uniqueIndex" json:"mal_id,omitempty"`
	URL         string                `json:"url,omitempty"`
	ImageURL    string                `json:"image_url,omitempty"`
	Name        string                `json:"name,omitempty"`
	Role        string                `gorm:"-" json:"role,omitempty"`
	VoiceActors []CharacterVoiceActor `gorm:"foreignKey:CharacterID" json:"voice_actors,omitempty"`
}

type VoiceActor struct {
	BaseModel
	MALID int    `gorm:"uniqueIndex" json:"mal_id,omitempty"`
	URL   string `json:"url,omitempty"`
	Image string `json:"image_url,omitempty"`
	Name  string `json:"name,omitempty"`
}

type AnimeCharacter struct {
	AnimeID     uint `gorm:"primaryKey"`
	CharacterID uint `gorm:"primaryKey"`
	Role        string
}

type CharacterVoiceActor struct {
	CharacterID  uint        `gorm:"primaryKey" json:"-"`
	VoiceActorID uint        `gorm:"primaryKey" json:"-"`
	Language     string      `json:"language,omitempty"`
	VoiceActor   *VoiceActor `gorm:"foreignKey:ID;references:VoiceActorID" json:"voice_actor,omitempty"`
}
