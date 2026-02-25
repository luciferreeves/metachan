package types

type JikanGenericPaginationEntity struct {
	LastVisiblePage int  `json:"last_visible_page"`
	HasNextPage     bool `json:"has_next_page"`
}

type JikanExtendedPaginationItems struct {
	Count   int `json:"count"`
	Total   int `json:"total"`
	PerPage int `json:"per_page"`
}

type JikanExtendedPagination struct {
	LastVisiblePage int                          `json:"last_visible_page"`
	HasNextPage     bool                         `json:"has_next_page"`
	CurrentPage     int                          `json:"current_page"`
	Items           JikanExtendedPaginationItems `json:"items"`
}

type JikanGenericTitleEntity struct {
	Type  string `json:"type"`
	Title string `json:"title"`
}

type JikanGenericImageSizeEntity struct {
	ImageURL        string `json:"image_url,omitempty"`
	SmallImageURL   string `json:"small_image_url,omitempty"`
	MediumImageURL  string `json:"medium_image_url,omitempty"`
	LargeImageURL   string `json:"large_image_url,omitempty"`
	MaximumImageURL string `json:"maximum_image_url,omitempty"`
}

type JikanGenericImageEntity struct {
	JPG  JikanGenericImageSizeEntity `json:"jpg"`
	WebP JikanGenericImageSizeEntity `json:"webp,omitempty"`
}

type JikanGenericRelatedEntity struct {
	MALID int    `json:"mal_id"`
	Type  string `json:"type"`
	URL   string `json:"url"`
	Name  string `json:"name"`
}

type JikanGenericDate struct {
	Day   int `json:"day"`
	Month int `json:"month"`
	Year  int `json:"year"`
}

type JikanGenericSchedule struct {
	From JikanGenericDate `json:"from"`
	To   JikanGenericDate `json:"to"`
}

type JikanGenericRelation struct {
	Relation string                      `json:"relation"`
	Entry    []JikanGenericRelatedEntity `json:"entry"`
}

type JikanGenericURL struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type JikanAiringSchedule struct {
	From   string               `json:"from"`
	To     string               `json:"to"`
	Prop   JikanGenericSchedule `json:"prop"`
	String string               `json:"string"`
}

type JikanBroadcastSchedule struct {
	Day      string `json:"day"`
	Time     string `json:"time"`
	Timezone string `json:"timezone"`
	String   string `json:"string"`
}

type JikanAnimeTrailer struct {
	YoutubeID string                  `json:"youtube_id"`
	URL       string                  `json:"url"`
	EmbedURL  string                  `json:"embed_url"`
	Images    JikanGenericImageEntity `json:"images"`
}

type JikanAnimeTheme struct {
	Openings []string `json:"openings"`
	Endings  []string `json:"endings"`
}

type JikanSingleAnime struct {
	MALID          int                         `json:"mal_id"`
	URL            string                      `json:"url"`
	Title          string                      `json:"title"`
	TitleEnglish   string                      `json:"title_english"`
	TitleJapanese  string                      `json:"title_japanese"`
	TitleSynonyms  []string                    `json:"title_synonyms"`
	Type           string                      `json:"type"`
	Source         string                      `json:"source"`
	Episodes       int                         `json:"episodes"`
	Status         string                      `json:"status"`
	Airing         bool                        `json:"airing"`
	Synopsis       string                      `json:"synopsis"`
	Score          float64                     `json:"score"`
	ScoredBy       int                         `json:"scored_by"`
	Rank           int                         `json:"rank"`
	Popularity     int                         `json:"popularity"`
	Members        int                         `json:"members"`
	Favorites      int                         `json:"favorites"`
	Season         string                      `json:"season"`
	Year           int                         `json:"year"`
	Images         JikanGenericImageEntity     `json:"images"`
	Trailer        JikanAnimeTrailer           `json:"trailer"`
	Approved       bool                        `json:"approved"`
	Titles         []JikanGenericTitleEntity   `json:"titles"`
	Aired          JikanAiringSchedule         `json:"aired"`
	Duration       string                      `json:"duration"`
	Rating         string                      `json:"rating"`
	Background     string                      `json:"background"`
	Broadcast      JikanBroadcastSchedule      `json:"broadcast"`
	Genres         []JikanGenericRelatedEntity `json:"genres"`
	ExplicitGenres []JikanGenericRelatedEntity `json:"explicit_genres"`
	Themes         []JikanGenericRelatedEntity `json:"themes"`
	Demographics   []JikanGenericRelatedEntity `json:"demographics"`
	Producers      []JikanGenericRelatedEntity `json:"producers"`
	Licensors      []JikanGenericRelatedEntity `json:"licensors"`
	Studios        []JikanGenericRelatedEntity `json:"studios"`
	Relations      []JikanGenericRelation      `json:"relations"`
	Theme          JikanAnimeTheme             `json:"theme"`
	External       []JikanGenericURL           `json:"external"`
	Streaming      []JikanGenericURL           `json:"streaming"`
}

type JikanAnimeResponse struct {
	Data JikanSingleAnime `json:"data"`
}

type JikanAnimeSingleEpisode struct {
	MALID         int     `json:"mal_id"`
	URL           string  `json:"url"`
	Title         string  `json:"title"`
	TitleJapanese string  `json:"title_japanese"`
	TitleRomaji   string  `json:"title_romaji"`
	Aired         string  `json:"aired"`
	Score         float64 `json:"score"`
	Filler        bool    `json:"filler"`
	Recap         bool    `json:"recap"`
	ForumURL      string  `json:"forum_url"`
}

type JikanAnimeEpisodeResponse struct {
	Pagination JikanGenericPaginationEntity `json:"pagination"`
	Data       []JikanAnimeSingleEpisode    `json:"data"`
}

type JikanCharacterPerson struct {
	MALID  int                     `json:"mal_id"`
	URL    string                  `json:"url"`
	Images JikanGenericImageEntity `json:"images"`
	Name   string                  `json:"name"`
}

type JikanSingleCharacter struct {
	Character   JikanCharacterPerson `json:"character"`
	Role        string               `json:"role"`
	VoiceActors []JikanVoiceActor    `json:"voice_actors"`
}

type JikanVoiceActor struct {
	Person   JikanCharacterPerson `json:"person"`
	Language string               `json:"language"`
}

type JikanAnimeCharacterResponse struct {
	Data []JikanSingleCharacter `json:"data"`
}

type JikanCharacterSimpleAnime struct {
	MALID  int                     `json:"mal_id"`
	URL    string                  `json:"url"`
	Images JikanGenericImageEntity `json:"images"`
	Title  string                  `json:"title"`
}

type JikanCharacterAnimeEntry struct {
	Role  string                    `json:"role"`
	Anime JikanCharacterSimpleAnime `json:"anime"`
}

type JikanFullCharacterData struct {
	MALID     int                        `json:"mal_id"`
	URL       string                     `json:"url"`
	Images    JikanGenericImageEntity    `json:"images"`
	Name      string                     `json:"name"`
	NameKanji string                     `json:"name_kanji"`
	Nicknames []string                   `json:"nicknames"`
	Favorites int                        `json:"favorites"`
	About     string                     `json:"about"`
	Anime     []JikanCharacterAnimeEntry `json:"anime"`
	Voices    []JikanVoiceActor          `json:"voices"`
}

type JikanCharacterFullResponse struct {
	Data JikanFullCharacterData `json:"data"`
}

type JikanGenre struct {
	MALID int    `json:"mal_id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count int    `json:"count"`
}

type JikanGenresResponse struct {
	Data []JikanGenre `json:"data"`
}

type JikanAnimeSearchResponse struct {
	Pagination JikanExtendedPagination `json:"pagination"`
	Data       []JikanSingleAnime      `json:"data"`
}

type JikanSingleProducer struct {
	MALID       int                       `json:"mal_id"`
	URL         string                    `json:"url"`
	Titles      []JikanGenericTitleEntity `json:"titles"`
	Images      JikanGenericImageEntity   `json:"images"`
	Favorites   int                       `json:"favorites"`
	Count       int                       `json:"count"`
	Established string                    `json:"established"`
	About       string                    `json:"about"`
	External    []JikanGenericURL         `json:"external,omitempty"`
}

type JikanProducersResponse struct {
	Data       []JikanSingleProducer        `json:"data"`
	Pagination JikanGenericPaginationEntity `json:"pagination"`
}

type JikanSingleProducerResponse struct {
	Data JikanSingleProducer `json:"data"`
}

type JikanPersonVoiceRole struct {
	Role      string                    `json:"role"`
	Anime     JikanCharacterSimpleAnime `json:"anime"`
	Character JikanCharacterPerson      `json:"character"`
}

type JikanPersonAnimeCredit struct {
	Position string                    `json:"position"`
	Anime    JikanCharacterSimpleAnime `json:"anime"`
}

type JikanPersonMangaCredit struct {
	Position string                    `json:"position"`
	Manga    JikanCharacterSimpleAnime `json:"manga"`
}

type JikanFullPersonData struct {
	MALID          int                      `json:"mal_id"`
	URL            string                   `json:"url"`
	WebsiteURL     *string                  `json:"website_url"`
	Images         JikanGenericImageEntity  `json:"images"`
	Name           string                   `json:"name"`
	GivenName      string                   `json:"given_name"`
	FamilyName     string                   `json:"family_name"`
	AlternateNames []string                 `json:"alternate_names"`
	Birthday       *string                  `json:"birthday"`
	Favorites      int                      `json:"favorites"`
	About          string                   `json:"about"`
	Anime          []JikanPersonAnimeCredit `json:"anime"`
	Manga          []JikanPersonMangaCredit `json:"manga"`
	Voices         []JikanPersonVoiceRole   `json:"voices"`
}

type JikanPersonFullResponse struct {
	Data JikanFullPersonData `json:"data"`
}
