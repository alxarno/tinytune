package index

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"slices"
	"time"
)

const (
	_ = iota
	ContentTypeVideo
	ContentTypeImage
	ContentTypeOther
	ContentTypeDir
)

type ID string
type Path string
type RelativePath string

type Meta struct {
	ID           ID
	AbsolutePath Path
	RelativePath RelativePath
	Name         string
	ModTime      time.Time
	IsDir        bool
	Preview      PreviewLocation
	Duration     time.Duration
	Resolution   string
	Type         int
}

type PreviewLocation struct {
	Length uint32
	Offset uint32
}

func (m *Meta) IsImage() bool {
	return m.Type == ContentTypeImage
}

func (m *Meta) IsVideo() bool {
	return m.Type == ContentTypeVideo
}

func (m *Meta) IsOtherFile() bool {
	return m.Type == ContentTypeOther
}

func (m *Meta) Path() string {
	return string(m.AbsolutePath)
}

func (m *Meta) generateID() {
	idSource := []byte(fmt.Sprintf("%s%s", m.RelativePath, m.ModTime))
	fileID := sha256.Sum256(idSource)
	m.ID = ID(hex.EncodeToString(fileID[:5]))
}

func (m *Meta) setContentType() {
	if m.IsDir {
		m.Type = ContentTypeDir

		return
	}

	//nolint:lll
	videoFormats := []string{"3gp", "avi", "f4v", "flv", "gif", "hevc", "m4v", "mlv", "mov", "mp4", "m4a", "3g2", "mj2", "mpeg", "ogv", "webm"}
	imageFormats := []string{"jpeg", "png", "jpg", "webp", "bmp"}

	ext := filepath.Ext(string(m.AbsolutePath))
	minExtensionLength := 2

	if len(ext) < minExtensionLength {
		m.Type = ContentTypeOther

		return
	}

	if slices.Contains(imageFormats, ext[1:]) {
		m.Type = ContentTypeImage
	}

	if slices.Contains(videoFormats, ext[1:]) {
		m.Type = ContentTypeVideo
	}

	m.Type = ContentTypeOther
}
