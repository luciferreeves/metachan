package types

import "metachan/enums"

type MappingResponse struct {
	AniDB          int                    `json:"anidb_id"`
	Anilist        int                    `json:"anilist_id"`
	AnimeCountdown int                    `json:"animecountdown_id"`
	AnimePlanet    any                    `json:"anime-planet_id"`
	AniSearch      int                    `json:"anisearch_id"`
	IMDB           string                 `json:"imdb_id"`
	Kitsu          int                    `json:"kitsu_id"`
	LiveChart      int                    `json:"livechart_id"`
	MAL            int                    `json:"mal_id"`
	NotifyMoe      string                 `json:"notify.moe_id"`
	Simkl          int                    `json:"simkl_id"`
	TMDB           any                    `json:"themoviedb_id"`
	TVDB           int                    `json:"thetvdb_id"`
	Type           enums.MappingAnimeType `json:"type"`
}
