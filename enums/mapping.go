package enums

type MappingType string

const (
	AniDB          MappingType = "anidb"
	Anilist        MappingType = "anilist"
	AnimeCountdown MappingType = "anime_countdown"
	AnimePlanet    MappingType = "anime_planet"
	AniSearch      MappingType = "ani_search"
	IMDB           MappingType = "imdb"
	Kitsu          MappingType = "kitsu"
	LiveChart      MappingType = "live_chart"
	MAL            MappingType = "mal"
	NotifyMoe      MappingType = "notify_moe"
	Simkl          MappingType = "simkl"
	TMDB           MappingType = "tmdb"
	TVDB           MappingType = "tvdb"
)

type MappingAnimeType string

const (
	SPECIAL MappingAnimeType = "SPECIAL"
	TV      MappingAnimeType = "TV"
	OVA     MappingAnimeType = "OVA"
	MOVIE   MappingAnimeType = "MOVIE"
	ONA     MappingAnimeType = "ONA"
	UNKNOWN MappingAnimeType = "UNKNOWN"
)
