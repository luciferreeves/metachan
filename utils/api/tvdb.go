package api

import (
	"fmt"
	"metachan/database"
	"metachan/entities"
	"metachan/types"
	"metachan/utils/logger"
)

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
