package internal

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"slices"
	"time"
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

func (c CrawlerOS) Scan(exclude ...string) ([]FileMeta, error) {
	fileInfo, err := os.Stat(c.path)
	if err != nil {
		return nil, err
	}
	if !fileInfo.IsDir() {
		return nil, errors.New("path is not dir")
	}
	files := []FileMeta{}
	err = filepath.Walk(c.path,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if path == c.path {
				return nil
			}
			if slices.Contains(exclude, path) {
				return nil
			}
			relativePath, err := filepath.Rel(c.path, path)
			if err != nil {
				return err
			}
			files = append(files, &CrawlerOSFile{info, path, relativePath})
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return files, nil
}
