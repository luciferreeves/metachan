package entities

type Season struct {
	BaseModel
	ParentAnimeID  uint          `json:"-"`
	MALID          int           `json:"mal_id,omitempty"`
	TitleID        uint          `json:"-"`
	ImagesID       *uint         `json:"-"`
	ScoresID       *uint         `json:"-"`
	AiringStatusID *uint         `json:"-"`
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
	SeasonNumber   int           `json:"season_number,omitempty"`
	Images         *Images       `gorm:"foreignKey:ImagesID" json:"images,omitempty"`
	Scores         *Scores       `gorm:"foreignKey:ScoresID" json:"scores,omitempty"`
	AiringStatus   *AiringStatus `gorm:"foreignKey:AiringStatusID" json:"airing_status,omitempty"`
}
