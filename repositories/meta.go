package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/utils/logger"

	"gorm.io/gorm/clause"
)

func CreateOrUpdateSimpleImage(image *entities.SimpleImage) (uint, error) {
	result := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "image_url"}},
		DoUpdates: clause.AssignmentColumns([]string{"image_url"}),
	}).Create(image)

	if result.Error != nil {
		logger.Errorf("Meta", "Failed to create or update image: %v", result.Error)
		return 0, errors.New("failed to create or update image")
	}

	return image.ID, nil
}

func BatchCreateSimpleImages(images []entities.SimpleImage) error {
	if len(images) == 0 {
		return nil
	}

	result := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "image_url"}},
		DoNothing: true,
	}).CreateInBatches(&images, 100)

	if result.Error != nil {
		logger.Errorf("Meta", "Failed to batch create images: %v", result.Error)
		return errors.New("failed to batch create images")
	}

	return nil
}

func BatchCreateSimpleTitles(titles []entities.SimpleTitle) error {
	if len(titles) == 0 {
		return nil
	}

	result := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "type"}, {Name: "title"}},
		DoNothing: true,
	}).CreateInBatches(&titles, 100)

	if result.Error != nil {
		logger.Errorf("Meta", "Failed to batch create titles: %v", result.Error)
		return errors.New("failed to batch create titles")
	}

	return nil
}

func CreateOrUpdateSimpleTitle(title *entities.SimpleTitle) (uint, error) {
	result := DB.Clauses(clause.OnConflict{
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
	result := DB.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "name"}, {Name: "url"}},
		DoUpdates: clause.AssignmentColumns([]string{"name", "url"}),
	}).Create(url)

	if result.Error != nil {
		logger.Errorf("Meta", "Failed to create or update external URL: %v", result.Error)
		return 0, errors.New("failed to create or update external URL")
	}

	return url.ID, nil
}
