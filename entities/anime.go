package entities

import "gorm.io/gorm"

type MappingType string

const (
	m_SPECIAL MappingType = "SPECIAL"
	m_TV      MappingType = "TV"
	m_OVA     MappingType = "OVA"
	m_MOVIE   MappingType = "MOVIE"
	m_ONA     MappingType = "ONA"
	m_UNKNOWN MappingType = "UNKNOWN"
)

type AnimeMapping struct {
	gorm.Model
	AniDB               int
	Anilist             int
	AnimeCountdown      int
	AnimePlanet         string
	AniSearch           int
	IMDB                string
	Kitsu               int
	LiveChart           int
	MAL                 int
	NotifyMoe           string
	Simkl               int
	TMDB                int
	TVDB                int
	Type                MappingType
	MALAnilistComposite *string `gorm:"uniqueIndex"`
}
