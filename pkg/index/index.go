package index

import (
	"context"
	"crypto/rand"
	"encoding/gob"
	"fmt"
	"io"
	"sync"
	"time"
)

const (
	_ = iota
	ContentTypeVideo
	ContentTypeImage
	ContentTypeOther
)

type Index struct {
	meta     map[string]IndexMeta
	data     []byte
	outDated bool
	preview  func(string) (time.Duration, int, []byte, error)
	id       func(m FileMeta) (string, error)
	files    []FileMeta
	context  context.Context
	progress func()
	newFiles func()
}

type IndexOption func(*Index)

type IndexMeta struct {
	ID           string
	Path         string
	RelativePath string
	Name         string
	ModTime      time.Time
	IsDir        bool
	Preview      IndexMetaPreview
	Duration     time.Duration
	Type         int
}

type IndexMetaPreview struct {
	Length uint32
	Offset uint32
}

type FileMeta interface {
	Path() string
	RelativePath() string
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
		data:     defaultData,
		meta:     defaultMeta,
		preview:  defaultPreview,
		id:       defaultID,
		files:    defaultFiles,
		context:  context.Background(),
		outDated: false,
		progress: func() {},
		newFiles: func() {},
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
	return func(i *Index) {
		i.preview = f
	}
}

func WithID(f func(m FileMeta) (string, error)) IndexOption {
	return func(i *Index) {
		i.id = f
	}
}

func WithFiles(files []FileMeta) IndexOption {
	return func(i *Index) {
		i.files = files
	}
}

func WithContext(ctx context.Context) IndexOption {
	return func(i *Index) {
		i.context = ctx
	}
}

func WithProgress(f func()) IndexOption {
	return func(i *Index) {
		i.progress = f
	}
}

func WithNewFiles(f func()) IndexOption {
	return func(i *Index) {
		i.newFiles = f
	}
}

func (index Index) OutDated() bool {
	return index.outDated
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

func (index Index) FilesWithPreviewStat() (int, uint32) {
	count := 0
	size := uint32(0)
	for _, v := range index.meta {
		if v.Preview.Length != 0 {
			count++
			size += v.Preview.Length
		}
	}
	return count, size
}

type poolPreviewWorkerResult struct {
	meta IndexMeta
	data []byte
}

func (index Index) poolPreviewWorker(metaCh chan IndexMeta, result chan poolPreviewWorkerResult, wg *sync.WaitGroup) {
	defer wg.Done()
	for meta := range metaCh {
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
		select {
		case <-index.context.Done():
			return
		default:
		}
	}
}

func (index *Index) loadFiles() error {
	// logBar := bar(len(index.files), "Processing files for index...")
	metaChannel := make(chan IndexMeta)
	resultChannel := make(chan poolPreviewWorkerResult, len(index.files))
	wg := new(sync.WaitGroup)
	workers := 1
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go index.poolPreviewWorker(metaChannel, resultChannel, wg)
	}
	for _, p := range index.files {
		index.progress()
		id, err := index.id(p)
		if err != nil {
			return err
		}
		if _, ok := index.meta[id]; ok {
			continue
		}
		meta := IndexMeta{
			Path:         p.Path(),
			RelativePath: p.RelativePath(),
			Name:         p.Name(),
			ModTime:      p.ModTime(),
			IsDir:        p.IsDir(),
			ID:           id,
		}
		metaChannel <- meta
	}
	close(metaChannel)
	wg.Wait()
	close(resultChannel)
	for r := range resultChannel {
		index.newFiles()
		if r.meta.Preview.Length != 0 {
			r.meta.Preview.Offset = uint32(len(index.data))
			index.data = append(index.data, r.data...)
		}
		index.meta[r.meta.ID] = r.meta
		index.outDated = true
	}
	return nil
}
