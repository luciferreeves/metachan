package entities

type Title struct {
	BaseModel
	English  string   `json:"english,omitempty"`
	Japanese string   `json:"japanese,omitempty"`
	Romaji   string   `json:"romaji,omitempty"`
	Synonyms []string `gorm:"serializer:json" json:"synonyms,omitempty"`
}

type Scores struct {
	BaseModel
	Score      float64 `json:"score,omitempty"`
	ScoredBy   int     `json:"scored_by,omitempty"`
	Rank       int     `json:"rank,omitempty"`
	Popularity int     `json:"popularity,omitempty"`
	Members    int     `json:"members,omitempty"`
	Favorites  int     `json:"favorites,omitempty"`
}

type Date struct {
	BaseModel
	Day    int    `json:"day,omitempty"`
	Month  int    `json:"month,omitempty"`
	Year   int    `json:"year,omitempty"`
	String string `json:"string,omitempty"`
}

type AiringStatus struct {
	BaseModel
	FromID *uint  `json:"-"`
	ToID   *uint  `json:"-"`
	String string `json:"string,omitempty"`
	From   *Date  `gorm:"foreignKey:FromID" json:"from,omitempty"`
	To     *Date  `gorm:"foreignKey:ToID" json:"to,omitempty"`
}

type Broadcast struct {
	BaseModel
	Day      string `json:"day,omitempty"`
	Time     string `json:"time,omitempty"`
	Timezone string `json:"timezone,omitempty"`
	String   string `json:"string,omitempty"`
}

type Images struct {
	BaseModel
	Small    string `json:"small,omitempty"`
	Large    string `json:"large,omitempty"`
	Original string `json:"original,omitempty"`
}

type Logos struct {
	BaseModel
	Small    string `json:"small,omitempty"`
	Medium   string `json:"medium,omitempty"`
	Large    string `json:"large,omitempty"`
	XLarge   string `json:"xlarge,omitempty"`
	Original string `json:"original,omitempty"`
}

type ExternalURL struct {
	BaseModel
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type SimpleTitle struct {
	BaseModel
	Type  string `json:"type,omitempty"`
	Title string `json:"title,omitempty"`
}

type SimpleImage struct {
	BaseModel
	ImageURL string `gorm:"uniqueIndex" json:"image_url,omitempty"`
}
