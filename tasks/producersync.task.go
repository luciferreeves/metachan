package tasks

import (
	"fmt"
	"metachan/database"
	"metachan/entities"
	"metachan/utils/api/jikan"
	"metachan/utils/logger"
	"time"
)

func ProducerSync() error {
	logger.Log("Starting producer sync (includes studios and licensors)...", logger.LogOptions{
		Level:  logger.Info,
		Prefix: "ProducerSync",
	})

	client := jikan.NewJikanClient()
	page := 1
	totalFetched := 0
	var totalPages int
	var totalProducers int
	startTime := time.Now()

	for {
		logger.Log(fmt.Sprintf("Fetching producers page %d...", page), logger.LogOptions{
			Level:  logger.Info,
			Prefix: "ProducerSync",
		})

		response, err := client.GetAnimeProducers(page)
		if err != nil {
			logger.Log(fmt.Sprintf("Failed to fetch producers page %d: %v", page, err), logger.LogOptions{
				Level:  logger.Error,
				Prefix: "ProducerSync",
			})
			// If we fetched at least one page, continue with what we have
			if page > 1 {
				break
			}
			return err
		}

		if len(response.Data) == 0 {
			break
		}

		// Set total pages from first response
		if page == 1 {
			totalPages = response.Pagination.LastVisiblePage
			totalProducers = totalPages * len(response.Data)
			logger.Log(fmt.Sprintf("Total pages: %d, Estimated producers: %d", totalPages, totalProducers), logger.LogOptions{
				Level:  logger.Info,
				Prefix: "ProducerSync",
			})
		}

		// Process each producer
		for _, producerData := range response.Data {
			// Check if producer already exists
			var existingProducer entities.Producer
			result := database.DB.Where("mal_id = ?", producerData.MALID).First(&existingProducer)

			if result.Error != nil {
				// Producer doesn't exist, create new
				producer := entities.Producer{
					MALID:       producerData.MALID,
					URL:         producerData.URL,
					Favorites:   producerData.Favorites,
					Count:       producerData.Count,
					Established: producerData.Established,
					About:       producerData.About,
				}

				// Create producer in database
				if err := database.DB.Create(&producer).Error; err != nil {
					logger.Log(fmt.Sprintf("Failed to create producer %d: %v", producerData.MALID, err), logger.LogOptions{
						Level:  logger.Error,
						Prefix: "ProducerSync",
					})
					continue
				}

				// Add titles
				for _, title := range producerData.Titles {
					producerTitle := entities.ProducerTitle{
						ProducerID: producer.ID,
						Type:       title.Type,
						Title:      title.Title,
					}
					if err := database.DB.Create(&producerTitle).Error; err != nil {
						logger.Log(fmt.Sprintf("Failed to create producer title for %d: %v", producerData.MALID, err), logger.LogOptions{
							Level:  logger.Error,
							Prefix: "ProducerSync",
						})
					}
				}

				// Add image
				if producerData.Images.JPG.ImageURL != "" {
					producerImage := entities.ProducerImage{
						ProducerID: producer.ID,
						ImageURL:   producerData.Images.JPG.ImageURL,
					}
					if err := database.DB.Create(&producerImage).Error; err != nil {
						logger.Log(fmt.Sprintf("Failed to create producer image for %d: %v", producerData.MALID, err), logger.LogOptions{
							Level:  logger.Error,
							Prefix: "ProducerSync",
						})
					}
				}

				// Fetch and add external URLs
				time.Sleep(350 * time.Millisecond) // Rate limiting
				externalResp, err := client.GetProducerExternal(producerData.MALID)
				if err != nil {
					logger.Log(fmt.Sprintf("Failed to fetch external URLs for producer %d: %v", producerData.MALID, err), logger.LogOptions{
						Level:  logger.Error,
						Prefix: "ProducerSync",
					})
				} else {
					for _, ext := range externalResp.Data {
						producerExt := entities.ProducerExternalURL{
							ProducerID: producer.ID,
							Name:       ext.Name,
							URL:        ext.URL,
						}
						if err := database.DB.Create(&producerExt).Error; err != nil {
							logger.Log(fmt.Sprintf("Failed to create external URL for producer %d: %v", producerData.MALID, err), logger.LogOptions{
								Level:  logger.Error,
								Prefix: "ProducerSync",
							})
						}
					}
				}

				// Get primary title (default or first available)
				primaryTitle := "Unknown"
				for _, title := range producerData.Titles {
					if title.Type == "Default" {
						primaryTitle = title.Title
						break
					}
				}
				if primaryTitle == "Unknown" && len(producerData.Titles) > 0 {
					primaryTitle = producerData.Titles[0].Title
				}

				logger.Log(fmt.Sprintf("Created producer: %s (ID: %d, Count: %d)", primaryTitle, producerData.MALID, producerData.Count), logger.LogOptions{
					Level:  logger.Success,
					Prefix: "ProducerSync",
				})
			} else {
				// Producer exists, update it
				existingProducer.URL = producerData.URL
				existingProducer.Favorites = producerData.Favorites
				existingProducer.Count = producerData.Count
				existingProducer.Established = producerData.Established
				existingProducer.About = producerData.About

				if err := database.DB.Save(&existingProducer).Error; err != nil {
					logger.Log(fmt.Sprintf("Failed to update producer %d: %v", producerData.MALID, err), logger.LogOptions{
						Level:  logger.Error,
						Prefix: "ProducerSync",
					})
					continue
				}

				// Delete and recreate titles
				database.DB.Where("producer_id = ?", existingProducer.ID).Delete(&entities.ProducerTitle{})
				for _, title := range producerData.Titles {
					producerTitle := entities.ProducerTitle{
						ProducerID: existingProducer.ID,
						Type:       title.Type,
						Title:      title.Title,
					}
					database.DB.Create(&producerTitle)
				}

				// Update image
				var existingImage entities.ProducerImage
				if database.DB.Where("producer_id = ?", existingProducer.ID).First(&existingImage).Error == nil {
					existingImage.ImageURL = producerData.Images.JPG.ImageURL
					database.DB.Save(&existingImage)
				} else if producerData.Images.JPG.ImageURL != "" {
					producerImage := entities.ProducerImage{
						ProducerID: existingProducer.ID,
						ImageURL:   producerData.Images.JPG.ImageURL,
					}
					database.DB.Create(&producerImage)
				}

				// Update external URLs
				time.Sleep(350 * time.Millisecond) // Rate limiting
				externalResp, err := client.GetProducerExternal(producerData.MALID)
				if err != nil {
					logger.Log(fmt.Sprintf("Failed to fetch external URLs for producer %d: %v", producerData.MALID, err), logger.LogOptions{
						Level:  logger.Error,
						Prefix: "ProducerSync",
					})
				} else {
					database.DB.Where("producer_id = ?", existingProducer.ID).Delete(&entities.ProducerExternalURL{})
					for _, ext := range externalResp.Data {
						producerExt := entities.ProducerExternalURL{
							ProducerID: existingProducer.ID,
							Name:       ext.Name,
							URL:        ext.URL,
						}
						database.DB.Create(&producerExt)
					}
				}

				primaryTitle := "Unknown"
				for _, title := range producerData.Titles {
					if title.Type == "Default" {
						primaryTitle = title.Title
						break
					}
				}
				if primaryTitle == "Unknown" && len(producerData.Titles) > 0 {
					primaryTitle = producerData.Titles[0].Title
				}

				logger.Log(fmt.Sprintf("Updated producer: %s (ID: %d, Count: %d)", primaryTitle, producerData.MALID, producerData.Count), logger.LogOptions{
					Level:  logger.Success,
					Prefix: "ProducerSync",
				})
			}

			totalFetched++

			// Progress update every 10 producers
			if totalFetched%10 == 0 && totalProducers > 0 {
				progress := float64(totalFetched) / float64(totalProducers) * 100
				elapsed := time.Since(startTime)
				avgTimePerProducer := elapsed / time.Duration(totalFetched)
				remaining := totalProducers - totalFetched
				eta := avgTimePerProducer * time.Duration(remaining)

				logger.Log(fmt.Sprintf("Progress: %d/%d - %.1f%% | ETA: %v", totalFetched, totalProducers, progress, eta.Round(time.Second)), logger.LogOptions{
					Level:  logger.Info,
					Prefix: "ProducerSync",
				})
			}

			time.Sleep(350 * time.Millisecond) // Rate limiting between producers
		}

		// Check if there's more data
		if !response.Pagination.HasNextPage {
			break
		}

		page++
		time.Sleep(1 * time.Second) // Additional delay between pages
	}

	logger.Log(fmt.Sprintf("Producer sync completed successfully. Total: %d producers", totalFetched), logger.LogOptions{
		Level:  logger.Success,
		Prefix: "ProducerSync",
	})

	return nil
}
