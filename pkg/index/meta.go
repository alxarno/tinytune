package index

import "time"

const (
	_ = iota
	ContentTypeVideo
	ContentTypeImage
	ContentTypeOther
	ContentTypeDir
)

type Meta struct {
	ID           string
	Path         string
	RelativePath string
	Name         string
	ModTime      time.Time
	IsDir        bool
	Preview      Preview
	Duration     time.Duration
	Resolution   string
	Type         int
}

func (m Meta) IsImage() bool {
	return m.Type == ContentTypeImage
}

func (m Meta) IsVideo() bool {
	return m.Type == ContentTypeVideo
}

func (m Meta) IsOtherFile() bool {
	return m.Type == ContentTypeOther
}

type Preview struct {
	Length uint32
	Offset uint32
}
