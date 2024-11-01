package internal

import (
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"
)

type Index struct {
	meta    map[string]IndexMeta
	data    []byte
	preview func(string) (time.Duration, int, []byte, error)
	id      func(m FileMeta) (string, error)
	files   []FileMeta
}

type IndexOption func(*Index)

type IndexMeta struct {
	Path     string
	Name     string
	ModTime  time.Time
	Hash     string
	IsDir    bool
	Preview  IndexMetaPreview
	Duration time.Duration
	Type     int
}

type IndexMetaPreview struct {
	Length uint32
	Offset uint32
}

type FileMeta interface {
	Path() string
	Name() string
	ModTime() time.Time
	IsDir() bool
	Size() int64
}

func NewIndex(r io.Reader, opts ...IndexOption) Index {
	gob.Register(IndexMeta{})
	var (
		defaultData    = []byte{}
		defaultMeta    = map[string]IndexMeta{}
		defaultFiles   = []FileMeta{}
		defaultPreview = func(string) (time.Duration, int, []byte, error) { return 0, ContentTypeOther, nil, nil }
		defaultID      = func(m FileMeta) (string, error) {
			buff := make([]byte, 100)
			rand.Read(buff)
			return string(buff), nil
		}
	)
	index := Index{
		data:    defaultData,
		meta:    defaultMeta,
		preview: defaultPreview,
		id:      defaultID,
		files:   defaultFiles,
	}
	for _, opt := range opts {
		opt(&index)
	}
	if r != nil {
		index.Decode(r)
	}
	if len(index.files) != 0 {
		index.loadFiles()
	}
	return index
}

func WithPreview(f func(path string) (time.Duration, int, []byte, error)) IndexOption {
	return func(h *Index) {
		h.preview = f
	}
}

func WithID(f func(m FileMeta) (string, error)) IndexOption {
	return func(h *Index) {
		h.id = f
	}
}

func WithFiles(files []FileMeta) IndexOption {
	return func(h *Index) {
		h.files = files
	}
}

func (index Index) PullPreview(hash string) ([]byte, error) {
	m, ok := index.meta[hash]
	if !ok {
		return nil, fmt.Errorf("not found")
	}
	if m.Preview.Length == 0 {
		return nil, nil
	}
	return index.data[m.Preview.Offset : m.Preview.Offset+m.Preview.Length], nil
}

type poolPreviewWorkerResult struct {
	meta IndexMeta
	data []byte
}

func (index Index) poolPreviewWorker(metaCh chan IndexMeta, result chan poolPreviewWorkerResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for meta := range metaCh {
		if meta.IsDir {
			result <- poolPreviewWorkerResult{meta, nil}
			continue
		}
		duration, t, data, err := index.preview(meta.Path)
		if err != nil {
			fmt.Println(err)
			continue
		}
		meta.Duration = duration
		meta.Type = t
		meta.Preview = IndexMetaPreview{
			Length: uint32(len(data)),
		}
		result <- poolPreviewWorkerResult{meta, data}
	}
}

func (index *Index) loadFiles() error {
	pathCh := make(chan IndexMeta)
	resultCh := make(chan poolPreviewWorkerResult, len(index.files))
	wg := new(sync.WaitGroup)
	workers := 4
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go index.poolPreviewWorker(pathCh, resultCh, wg)
	}
	for _, p := range index.files {
		hash, err := index.id(p)
		if err != nil {
			return err
		}
		meta := IndexMeta{
			Path:    p.Path(),
			Name:    p.Name(),
			ModTime: p.ModTime(),
			IsDir:   p.IsDir(),
			Hash:    hash,
		}
		pathCh <- meta
	}
	close(pathCh)
	wg.Wait()
	close(resultCh)
	for r := range resultCh {
		if r.meta.Preview.Length != 0 {
			r.meta.Preview.Offset = uint32(len(index.data))
			index.data = append(index.data, r.data...)
		}
		index.meta[r.meta.Hash] = r.meta
	}
	return nil

}
