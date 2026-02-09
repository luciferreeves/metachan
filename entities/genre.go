package entities

type Genre struct {
	BaseModel
	Name    string  `json:"name,omitempty"`
	GenreID int     `gorm:"uniqueIndex" json:"genre_id,omitempty"`
	URL     string  `json:"url,omitempty"`
	Count   int     `gorm:"default:0" json:"count,omitempty"`
	Anime   []Anime `gorm:"many2many:anime_genres;" json:"anime,omitempty"`
}
