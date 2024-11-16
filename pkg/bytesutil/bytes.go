package bytesutil

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func PrettyByteSize[V int64 | uint64 | uint32 | int](b V) string {
	byteSize := float64(b)
	kilo := 1024.0

	for _, unit := range []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi"} {
		if math.Abs(byteSize) < kilo {
			return fmt.Sprintf("%3.1f%sB", byteSize, unit)
		}

		byteSize /= kilo
	}

	return fmt.Sprintf("%.1fYiB", byteSize)
}

func ParseByteSize(size string) int64 {
	coefficients := []string{"", "KB", "MB", "GB", "TB", "PB", "EB", "ZB"}
	byteSize := int64(0)
	kilo := 1024.0
	re := regexp.MustCompile("[A-z]+")
	unit := strings.ToUpper(re.FindString(size))
	number, _ := strings.CutSuffix(size, unit)

	sizeValue, err := strconv.Atoi(number)
	if err != nil {
		panic(err)
	}

	for i, v := range coefficients {
		if v == unit || i == 0 && unit == "B" {
			byteSize = int64(sizeValue) * int64(math.Pow(kilo, float64(i)))
		}
	}

	return byteSize
}
