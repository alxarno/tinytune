package preview

import "time"

type Data interface {
	Data() []byte
	Duration() time.Duration
	Resolution() string
}

type data struct {
	duration   time.Duration
	resolution string
	data       []byte
}

func (d data) Duration() time.Duration {
	return d.duration
}

func (d data) Resolution() string {
	return d.resolution
}

func (d data) Data() []byte {
	return d.data
}
