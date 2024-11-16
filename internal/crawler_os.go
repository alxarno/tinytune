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

// type OutputSlice interface {
// 	[]filterItem | []index.FileMeta
// }

// func Convert[S []*CrawlerOSFile, T []filterItem | []index.FileMeta](src S, dst *T) *T {
// 	result := make(T, len(files))
// 	for i, v := range src {
// 		dst[i] = v
// 	}

// 	return dst

// }

// func Convert[T, B []*CrawlerOSFile | []filterItem | []index.FileMeta](files []T) B {
// 	result := make(B, len(files))
// 	for i, v := range files {
// 		result[i] = v
// 	}

// 	return result
// }

// func Convert[S ~[]J, J filterItem](files []*CrawlerOSFile) S {
// 	result := make(S, len(files))
// 	for i, v := range files {
// 		result[i] = v
// 	}

// 	return result
// }

// func Convert[J, T interface{}](src J) T {
// 	return

// }

// func FilesToMeta(files []*CrawlerOSFile) []index.FileMeta {
// 	result := make([]index.FileMeta, len(files))
// 	for i, v := range files {
// 		result[i] = v
// 	}
// 	_ = index.FileMeta(files[1])
// 	return result
// }

// func FilesToFilter(files []*CrawlerOSFile) []filterItem {
// 	result := make([]filterItem, len(files))
// 	for i, v := range files {
// 		result[i] = v
// 	}

// 	return result
// }
