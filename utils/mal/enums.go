package mal

type Type string

const (
	TypeTV      Type = "TV"
	TypeMovie   Type = "Movie"
	TypeOVA     Type = "OVA"
	TypeONA     Type = "ONA"
	TypeSpecial Type = "Special"
	TypeMusic   Type = "Music"
	TypeUnknown Type = "Unknown"
)

type Status string

const (
	StatusAiring      Status = "Currently Airing"
	StatusFinished    Status = "Finished Airing"
	StatusNotYetAired Status = "Not yet aired"
)

type Source string

const (
	SourceOriginal      Source = "Original"
	SourceManga         Source = "Manga"
	SourceLightNovel    Source = "Light novel"
	SourceVisualNovel   Source = "Visual novel"
	SourceGame          Source = "Game"
	SourceNovel         Source = "Novel"
	SourceWebManga      Source = "Web manga"
	SourceWebNovel      Source = "Web novel"
	SourceCardGame      Source = "Card game"
	SourceFourKomaManga Source = "4-koma manga"
	SourceBook          Source = "Book"
	SourcePictureBook   Source = "Picture book"
	SourceRadio         Source = "Radio"
	SourceMusic         Source = "Music"
	SourceOther         Source = "Other"
	SourceUnknown       Source = "Unknown"
)

type Rating string

const (
	RatingG    Rating = "G - All Ages"
	RatingPG   Rating = "PG - Children"
	RatingPG13 Rating = "PG-13 - Teens 13 or older"
	RatingR17  Rating = "R - 17+ (violence & profanity)"
	RatingR    Rating = "R+ - Mild Nudity"
	RatingRx   Rating = "Rx - Hentai"
)

type Season string

const (
	SeasonWinter Season = "Winter"
	SeasonSpring Season = "Spring"
	SeasonSummer Season = "Summer"
	SeasonFall   Season = "Fall"
)