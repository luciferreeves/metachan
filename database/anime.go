package database

import "metachan/entities"

func GetAnimeMappingViaMALID(malID int) (*entities.AnimeMapping, error) {
	var mapping entities.AnimeMapping
	if err := DB.Where("mal = ?", malID).First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}

// GetAnimeMappingsByTVDBID retrieves all anime mappings that share the same TVDB ID
func GetAnimeMappingsByTVDBID(tvdbID int) ([]entities.AnimeMapping, error) {
	var mappings []entities.AnimeMapping
	if err := DB.Where("tvdb = ?", tvdbID).Find(&mappings).Error; err != nil {
		return nil, err
	}
	return mappings, nil
}
