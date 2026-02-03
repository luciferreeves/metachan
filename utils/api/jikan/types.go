package jikan

// JikanPagination represents the pagination data in Jikan API responses
type JikanPagination struct {
	LastVisiblePage int  `json:"last_visible_page"`
	HasNextPage     bool `json:"has_next_page"`
}

// JikanGenericMALStructure represents a common structure for various MAL entities
type JikanGenericMALStructure struct {
	MALID int    `json:"mal_id"`
	Type  string `json:"type"`
	URL   string `json:"url"`
	Name  string `json:"name"`
}

// JikanGenre represents a genre from Jikan genres API
type JikanGenre struct {
	MALID int    `json:"mal_id"`
	Name  string `json:"name"`
	URL   string `json:"url"`
	Count int    `json:"count"`
}

// JikanGenresResponse represents the genres response from Jikan API
type JikanGenresResponse struct {
	Data []JikanGenre `json:"data"`
}

// JikanAnimeListItem represents a single anime in list responses
type JikanAnimeListItem struct {
	MALID         int      `json:"mal_id"`
	URL           string   `json:"url"`
	Title         string   `json:"title"`
	TitleEnglish  string   `json:"title_english"`
	TitleJapanese string   `json:"title_japanese"`
	TitleSynonyms []string `json:"title_synonyms"`
	Type          string   `json:"type"`
	Source        string   `json:"source"`
	Episodes      int      `json:"episodes"`
	Status        string   `json:"status"`
	Airing        bool     `json:"airing"`
	Synopsis      string   `json:"synopsis"`
	Score         float64  `json:"score"`
	ScoredBy      int      `json:"scored_by"`
	Rank          int      `json:"rank"`
	Popularity    int      `json:"popularity"`
	Members       int      `json:"members"`
	Favorites     int      `json:"favorites"`
	Season        string   `json:"season"`
	Year          int      `json:"year"`
	Images        struct {
		JPG struct {
			ImageURL      string `json:"image_url"`
			SmallImageURL string `json:"small_image_url"`
			LargeImageURL string `json:"large_image_url"`
		} `json:"jpg"`
	} `json:"images"`
	Genres         []JikanGenericMALStructure `json:"genres"`
	ExplicitGenres []JikanGenericMALStructure `json:"explicit_genres"`
	Producers      []JikanGenericMALStructure `json:"producers"`
	Licensors      []JikanGenericMALStructure `json:"licensors"`
	Studios        []JikanGenericMALStructure `json:"studios"`
}

// JikanAnimeListResponse represents paginated anime list response
type JikanAnimeListResponse struct {
	Pagination struct {
		LastVisiblePage int  `json:"last_visible_page"`
		HasNextPage     bool `json:"has_next_page"`
		CurrentPage     int  `json:"current_page"`
		Items           struct {
			Count   int `json:"count"`
			Total   int `json:"total"`
			PerPage int `json:"per_page"`
		} `json:"items"`
	} `json:"pagination"`
	Data []JikanAnimeListItem `json:"data"`
}

// JikanAnimeResponse represents the main anime response from Jikan API
type JikanAnimeResponse struct {
	Data struct {
		MALID  int    `json:"mal_id"`
		URL    string `json:"url"`
		Images struct {
			JPG struct {
				ImageURL      string `json:"image_url"`
				SmallImageURL string `json:"small_image_url"`
				LargeImageURL string `json:"large_image_url"`
			} `json:"jpg"`
			WebP struct {
				ImageURL      string `json:"image_url"`
				SmallImageURL string `json:"small_image_url"`
				LargeImageURL string `json:"large_image_url"`
			} `json:"webp"`
		} `json:"images"`
		Trailer struct {
			YoutubeID string `json:"youtube_id"`
			URL       string `json:"url"`
			EmbedURL  string `json:"embed_url"`
			Images    struct {
				ImageURL        string `json:"image_url"`
				SmallImageURL   string `json:"small_image_url"`
				MediumImageURL  string `json:"medium_image_url"`
				LargeImageURL   string `json:"large_image_url"`
				MaximumImageURL string `json:"maximum_image_url"`
			} `json:"images"`
		} `json:"trailer"`
		Approved bool `json:"approved"`
		Titles   []struct {
			Type  string `json:"type"`
			Title string `json:"title"`
		} `json:"titles"`
		Title         string   `json:"title"`
		TitleEnglish  string   `json:"title_english"`
		TitleJapanese string   `json:"title_japanese"`
		TitleSynonyms []string `json:"title_synonyms"`
		Type          string   `json:"type"`
		Source        string   `json:"source"`
		Episodes      int      `json:"episodes"`
		Status        string   `json:"status"`
		Airing        bool     `json:"airing"`
		Aired         struct {
			From string `json:"from"`
			To   string `json:"to"`
			Prop struct {
				From struct {
					Day   int `json:"day"`
					Month int `json:"month"`
					Year  int `json:"year"`
				} `json:"from"`
				To struct {
					Day   int `json:"day"`
					Month int `json:"month"`
					Year  int `json:"year"`
				} `json:"to"`
			} `json:"prop"`
			String string `json:"string"`
		} `json:"aired"`
		Duration   string  `json:"duration"`
		Rating     string  `json:"rating"`
		Score      float64 `json:"score"`
		ScoredBy   int     `json:"scored_by"`
		Rank       int     `json:"rank"`
		Popularity int     `json:"popularity"`
		Members    int     `json:"members"`
		Favorites  int     `json:"favorites"`
		Synopsis   string  `json:"synopsis"`
		Background string  `json:"background"`
		Season     string  `json:"season"`
		Year       int     `json:"year"`
		Broadcast  struct {
			Day      string `json:"day"`
			Time     string `json:"time"`
			Timezone string `json:"timezone"`
			String   string `json:"string"`
		} `json:"broadcast"`
		Producers      []JikanGenericMALStructure `json:"producers"`
		Licensors      []JikanGenericMALStructure `json:"licensors"`
		Studios        []JikanGenericMALStructure `json:"studios"`
		Genres         []JikanGenericMALStructure `json:"genres"`
		ExplicitGenres []JikanGenericMALStructure `json:"explicit_genres"`
		Themes         []JikanGenericMALStructure `json:"themes"`
		Demographics   []JikanGenericMALStructure `json:"demographics"`
		Relations      []struct {
			Relation string                     `json:"relation"`
			Entry    []JikanGenericMALStructure `json:"entry"`
		} `json:"relations"`
		Theme struct {
			Openings []string `json:"openings"`
			Endings  []string `json:"endings"`
		} `json:"theme"`
		External []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"external"`
		Streaming []struct {
			Name string `json:"name"`
			URL  string `json:"url"`
		} `json:"streaming"`
	} `json:"data"`
}

// JikanAnimeEpisode represents an episode from Jikan API
type JikanAnimeEpisode struct {
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

// JikanAnimeEpisodeResponse represents the episodes response from Jikan API
type JikanAnimeEpisodeResponse struct {
	Pagination JikanPagination     `json:"pagination"`
	Data       []JikanAnimeEpisode `json:"data"`
}

// JikanAnimeCharacterResponse represents the characters response from Jikan API
type JikanAnimeCharacterResponse struct {
	Data []struct {
		Character struct {
			MALID  int    `json:"mal_id"`
			URL    string `json:"url"`
			Images struct {
				JPG struct {
					ImageURL      string `json:"image_url"`
					SmallImageURL string `json:"small_image_url"`
				} `json:"jpg"`
				WebP struct {
					ImageURL      string `json:"image_url"`
					SmallImageURL string `json:"small_image_url"`
				} `json:"webp"`
			} `json:"images"`
			Name string `json:"name"`
		} `json:"character"`
		Role        string `json:"role"`
		VoiceActors []struct {
			Person struct {
				MALID  int    `json:"mal_id"`
				URL    string `json:"url"`
				Images struct {
					JPG struct {
						ImageURL string `json:"image_url"`
					} `json:"jpg"`
					WebP struct {
						ImageURL string `json:"image_url"`
					} `json:"webp"`
				} `json:"images"`
				Name string `json:"name"`
			} `json:"person"`
			Language string `json:"language"`
		} `json:"voice_actors"`
	} `json:"data"`
}
