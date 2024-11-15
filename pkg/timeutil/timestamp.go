package timeutil

import (
	"fmt"
	"time"
)

func String(duration time.Duration) string {
	result := ""
	timeBound := 60
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) - hours*timeBound
	seconds := int(duration.Seconds()) - minutes*timeBound - hours*timeBound*timeBound

	if hours > 0 {
		result += fmt.Sprintf("%02d:", hours)
	}

	result += fmt.Sprintf("%02d:", minutes)
	result += fmt.Sprintf("%02d", seconds)

	return result
}
