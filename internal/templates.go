package internal

import (
	"fmt"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

func extension(name string) string {
	extension := path.Ext(name)
	if extension != "" {
		return extension[1:]
	}

	return ""
}

func width(res string) string {
	return strings.Split(res, "x")[0]
}

func height(res string) string {
	return strings.Split(res, "x")[1]
}

func durationPrint(duration time.Duration) string {
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

func streaming(path string) bool {
	streamingFormats := []string{"avi", "f4v", "flv"}
	ext := filepath.Ext(path)
	minExtensionLength := 2

	if len(ext) < minExtensionLength {
		return false
	}

	if slices.Contains(streamingFormats, ext[1:]) {
		return true
	}

	return false
}
