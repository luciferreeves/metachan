package anime

import (
	"fmt"
	"metachan/database"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
	"sort"
	"strings"
)

// GetAnimeSeason prepares season information for an anime
func GetAnimeSeason(mappings *[]entities.AnimeMapping, currentMALID int) ([]types.AnimeSeason, error) {
	var seasons []types.AnimeSeason

	for _, mapping := range *mappings {
		// Fetch basic anime info from Jikan API
		animeDetails, err := getAnimeViaJikan(mapping.MAL)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to get anime details for MAL ID %d: %v", mapping.MAL, err), types.LogOptions{
				Level:  types.Warn,
				Prefix: "AnimeSeason",
			})
			continue
		}

		// Build the season object
		season := types.AnimeSeason{
			MALID: mapping.MAL,
			Titles: types.AnimeTitles{
				English:  animeDetails.Data.TitleEnglish,
				Japanese: animeDetails.Data.TitleJapanese,
				Romaji:   animeDetails.Data.Title,
				Synonyms: animeDetails.Data.TitleSynonyms,
			},
			Synopsis: animeDetails.Data.Synopsis,
			Type:     types.AniSyncType(mapping.Type),
			Source:   animeDetails.Data.Source,
			Airing:   animeDetails.Data.Airing,
			Status:   animeDetails.Data.Status,
			AiringStatus: types.AiringStatus{
				From: types.AiringStatusDates{
					Day:    animeDetails.Data.Aired.Prop.From.Day,
					Month:  animeDetails.Data.Aired.Prop.From.Month,
					Year:   animeDetails.Data.Aired.Prop.From.Year,
					String: animeDetails.Data.Aired.From,
				},
				To: types.AiringStatusDates{
					Day:    animeDetails.Data.Aired.Prop.To.Day,
					Month:  animeDetails.Data.Aired.Prop.To.Month,
					Year:   animeDetails.Data.Aired.Prop.To.Year,
					String: animeDetails.Data.Aired.To,
				},
				String: animeDetails.Data.Aired.String,
			},
			Duration: animeDetails.Data.Duration,
			Images: types.AnimeImages{
				Small:    animeDetails.Data.Images.JPG.SmallImageURL,
				Large:    animeDetails.Data.Images.JPG.LargeImageURL,
				Original: animeDetails.Data.Images.JPG.ImageURL,
			},
			Scores: types.AnimeScores{
				Score:      animeDetails.Data.Score,
				ScoredBy:   animeDetails.Data.ScoredBy,
				Rank:       animeDetails.Data.Rank,
				Popularity: animeDetails.Data.Popularity,
				Members:    animeDetails.Data.Members,
				Favorites:  animeDetails.Data.Favorites,
			},
			Season:  animeDetails.Data.Season,
			Year:    animeDetails.Data.Year,
			Current: mapping.MAL == currentMALID, // Mark if this is the current season
		}

		seasons = append(seasons, season)
	}

	// Sort seasons chronologically by air date
	if len(seasons) > 1 {
		sortSeasonsByAirDate(&seasons)
		logger.Log(fmt.Sprintf("Found and sorted %d seasons for anime", len(seasons)), types.LogOptions{
			Level:  types.Info,
			Prefix: "AnimeSeason",
		})
	}

	return seasons, nil
}

// sortSeasonsByAirDate sorts the seasons array chronologically by air date
func sortSeasonsByAirDate(seasons *[]types.AnimeSeason) {
	// Use a slice instead of a pointer to slice to make the code cleaner
	s := *seasons

	// Sort by air date using the structured fields (year, month, day)
	sort.Slice(s, func(i, j int) bool {
		// Compare years first
		if s[i].AiringStatus.From.Year != s[j].AiringStatus.From.Year {
			return s[i].AiringStatus.From.Year < s[j].AiringStatus.From.Year
		}

		// If years are equal, compare months
		if s[i].AiringStatus.From.Month != s[j].AiringStatus.From.Month {
			return s[i].AiringStatus.From.Month < s[j].AiringStatus.From.Month
		}

		// If months are equal, compare days
		if s[i].AiringStatus.From.Day != s[j].AiringStatus.From.Day {
			return s[i].AiringStatus.From.Day < s[j].AiringStatus.From.Day
		}

		// If all date fields are equal, use season as a tiebreaker
		if s[i].Season != s[j].Season {
			seasonOrder := map[string]int{
				"winter": 0,
				"spring": 1,
				"summer": 2,
				"fall":   3,
				"":       4, // Unknown season comes last
			}

			seasonI := strings.ToLower(s[i].Season)
			seasonJ := strings.ToLower(s[j].Season)

			return seasonOrder[seasonI] < seasonOrder[seasonJ]
		}

		// If everything is equal, preserve original order (stable sort)
		return i < j
	})

	// Update the original slice
	*seasons = s
}

// FindSeasonMappings finds all anime mappings that belong to the same series based on TVDB ID
func FindSeasonMappings(tvdbID int) ([]entities.AnimeMapping, error) {
	logger.Log(fmt.Sprintf("Finding season mappings for TVDB ID %d", tvdbID), types.LogOptions{
		Level:  types.Debug,
		Prefix: "TVDB",
	})

	// Use our database function to find all mappings with the same TVDB ID
	mappings, err := database.GetAnimeMappingsByTVDBID(tvdbID)
	if err != nil {
		return nil, fmt.Errorf("failed to get season mappings: %w", err)
	}

	if len(mappings) == 0 {
		logger.Log(fmt.Sprintf("No season mappings found for TVDB ID %d", tvdbID), types.LogOptions{
			Level:  types.Debug,
			Prefix: "TVDB",
		})
	} else {
		logger.Log(fmt.Sprintf("Found %d season mappings for TVDB ID %d", len(mappings), tvdbID), types.LogOptions{
			Level:  types.Info,
			Prefix: "TVDB",
		})
	}

	return mappings, nil
}
