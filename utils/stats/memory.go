package stats

import (
	"fmt"
	"math"
	"metachan/types"
	"runtime"
	"time"
)

var startTime = time.Now()

func GetMemoryStats() types.MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	used := m.Alloc / 1024 // Total allocated memory in MiB
	total := m.Sys / 1024  // Total system memory obtained in MiB
	free := (m.Sys - m.Alloc) / 1024
	usage := math.Round(float64(m.Alloc)/float64(m.Sys)*100*100) / 100 // Memory usage percentage

	return types.MemoryStats{
		Used:  fmt.Sprintf("%d MiB", used),
		Total: fmt.Sprintf("%d MiB", total),
		Free:  fmt.Sprintf("%d MiB", free),
		Usage: fmt.Sprintf("%.2f%%", usage),
	}
}

func GetUptime() string {
	var timeString string
	uptime := time.Since(startTime)
	// return in format "2d 3h 4m 5s"
	days := int(uptime.Hours() / 24)
	hours := int(uptime.Hours()) % 24
	minutes := int(uptime.Minutes()) % 60
	seconds := int(uptime.Seconds()) % 60

	if days > 0 {
		timeString += fmt.Sprintf("%dd ", days)
	}
	if hours > 0 {
		timeString += fmt.Sprintf("%dh ", hours)
	}
	if minutes > 0 {
		timeString += fmt.Sprintf("%dm ", minutes)
	}
	if seconds > 0 {
		timeString += fmt.Sprintf("%ds ", seconds)
	}
	if timeString == "" {
		timeString = "0s"
	} else {
		timeString = timeString[:len(timeString)-1] // Remove trailing space
	}
	return timeString
}
