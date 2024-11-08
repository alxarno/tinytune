package index

import "time"

const (
	_ = iota
	ContentTypeVideo
	ContentTypeImage
	ContentTypeOther
	ContentTypeDir
)

type IndexMeta struct {
	ID           string
	Path         string
	RelativePath string
	Name         string
	ModTime      time.Time
	IsDir        bool
	Preview      IndexMetaPreview
	Duration     time.Duration
	Resolution   string
	Type         int
}

func (m IndexMeta) IsImage() bool {
	return m.Type == ContentTypeImage
}

func (m IndexMeta) IsVideo() bool {
	return m.Type == ContentTypeVideo
}

func (m IndexMeta) IsOtherFile() bool {
	return m.Type == ContentTypeOther
}

type IndexMetaPreview struct {
	Length uint32
	Offset uint32
}
