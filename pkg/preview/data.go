package preview

import (
	"time"
)

type Data interface {
	Data() []byte
	Duration() time.Duration
	Resolution() (int, int)
}

type data struct {
	duration time.Duration
	width    int
	height   int
	data     []byte
}

func (d data) Duration() time.Duration {
	return d.duration
}

func (d data) Resolution() (int, int) {
	return d.width, d.height
}

func (d data) Data() []byte {
	return d.data
}
