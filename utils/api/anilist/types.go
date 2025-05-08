package anilist

// AnilistAnimeResponse represents the response from AniList API
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
				} `json:"nodes"`
			} `json:"airingSchedule"`
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
