package types

type MalsyncStreamingSite struct {
	ID         int    `json:"id,omitempty"`
	Identifier any    `json:"identifier"`
	Image      string `json:"image,omitempty"`
	MalID      int    `json:"malId,omitempty"`
	AniID      int    `json:"aniId,omitempty"`
	Page       string `json:"page"`
	Title      string `json:"title"`
	Type       string `json:"type"`
	URL        string `json:"url"`
	External   bool   `json:"external,omitempty"`
}

type MalsyncSitesCollection map[string]map[string]MalsyncStreamingSite

type MalsyncAnimeResponse struct {
	ID      int                    `json:"id"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	URL     string                 `json:"url"`
	Total   int                    `json:"total"`
	Image   string                 `json:"image"`
	AnidbID int                    `json:"anidbId,omitempty"`
	Sites   MalsyncSitesCollection `json:"Sites"`
}
