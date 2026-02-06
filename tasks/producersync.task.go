package tasks

import (
	"metachan/entities"
	"metachan/repositories"
	"metachan/utils/api/jikan"
	"metachan/utils/logger"
	"time"
)

func ProducerSync() error {
	logger.Infof("ProducerSync", "Starting producer sync (includes studios and licensors)")

	response, err := jikan.GetAnimeProducers()
	if err != nil {
		logger.Errorf("ProducerSync", "Failed to fetch producers: %v", err)
		return err
	}

	total := len(response.Data)
	logger.Infof("ProducerSync", "Fetched %d producers from MAL", total)

	startTime := time.Now()

	for i, producerData := range response.Data {
		producerDetail, err := jikan.GetProducerByID(producerData.MALID)
		if err != nil {
			logger.Warnf("ProducerSync", "Failed to fetch details for producer %d: %v", producerData.MALID, err)
			continue
		}

		var imageID *uint
		if producerDetail.Data.Images.JPG.ImageURL != "" {
			image := entities.SimpleImage{
				ImageURL: producerDetail.Data.Images.JPG.ImageURL,
			}
			id, err := repositories.CreateOrUpdateSimpleImage(&image)
			if err == nil {
				imageID = &id
			}
		}

		producer := entities.Producer{
			MALID:       producerDetail.Data.MALID,
			URL:         producerDetail.Data.URL,
			Favorites:   producerDetail.Data.Favorites,
			Count:       producerDetail.Data.Count,
			Established: producerDetail.Data.Established,
			About:       producerDetail.Data.About,
			ImageID:     imageID,
		}

		for _, title := range producerDetail.Data.Titles {
			producer.Titles = append(producer.Titles, entities.SimpleTitle{
				Type:  title.Type,
				Title: title.Title,
			})
		}

		for _, ext := range producerDetail.Data.External {
			producer.ExternalURLs = append(producer.ExternalURLs, entities.ExternalURL{
				Name: ext.Name,
				URL:  ext.URL,
			})
		}

		if err := repositories.CreateOrUpdateProducer(&producer); err != nil {
			logger.Warnf("ProducerSync", "Failed to sync producer %d: %v", producerData.MALID, err)
			continue
		}

		progress, eta := calculateProgress(i+1, total, startTime)
		logger.Infof("ProducerSync", "Progress: %d/%d (%.1f%%) | ETA: %v", i+1, total, progress, eta)
	}

	logger.Successf("ProducerSync", "Producer sync completed. Total: %d producers", total)
	return nil
}
