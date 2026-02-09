package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/utils/logger"

	"gorm.io/gorm/clause"
)

func CreateOrUpdateProducer(producer *entities.Producer) error {
	result := DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "mal_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"url", "favorites", "count", "established", "about", "image_id",
		}),
	}).Create(producer)

	if result.Error != nil {
		logger.Errorf("Producer", "Failed to create or update producer: %v", result.Error)
		return errors.New("failed to create or update producer")
	}

	return nil
}

func BatchCreateProducers(producers []entities.Producer) error {
	if len(producers) == 0 {
		return nil
	}

	tx := DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	for i := range producers {
		result := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "mal_id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"url", "favorites", "count", "established", "about", "image_id",
			}),
		}).Omit("Titles", "ExternalURLs").Create(&producers[i])

		if result.Error != nil {
			tx.Rollback()
			logger.Errorf("Producer", "Failed to create producer %d: %v", producers[i].MALID, result.Error)
			return errors.New("failed to batch create producers")
		}

		if len(producers[i].Titles) > 0 {
			if err := tx.Model(&producers[i]).Association("Titles").Replace(producers[i].Titles); err != nil {
				tx.Rollback()
				logger.Errorf("Producer", "Failed to associate titles for producer %d: %v", producers[i].MALID, err)
				return errors.New("failed to batch create producers")
			}
		}

		if len(producers[i].ExternalURLs) > 0 {
			if err := tx.Model(&producers[i]).Association("ExternalURLs").Replace(producers[i].ExternalURLs); err != nil {
				tx.Rollback()
				logger.Errorf("Producer", "Failed to associate external URLs for producer %d: %v", producers[i].MALID, err)
				return errors.New("failed to batch create producers")
			}
		}
	}

	return tx.Commit().Error
}
