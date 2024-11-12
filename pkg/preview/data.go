package preview

import "time"

type Data struct {
	Duration    time.Duration
	ContentType int
	Resolution  string
	Data        []byte
}
