package entities

import "gorm.io/gorm"

type Producer struct {
	gorm.Model
	MALID        int           `gorm:"uniqueIndex" json:"mal_id,omitempty"`
	URL          string        `json:"url,omitempty"`
	Favorites    int           `json:"favorites,omitempty"`
	Count        int           `json:"count,omitempty"`
	Established  string        `json:"established,omitempty"`
	About        string        `gorm:"type:text" json:"about,omitempty"`
	ImageID      *uint         `json:"image_id,omitempty"`
	Image        *SimpleImage  `gorm:"foreignKey:ImageID" json:"image,omitempty"`
	Titles       []SimpleTitle `gorm:"many2many:producer_titles" json:"titles,omitempty"`
	ExternalURLs []ExternalURL `gorm:"many2many:producer_external_urls" json:"external_urls,omitempty"`
}
