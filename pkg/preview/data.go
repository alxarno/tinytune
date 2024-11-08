package preview

import "time"

type PreviewData struct {
	Duration    time.Duration
	ContentType int
	Resolution  string
	Data        []byte
}
