package entities

type ExternalURL struct {
	BaseModel
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type SimpleTitle struct {
	BaseModel
	Type  string `gorm:"uniqueIndex:idx_simple_title" json:"type,omitempty"`
	Title string `gorm:"uniqueIndex:idx_simple_title" json:"title,omitempty"`
}

type SimpleImage struct {
	BaseModel
	ImageURL string `gorm:"uniqueIndex" json:"image_url,omitempty"`
}
