package tasks

import "time"

func calculateProgress(current, total int, startTime time.Time) (progress float64, eta time.Duration) {
	progress = float64(current) / float64(total) * 100
	elapsed := time.Since(startTime)
	avgTimePerItem := elapsed / time.Duration(current)
	remaining := total - current
	eta = avgTimePerItem * time.Duration(remaining)
	return progress, eta.Round(time.Second)
}
