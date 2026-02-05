package repositories

import (
	"errors"
	"metachan/database"
	"metachan/entities"
	"metachan/utils/logger"

	"gorm.io/gorm/clause"
)

func CreateOrUpdateSimpleImage(image *entities.SimpleImage) (uint, error) {
	result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "image_url"}},
		DoUpdates: clause.AssignmentColumns([]string{"image_url"}),
	}).Create(image)

	if result.Error != nil {
		logger.Errorf("Meta", "Failed to create or update image: %v", result.Error)
		return 0, errors.New("failed to create or update image")
	}

	return image.ID, nil
}

func CreateOrUpdateSimpleTitle(title *entities.SimpleTitle) (uint, error) {
	result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "title"}},
		DoUpdates: clause.AssignmentColumns([]string{"type", "title"}),
	}).Create(title)

	if result.Error != nil {
		logger.Errorf("Meta", "Failed to create or update title: %v", result.Error)
		return 0, errors.New("failed to create or update title")
	}

	return title.ID, nil
}

func CreateOrUpdateExternalURL(url *entities.ExternalURL) (uint, error) {
	result := database.DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "url"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "url"}),
	}).Create(url)

	if result.Error != nil {
		logger.Errorf("Meta", "Failed to create or update external URL: %v", result.Error)
		return 0, errors.New("failed to create or update external URL")
	}

	return url.ID, nil
}
