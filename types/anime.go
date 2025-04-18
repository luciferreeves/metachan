package types

type AnimeTitles struct {
	English  string   `json:"english"`
	Japanese string   `json:"japanese"`
	Romaji   string   `json:"romaji"`
	Synonyms []string `json:"synonyms"`
}

type AnimeMappings struct {
	AniDB          int    `json:"anidb"`
	Anilist        int    `json:"anilist"`
	AnimeCountdown int    `json:"anime_countdown"`
	AnimePlanet    string `json:"anime_planet"`
	AniSearch      int    `json:"ani_search"`
	IMDB           string `json:"imdb"`
	Kitsu          int    `json:"kitsu"`
	LiveChart      int    `json:"live_chart"`
	NotifyMoe      string `json:"notify_moe"`
	Simkl          int    `json:"simkl"`
	TMDB           int    `json:"tmdb"`
	TVDB           int    `json:"tvdb"`
}

type EpisodeTitles struct {
	English  string `json:"english"`
	Japanese string `json:"japanese"`
	Romaji   string `json:"romaji"`
}

type AnimeSingleEpisode struct {
	Titles      EpisodeTitles `json:"titles"`
	Description string        `json:"description"`
	Aired       string        `json:"aired"`
	Score       float64       `json:"score"`
	Filler      bool          `json:"filler"`
	Recap       bool          `json:"recap"`
	ForumURL    string        `json:"forum_url"`
	URL         string        `json:"url"`
}

type AnimeEpisodes struct {
	Total    int                  `json:"total"`
	Aired    int                  `json:"aired"`
	Episodes []AnimeSingleEpisode `json:"episodes"`
}

type Anime struct {
	MALID    int           `json:"id"`
	Titles   AnimeTitles   `json:"titles"`
	Synopsis string        `json:"synopsis"`
	Type     AniSyncType   `json:"type"`
	Source   string        `json:"source"`
	Status   string        `json:"status"`
	Duration string        `json:"duration"`
	Episodes AnimeEpisodes `json:"episodes"`
	Mappings AnimeMappings `json:"mappings"`
}

type JikanPagination struct {
	LastVisiblePage int  `json:"last_visible_page"`
	HasNextPage     bool `json:"has_next_page"`
}

type JikanGenericMALStructure struct {
	MALID int    `json:"mal_id"`
	Type  string `json:"type"`
	URL   string `json:"url"`
	Name  string `json:"name"`
}

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

type JikanAnimeEpisodeResponse struct {
	Pagination JikanPagination     `json:"pagination"`
	Data       []JikanAnimeEpisode `json:"data"`
}

type AnilistAnimeResponse struct {
	Data struct {
		Media struct {
			ID    int `json:"id"`
			MALID int `json:"idMal"`
			Title struct {
				Romaji        string `json:"romaji"`
				English       string `json:"english"`
				Native        string `json:"native"`
				UserPreferred string `json:"userPreferred"`
			} `json:"title"`
			Type        string `json:"type"`
			Format      string `json:"format"`
			Status      string `json:"status"`
			Description string `json:"description"`
			StartDate   struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"startDate"`
			EndDate struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"endDate"`
			Season          string `json:"season"`
			SeasonYear      int    `json:"seasonYear"`
			Episodes        int    `json:"episodes"`
			Duration        int    `json:"duration"`
			Chapters        int    `json:"chapters"`
			Volumes         int    `json:"volumes"`
			CountryOfOrigin string `json:"countryOfOrigin"`
			IsLicensed      bool   `json:"isLicensed"`
			Source          string `json:"source"`
			Hashtag         string `json:"hashtag"`
			Trailer         struct {
				ID        string `json:"id"`
				Site      string `json:"site"`
				Thumbnail string `json:"thumbnail"`
			} `json:"trailer"`
			CoverImage struct {
				ExtraLarge string `json:"extraLarge"`
				Large      string `json:"large"`
				Medium     string `json:"medium"`
				Color      string `json:"color"`
			} `json:"coverImage"`
			BannerImage  string   `json:"bannerImage"`
			Genres       []string `json:"genres"`
			Synonyms     []string `json:"synonyms"`
			AverageScore int      `json:"averageScore"`
			MeanScore    int      `json:"meanScore"`
			Popularity   int      `json:"popularity"`
			IsLocked     bool     `json:"isLocked"`
			Trending     int      `json:"trending"`
			Favorites    int      `json:"favorites"`
			Tags         []struct {
				ID               int    `json:"id"`
				Name             string `json:"name"`
				Description      string `json:"description"`
				Category         string `json:"category"`
				Rank             int    `json:"rank"`
				IsGeneralSpoiler bool   `json:"isGeneralSpoiler"`
				IsMediaSpoiler   bool   `json:"isMediaSpoiler"`
				IsAdult          bool   `json:"isAdult"`
			} `json:"tags"`
			Relations struct {
				Edges []struct {
					ID           int    `json:"id"`
					RelationType string `json:"relationType"`
					Node         struct {
						ID    int `json:"id"`
						Title struct {
							Romaji        string `json:"romaji"`
							English       string `json:"english"`
							Native        string `json:"native"`
							UserPreferred string `json:"userPreferred"`
						} `json:"title"`
						Format     string `json:"format"`
						Type       string `json:"type"`
						Status     string `json:"status"`
						CoverImage struct {
							ExtraLarge string `json:"extraLarge"`
							Large      string `json:"large"`
							Medium     string `json:"medium"`
							Color      string `json:"color"`
						} `json:"coverImage"`
						BannerImage string `json:"bannerImage"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"relations"`
			Characters struct {
				Edges []struct {
					Role string `json:"role"`
					Node struct {
						ID   int `json:"id"`
						Name struct {
							First         string `json:"first"`
							Last          string `json:"last"`
							Middle        string `json:"middle"`
							Full          string `json:"full"`
							Native        string `json:"native"`
							UserPreferred string `json:"userPreferred"`
						} `json:"name"`
						Image struct {
							Large  string `json:"large"`
							Medium string `json:"medium"`
						} `json:"image"`
						Description string `json:"description"`
						Age         string `json:"age"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"characters"`
			Staff struct {
				Edges []struct {
					Role string `json:"role"`
					Node struct {
						ID   int `json:"id"`
						Name struct {
							First         string `json:"first"`
							Last          string `json:"last"`
							Middle        string `json:"middle"`
							Full          string `json:"full"`
							Native        string `json:"native"`
							UserPreferred string `json:"userPreferred"`
						} `json:"name"`
						Image struct {
							Large  string `json:"large"`
							Medium string `json:"medium"`
						} `json:"image"`
						Description        string   `json:"description"`
						PrimaryOccupations []string `json:"primaryOccupations"`
						Gender             string   `json:"gender"`
						Age                int      `json:"age"`
						LanguageV2         string   `json:"languageV2"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"staff"`
			Studios struct {
				Edges []struct {
					IsMain bool `json:"isMain"`
					Node   struct {
						ID   int    `json:"id"`
						Name string `json:"name"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"studios"`
			IsAdult           bool `json:"isAdult"`
			NextAiringEpisode struct {
				ID              int `json:"id"`
				AiringAt        int `json:"airingAt"`
				TimeUntilAiring int `json:"timeUntilAiring"`
				Episode         int `json:"episode"`
			} `json:"nextAiringEpisode"`
			AiringSchedule struct {
				Nodes []struct {
					ID              int `json:"id"`
					Episode         int `json:"episode"`
					AiringAt        int `json:"airingAt"`
					TimeUntilAiring int `json:"timeUntilAiring"`
				}
			} `json:"nodes"`
			Trends struct {
				Nodes []struct {
					Date       int `json:"date"`
					Trending   int `json:"trending"`
					Popularity int `json:"popularity"`
					InProgress int `json:"inProgress"`
				} `json:"nodes"`
			} `json:"trends"`
			ExternalLinks []struct {
				ID   int    `json:"id"`
				URL  string `json:"url"`
				Site string `json:"site"`
			} `json:"externalLinks"`
			StreamingEpisodes []struct {
				Title     string `json:"title"`
				Thumbnail string `json:"thumbnail"`
				URL       string `json:"url"`
				Site      string `json:"site"`
			} `json:"streamingEpisodes"`
			Rankings []struct {
				ID      int    `json:"id"`
				Rank    int    `json:"rank"`
				Type    string `json:"type"`
				Format  string `json:"format"`
				Year    int    `json:"year"`
				Season  string `json:"season"`
				AllTime bool   `json:"allTime"`
				Context string `json:"context"`
			} `json:"rankings"`
			Stats struct {
				ScoreDistribution []struct {
					Score  int `json:"score"`
					Amount int `json:"amount"`
				} `json:"scoreDistribution"`
				StatusDistribution []struct {
					Status string `json:"status"`
					Amount int    `json:"amount"`
				} `json:"statusDistribution"`
			} `json:"stats"`
			SiteURL string `json:"siteUrl"`
		} `json:"media"`
	} `json:"data"`
}
