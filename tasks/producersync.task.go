package tasks

import (
	"metachan/entities"
	"metachan/repositories"
	"metachan/types"
	"metachan/utils/api/jikan"
	"metachan/utils/logger"
	"time"
)

// ResumeProducerEnrichment is called on startup to resume any background enrichment
// that was interrupted by a previous shutdown.
func ResumeProducerEnrichment() {
	go enrichProducersInBackground()
}

func ProducerSync() error {
	logger.Infof("ProducerSync", "Starting producer sync (includes studios and licensors)")

	response, err := jikan.GetAnimeProducers()
	if err != nil {
		logger.Errorf("ProducerSync", "Failed to fetch producers: %v", err)
		return err
	}

	total := len(response.Data)
	logger.Infof("ProducerSync", "Fetched %d producers from MAL", total)

	if err := saveProducerListData(response.Data); err != nil {
		return err
	}

	logger.Successf("ProducerSync", "Saved basic data for %d producers, enriching external URLs in background", total)

	go enrichProducersInBackground()

	return nil
}

func saveProducerListData(producersData []types.JikanSingleProducer) error {
	total := len(producersData)
	const batchSize = 50

	for batchStart := 0; batchStart < total; batchStart += batchSize {
		batchEnd := min(batchStart+batchSize, total)
		batch := producersData[batchStart:batchEnd]

		type producerWithImage struct {
			producer entities.Producer
			imageURL string
		}

		producerItems := make([]producerWithImage, 0, len(batch))
		imageMap := make(map[string]struct{})

		for _, pd := range batch {
			producer := entities.Producer{
				MALID:       pd.MALID,
				URL:         pd.URL,
				Favorites:   pd.Favorites,
				Count:       pd.Count,
				Established: pd.Established,
				About:       pd.About,
			}
			for _, t := range pd.Titles {
				producer.Titles = append(producer.Titles, entities.SimpleTitle{
					Type:  t.Type,
					Title: t.Title,
				})
			}
			imageURL := pd.Images.JPG.ImageURL
			if imageURL != "" {
				imageMap[imageURL] = struct{}{}
			}
			producerItems = append(producerItems, producerWithImage{producer: producer, imageURL: imageURL})
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

		titleMap := make(map[string]entities.SimpleTitle)
		for _, pd := range producerItems {
			for _, t := range pd.producer.Titles {
				key := t.Type + ":" + t.Title
				titleMap[key] = t
			}
		}
		if len(titleMap) > 0 {
			titles := make([]entities.SimpleTitle, 0, len(titleMap))
			for _, t := range titleMap {
				titles = append(titles, t)
			}
			if err := repositories.BatchCreateSimpleTitles(titles); err != nil {
				logger.Errorf("ProducerSync", "Failed to batch insert titles: %v", err)
				return err
			}
		}

		imageIDMap, err := repositories.GetAllImagesMapped()
		if err != nil {
			logger.Errorf("ProducerSync", "Failed to query images: %v", err)
			return err
		}

		titleIDMap, err := repositories.GetAllTitlesMapped()
		if err != nil {
			logger.Errorf("ProducerSync", "Failed to query titles: %v", err)
			return err
		}

		producers := make([]entities.Producer, 0, len(producerItems))
		for _, pd := range producerItems {
			if pd.imageURL != "" {
				if id, ok := imageIDMap[pd.imageURL]; ok {
					pd.producer.ImageID = &id
				}
			}
			for i := range pd.producer.Titles {
				key := pd.producer.Titles[i].Type + ":" + pd.producer.Titles[i].Title
				if id, ok := titleIDMap[key]; ok {
					pd.producer.Titles[i].ID = id
				}
			}
			producers = append(producers, pd.producer)
		}

		if len(producers) > 0 {
			if err := repositories.BatchCreateProducers(producers); err != nil {
				logger.Errorf("ProducerSync", "Failed to batch insert producers: %v", err)
				return err
			}
			logger.Infof("ProducerSync", "Saved batch %d–%d of %d", batchStart+1, batchEnd, total)
		}
	}

	return nil
}

func enrichProducersInBackground() {
	producers, err := repositories.GetAllProducers()
	if err != nil {
		logger.Errorf("ProducerSync", "Failed to load producers for enrichment: %v", err)
		return
	}

	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	// Pre-check: skip entire run if nothing qualifies
	hasWork := false
	for _, p := range producers {
		if p.EnrichedAt == nil || !p.EnrichedAt.After(sevenDaysAgo) {
			hasWork = true
			break
		}
	}
	if !hasWork {
		return
	}

	logger.Infof("ProducerSync", "Resuming background enrichment for producers missing external URLs")

	total := len(producers)
	startTime := time.Now()
	enriched := 0

	for i, p := range producers {
		if p.EnrichedAt != nil && p.EnrichedAt.After(sevenDaysAgo) {
			continue
		}

		detail, err := jikan.GetProducerByID(p.MALID)
		if err != nil {
			logger.Warnf("ProducerSync", "Failed to fetch details for producer %d: %v", p.MALID, err)
			continue
		}

		data := detail.Data
		if err := repositories.UpdateProducerDetails(
			p.ID, data.URL, data.Established, data.About, data.Favorites, data.Count,
			data.Images.JPG.ImageURL,
		); err != nil {
			logger.Warnf("ProducerSync", "Failed to update details for producer %d: %v", p.MALID, err)
			continue
		}

		if len(data.External) > 0 {
			externalURLs := make([]entities.ExternalURL, 0, len(data.External))
			for _, ext := range data.External {
				externalURLs = append(externalURLs, entities.ExternalURL{Name: ext.Name, URL: ext.URL})
			}
			if err := repositories.ReplaceProducerExternalURLs(&p, externalURLs); err != nil {
				logger.Warnf("ProducerSync", "Failed to update external URLs for producer %d: %v", p.MALID, err)
			}
		}

		enriched++
		if (i+1)%10 == 0 || (i+1) == total {
			progress, eta := calculateProgress(i+1, total, startTime)
			logger.Infof("ProducerSync", "Enriching: %d/%d (%.1f%%) | ETA: %v", i+1, total, progress, eta)
		}
	}

	logger.Successf("ProducerSync", "Background enrichment complete. Enriched %d producers with external URLs", enriched)
}
