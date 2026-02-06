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
	const batchSize = 10
	totalProcessed := 0

	for batchStart := 0; batchStart < total; batchStart += batchSize {
		batchEnd := min(batchStart+batchSize, total)

		batchData := response.Data[batchStart:batchEnd]

		type producerWithImage struct {
			producer entities.Producer
			imageURL string
		}

		producersData := make([]producerWithImage, 0, len(batchData))
		imageMap := make(map[string]struct{})

		for i, producerData := range batchData {
			producerDetail, err := jikan.GetProducerByID(producerData.MALID)
			if err != nil {
				logger.Warnf("ProducerSync", "Failed to fetch details for producer %d: %v", producerData.MALID, err)
				continue
			}

			producer := entities.Producer{
				MALID:       producerDetail.Data.MALID,
				URL:         producerDetail.Data.URL,
				Favorites:   producerDetail.Data.Favorites,
				Count:       producerDetail.Data.Count,
				Established: producerDetail.Data.Established,
				About:       producerDetail.Data.About,
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

			imageURL := producerDetail.Data.Images.JPG.ImageURL
			if imageURL != "" {
				imageMap[imageURL] = struct{}{}
			}

			producersData = append(producersData, producerWithImage{
				producer: producer,
				imageURL: imageURL,
			})

			if (batchStart+i+1)%10 == 0 || (batchStart+i+1) == total {
				progress, eta := calculateProgress(batchStart+i+1, total, startTime)
				logger.Infof("ProducerSync", "Fetched: %d/%d (%.1f%%) | ETA: %v", batchStart+i+1, total, progress, eta)
			}
		}

		if len(imageMap) > 0 {
			images := make([]entities.SimpleImage, 0, len(imageMap))
			for url := range imageMap {
				images = append(images, entities.SimpleImage{ImageURL: url})
			}
			if err := repositories.BatchCreateSimpleImages(images); err != nil {
				logger.Errorf("ProducerSync", "Failed to batch insert images: %v", err)
				return err
			}
		}

		var dbImages []entities.SimpleImage
		if err := repositories.DB.Select("id, image_url").Find(&dbImages).Error; err != nil {
			logger.Errorf("ProducerSync", "Failed to query images: %v", err)
			return err
		}

		imageIDMap := make(map[string]uint)
		for _, img := range dbImages {
			imageIDMap[img.ImageURL] = img.ID
		}

		producers := make([]entities.Producer, 0, len(producersData))
		for _, pd := range producersData {
			if pd.imageURL != "" {
				if id, exists := imageIDMap[pd.imageURL]; exists {
					pd.producer.ImageID = &id
				}
			}
			producers = append(producers, pd.producer)
		}

		if len(producers) > 0 {
			if err := repositories.BatchCreateProducers(producers); err != nil {
				logger.Errorf("ProducerSync", "Failed to batch insert producers: %v", err)
				return err
			}
			totalProcessed += len(producers)
			logger.Infof("ProducerSync", "Committed batch: %d producers (Total: %d/%d)", len(producers), totalProcessed, total)
		}
	}

	logger.Successf("ProducerSync", "Producer sync completed. Total: %d producers", totalProcessed)
	return nil
}
