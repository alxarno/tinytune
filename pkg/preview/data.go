package preview

import "time"

type Data struct {
	Duration    time.Duration
	ContentType int
	Resolution  string
	Stream      bool
	Data        []byte
}
