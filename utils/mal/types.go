package mal

type ImageFormat struct {
	Small    string
	Medium   string
	Large    string
	Original string
}

type Image struct {
	JPG  ImageFormat
	WEBP ImageFormat
}

type Title struct {
	English  string
	Japanese string
	Romaji   string
	Synonyms []string
}

type AiredDate struct {
	Day    int
	Month  int
	Year   int
	String string
}

type Premiered struct {
	Season Season
	Year   int
}

type Aired struct {
	From   AiredDate
	To     AiredDate
	String string
}

type Broadcast struct {
	Day      string
	Time     string
	Timezone string
	String   string
}

type Statistics struct {
	Score      float64
	ScoredBy   int
	Rank       int
	Popularity int
	Members    int
	Favorites  int
}

type Preview struct {
	URL       string
	Thumbnail Image
}

type Trailer struct {
	YoutubeID string
	EmbedURL  string
	Preview
}

type EpisodeRange struct {
	Start int
	End   int
}

type ExternalLink struct {
	Name string
	URL  string
}

type ThemeSong struct {
	Title    Title
	Artist   string
	Episodes EpisodeRange
	Links    []ExternalLink
}

type PromotionalVideo struct {
	Title Title
	Preview
}

type MusicVideo struct {
	Title  Title
	Artist string
	Preview
}

type Episode struct {
	Number   int
	URL      string
	Title    Title
	Aired    AiredDate
	Score    float64
	Filler   bool
	Recap    bool
	ForumURL string
	Synopsis string
	Preview  Preview
}

type Anime struct {
	MALID        int
	URL          string
	Image        Image
	Title        Title
	Type         Type
	Source       Source
	Status       Status
	Airing       bool
	Rating       Rating
	Synopsis     string
	Background   string
	Duration     string
	EpisodeCount int
	Premiered    Premiered
	Aired        Aired
	Broadcast    Broadcast
	Statistics   Statistics
	Trailer      Trailer

	Openings    []ThemeSong
	Endings     []ThemeSong
	Videos      []PromotionalVideo
	MusicVideos []MusicVideo
	Episodes    []Episode

	Genres         []int
	ExplicitGenres []int
	Themes         []int
	Demographics   []int
	Producers      []int
	Studios        []int
	Licensors      []int

	External  []ExternalLink
	Streaming []ExternalLink
}