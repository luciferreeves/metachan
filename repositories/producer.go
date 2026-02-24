package repositories

import (
	"errors"
	"metachan/entities"
	"metachan/utils/logger"
	"time"

	"gorm.io/gorm/clause"
)

func CreateOrUpdateProducer(producer *entities.Producer) error {
	for i := range producer.Titles {
		t := &producer.Titles[i]
		DB.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "type"}, {Name: "title"}},
			DoNothing: true,
		}).Create(t)
		if t.ID == 0 {
			DB.Where("type = ? AND title = ?", t.Type, t.Title).First(t)
		}
	}

	result := DB.Clauses(clause.OnConflict{
		Columns: []clause.Column{{Name: "mal_id"}},
		DoUpdates: clause.AssignmentColumns([]string{
			"url", "favorites", "count", "established", "about", "image_id",
		}),
	}).Omit("Titles").Create(producer)

	if result.Error != nil {
		logger.Errorf("Producer", "Failed to create or update producer: %v", result.Error)
		return errors.New("failed to create or update producer")
	}

	if len(producer.Titles) > 0 {
		if err := DB.Model(producer).Association("Titles").Replace(producer.Titles); err != nil {
			logger.Errorf("Producer", "Failed to associate titles for producer %d: %v", producer.MALID, err)
		}
	}

	return nil
}

func GetAllProducers() ([]entities.Producer, error) {
	var producers []entities.Producer
	if err := DB.Select("id, mal_id, enriched_at").Find(&producers).Error; err != nil {
		return nil, err
	}
	return producers, nil
}

func GetProducerExternalURLCount(producer *entities.Producer) int64 {
	return DB.Model(producer).Association("ExternalURLs").Count()
}

func ReplaceProducerExternalURLs(producer *entities.Producer, urls []entities.ExternalURL) error {
	return DB.Model(producer).Association("ExternalURLs").Replace(urls)
}

func UpdateProducerDetails(id uint, url, established, about string, favorites, count int, imageURL string) error {
	updates := map[string]interface{}{
		"url":         url,
		"favorites":   favorites,
		"count":       count,
		"established": established,
		"about":       about,
	}
	if imageURL != "" {
		img := entities.SimpleImage{ImageURL: imageURL}
		imgID, err := CreateOrUpdateSimpleImage(&img)
		if err == nil && imgID != 0 {
			updates["image_id"] = imgID
		}
	}
	return DB.Model(&entities.Producer{}).Where("id = ?", id).Updates(updates).Error
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

func SetProducerEnriched(id uint) error {
	now := time.Now()
	return DB.Model(&entities.Producer{}).Where("id = ?", id).Update("enriched_at", now).Error
}
