package malsync

// MALSyncStreamingSite represents a single streaming site entry in the MALSync API
type MALSyncStreamingSite struct {
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

// MALSyncSitesCollection represents the nested structure of streaming sites
// Format: map[platformName]map[identifier]siteObject
type MALSyncSitesCollection map[string]map[string]MALSyncStreamingSite

// MALSyncAnimeResponse is the top-level response from the MALSync API
type MALSyncAnimeResponse struct {
	ID      int                    `json:"id"`
	Type    string                 `json:"type"`
	Title   string                 `json:"title"`
	URL     string                 `json:"url"`
	Total   int                    `json:"total"`
	Image   string                 `json:"image"`
	AnidbID int                    `json:"anidbId,omitempty"`
	Sites   MALSyncSitesCollection `json:"Sites"`
}
