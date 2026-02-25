package entities

type Season struct {
	BaseModel
	ParentAnimeID uint   `json:"-"`
	MALID         int    `json:"mal_id,omitempty"`
	SeasonNumber  int    `json:"season_number,omitempty"`
	Current       bool   `json:"current,omitempty"`
	TitleEnglish  string `json:"title_english,omitempty"`
	TitleRomaji   string `json:"title_romaji,omitempty"`
	ImageOriginal string `json:"image_original,omitempty"`
	Year          int    `json:"year,omitempty"`
	SeasonName    string `json:"season,omitempty"`
	Type          string `json:"type,omitempty"`
	Status        string `json:"status,omitempty"`
}
