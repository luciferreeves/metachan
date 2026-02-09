package entities

import (
	"metachan/enums"
)

type Mapping struct {
	BaseModel
	AniDB               int                    `json:"anidb,omitempty"`
	Anilist             int                    `json:"anilist,omitempty"`
	AnimeCountdown      int                    `json:"anime_countdown,omitempty"`
	AnimePlanet         string                 `json:"anime_planet,omitempty"`
	AniSearch           int                    `json:"ani_search,omitempty"`
	IMDB                string                 `json:"imdb,omitempty"`
	Kitsu               int                    `json:"kitsu,omitempty"`
	LiveChart           int                    `json:"live_chart,omitempty"`
	MAL                 int                    `gorm:"uniqueIndex" json:"mal,omitempty"`
	NotifyMoe           string                 `json:"notify_moe,omitempty"`
	Simkl               int                    `json:"simkl,omitempty"`
	TMDB                int                    `json:"tmdb,omitempty"`
	TVDB                int                    `json:"tvdb,omitempty"`
	Type                enums.MappingAnimeType `json:"type,omitempty"`
	MALAnilistComposite *string                `gorm:"uniqueIndex"`
}
