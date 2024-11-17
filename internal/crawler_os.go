package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"time"

	"github.com/alxarno/tinytune/pkg/index"
)

var (
	ErrDirNotFound              = errors.New("directory not found")
	ErrDirStatFailed            = errors.New("failed os.Stat")
	ErrDirWalkHandlerFailed     = errors.New("failed filepath.Walk handler")
	ErrFileRelativePathNotFound = errors.New("file relative path not found")
	ErrDirWalkFailed            = errors.New("failed filepath.Walk")
)

type RawFile interface {
	Name() string
	ModTime() time.Time
	IsDir() bool
	Size() int64
}

type CrawlerOSFile struct {
	os.FileInfo
	path         string
	relativePath string
}

func (cosf CrawlerOSFile) Path() string {
	return cosf.path
}

func (cosf CrawlerOSFile) RelativePath() string {
	return cosf.relativePath
}

type CrawlerOS struct {
	path string
}

func NewCrawlerOS(path string) CrawlerOS {
	return CrawlerOS{path}
}

func (c CrawlerOS) Scan(exclude ...string) ([]index.FileMeta, error) {
	fileInfo, err := os.Stat(c.path)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDirStatFailed, err)
	}

	if !fileInfo.IsDir() {
		return nil, ErrDirNotFound
	}

	files := []index.FileMeta{}
	err = filepath.Walk(c.path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("%w: %w", ErrDirWalkHandlerFailed, err)
			}

			if path == c.path {
				return nil
			}

			if slices.Contains(exclude, path) {
				return nil
			}

			relativePath, err := filepath.Rel(c.path, path)
			if err != nil {
				return fmt.Errorf("%w: %w", ErrFileRelativePathNotFound, err)
			}

			files = append(files, &CrawlerOSFile{info, path, relativePath})

			return nil
		})

	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDirWalkFailed, err)
	}

	return files, nil
}
