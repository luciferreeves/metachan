package database

import "metachan/entities"

func GetAnimeMappingViaMALID(malID int) (*entities.AnimeMapping, error) {
	var mapping entities.AnimeMapping
	if err := DB.Where("mal = ?", malID).First(&mapping).Error; err != nil {
		return nil, err
	}
	return &mapping, nil
}
