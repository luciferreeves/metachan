package entities

import "time"

type Character struct {
	BaseModel
	MALID            int                        `gorm:"uniqueIndex" json:"mal_id,omitempty"`
	URL              string                     `json:"url,omitempty"`
	ImageURL         string                     `json:"image_url,omitempty"`
	Name             string                     `json:"name,omitempty"`
	NameKanji        string                     `json:"name_kanji,omitempty"`
	Nicknames        []string                   `gorm:"serializer:json" json:"nicknames,omitempty"`
	Favorites        int                        `json:"favorites,omitempty"`
	About            string                     `gorm:"type:text" json:"about,omitempty"`
	EnrichedAt       *time.Time                 `json:"-"`
	Role             string                     `gorm:"-" json:"role,omitempty"`
	VoiceActors      []CharacterVoiceActor      `gorm:"foreignKey:CharacterID" json:"voice_actors,omitempty"`
	AnimeAppearances []CharacterAnimeAppearance `gorm:"foreignKey:CharacterID" json:"anime,omitempty"`
}

type CharacterAnimeAppearance struct {
	CharacterID uint   `gorm:"primaryKey" json:"-"`
	AnimeMALID  int    `gorm:"primaryKey" json:"mal_id,omitempty"`
	Title       string `json:"title,omitempty"`
	URL         string `json:"url,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	Role        string `json:"role,omitempty"`
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
	VoiceActor   *VoiceActor `gorm:"foreignKey:VoiceActorID;references:ID" json:"voice_actor,omitempty"`
}
