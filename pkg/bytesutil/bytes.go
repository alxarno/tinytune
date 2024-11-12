package bytesutil

import (
	"fmt"
	"math"
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
