package entities

import "gorm.io/gorm"

type Title struct {
	gorm.Model
	English  string   `json:"english,omitempty"`
	Japanese string   `json:"japanese,omitempty"`
	Romaji   string   `json:"romaji,omitempty"`
	Synonyms []string `gorm:"serializer:json" json:"synonyms,omitempty"`
}

type Scores struct {
	gorm.Model
	Score      float64 `json:"score,omitempty"`
	ScoredBy   int     `json:"scored_by,omitempty"`
	Rank       int     `json:"rank,omitempty"`
	Popularity int     `json:"popularity,omitempty"`
	Members    int     `json:"members,omitempty"`
	Favorites  int     `json:"favorites,omitempty"`
}

type Date struct {
	gorm.Model
	Day    int    `json:"day,omitempty"`
	Month  int    `json:"month,omitempty"`
	Year   int    `json:"year,omitempty"`
	String string `json:"string,omitempty"`
}

type AiringStatus struct {
	gorm.Model
	FromID *uint  `json:"from_id,omitempty"`
	ToID   *uint  `json:"to_id,omitempty"`
	String string `json:"string,omitempty"`
	From   *Date  `gorm:"foreignKey:FromID" json:"from,omitempty"`
	To     *Date  `gorm:"foreignKey:ToID" json:"to,omitempty"`
}

type Broadcast struct {
	gorm.Model
	Day      string `json:"day,omitempty"`
	Time     string `json:"time,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	String   string `json:"string,omitempty"`
}

type Images struct {
	gorm.Model
	Small    string `json:"small,omitempty"`
	Large    string `json:"large,omitempty"`
	Original string `json:"original,omitempty"`
}

type Logos struct {
	gorm.Model
	Small    string `json:"small,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Large    string `json:"large,omitempty"`
	XLarge   string `json:"xlarge,omitempty"`
	Original string `json:"original,omitempty"`
}

type ExternalURL struct {
	gorm.Model
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type SimpleTitle struct {
	gorm.Model
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

type SimpleImage struct {
	gorm.Model
	ImageURL string `json:"image_url,omitempty"`
}
