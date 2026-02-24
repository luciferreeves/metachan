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

type Person struct {
	BaseModel
	MALID          int                    `gorm:"uniqueIndex" json:"mal_id,omitempty"`
	URL            string                 `json:"url,omitempty"`
	WebsiteURL     string                 `json:"website_url,omitempty"`
	Image          string                 `json:"image_url,omitempty"`
	Name           string                 `json:"name,omitempty"`
	GivenName      string                 `json:"given_name,omitempty"`
	FamilyName     string                 `json:"family_name,omitempty"`
	AlternateNames []string               `gorm:"serializer:json" json:"alternate_names,omitempty"`
	Birthday       *time.Time             `json:"birthday,omitempty"`
	Favorites      int                    `json:"favorites,omitempty"`
	About          string                 `gorm:"type:text" json:"about,omitempty"`
	EnrichedAt     *time.Time             `json:"-"`
	Characters     []PersonCharacterEntry `gorm:"-" json:"characters,omitempty"`
	VoiceRoles     []PersonVoiceRole      `gorm:"foreignKey:PersonID" json:"voices,omitempty"`
	AnimeCredits   []PersonAnimeCredit    `gorm:"foreignKey:PersonID" json:"anime,omitempty"`
	MangaCredits   []PersonMangaCredit    `gorm:"foreignKey:PersonID" json:"manga,omitempty"`
}

type PersonCharacterEntry struct {
	Character *Character `json:"character,omitempty"`
	Language  string     `json:"language,omitempty"`
}

type PersonVoiceRole struct {
	PersonID          uint   `gorm:"primaryKey" json:"-"`
	AnimeMALID        int    `gorm:"primaryKey" json:"anime_mal_id,omitempty"`
	CharacterMALID    int    `gorm:"primaryKey" json:"character_mal_id,omitempty"`
	Role              string `json:"role,omitempty"`
	AnimeTitle        string `json:"anime_title,omitempty"`
	AnimeURL          string `json:"anime_url,omitempty"`
	AnimeImageURL     string `json:"anime_image_url,omitempty"`
	CharacterName     string `json:"character_name,omitempty"`
	CharacterURL      string `json:"character_url,omitempty"`
	CharacterImageURL string `json:"character_image_url,omitempty"`
}

type PersonAnimeCredit struct {
	PersonID      uint   `gorm:"primaryKey" json:"-"`
	AnimeMALID    int    `gorm:"primaryKey" json:"mal_id,omitempty"`
	Position      string `json:"position,omitempty"`
	AnimeTitle    string `json:"title,omitempty"`
	AnimeURL      string `json:"url,omitempty"`
	AnimeImageURL string `json:"image_url,omitempty"`
}

type PersonMangaCredit struct {
	PersonID      uint   `gorm:"primaryKey" json:"-"`
	MangaMALID    int    `gorm:"primaryKey" json:"mal_id,omitempty"`
	Position      string `json:"position,omitempty"`
	MangaTitle    string `json:"title,omitempty"`
	MangaURL      string `json:"url,omitempty"`
	MangaImageURL string `json:"image_url,omitempty"`
}

type AnimeCharacter struct {
	AnimeID     uint `gorm:"primaryKey"`
	CharacterID uint `gorm:"primaryKey"`
	Role        string
}

type CharacterVoiceActor struct {
	CharacterID uint    `gorm:"primaryKey" json:"-"`
	PersonID    uint    `gorm:"primaryKey" json:"-"`
	Language    string  `json:"language,omitempty"`
	Person      *Person `gorm:"foreignKey:PersonID;references:ID" json:"person,omitempty"`
}
