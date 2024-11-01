package internal

import (
	"errors"
	"log"
	"os"
	"path/filepath"
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
	path string
}

func (cosf CrawlerOSFile) Path() string {
	return cosf.path
}

type CrawlerOS struct {
	path string
}

func NewCrawlerOS(path string) CrawlerOS {
	return CrawlerOS{path}
}

func (c CrawlerOS) Scan() ([]FileMeta, error) {
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
			files = append(files, &CrawlerOSFile{info, path})
			return nil
		})
	if err != nil {
		log.Println(err)
	}
	return files, nil
}
