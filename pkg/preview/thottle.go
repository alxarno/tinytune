package preview

import (
	"bufio"
	"context"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

type Memory struct {
	MemTotal     int
	MemFree      int
	MemAvailable int
}

func parseLine(raw string) (string, int) {
	text := strings.ReplaceAll(raw[:len(raw)-2], " ", "")
	keyValue := strings.Split(text, ":")

	return keyValue[0], toInt(keyValue[1])
}

func toInt(raw string) int {
	if raw == "" {
		return 0
	}

	res, err := strconv.Atoi(raw)
	if err != nil {
		panic(err)
	}

	return res
}

func readMemoryStats() Memory {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	bufio.NewScanner(file)
	scanner := bufio.NewScanner(file)
	res := Memory{}

	for scanner.Scan() {
		key, value := parseLine(scanner.Text())
		switch key {
		case "MemTotal":
			res.MemTotal = value
		case "MemFree":
			res.MemFree = value
		case "MemAvailable":
			res.MemAvailable = value
		}
	}

	return res
}

func isMemoryPassCheck(maxMemoryOccupied float64) bool {
	stats := readMemoryStats()
	memoryOccupied := float64(stats.MemFree) / float64(stats.MemTotal)

	return memoryOccupied < maxMemoryOccupied
}

func throttle(ctx context.Context, maxMemoryOccupied float64, maxWaitingTime time.Duration) error {
	if isMemoryPassCheck(maxMemoryOccupied) {
		return nil
	}

	ticker := time.NewTicker(time.Second)

	timeout, cancel := context.WithTimeout(ctx, maxWaitingTime)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return context.Canceled
		case <-timeout.Done():
			slog.Warn("waiting timeout", slog.String("source", "throttle"))

			return nil
		case <-ticker.C:
			if isMemoryPassCheck(maxMemoryOccupied) {
				return nil
			}
		}
	}
}
