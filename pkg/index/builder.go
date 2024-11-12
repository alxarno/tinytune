package index

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alxarno/tinytune/pkg/preview"
	"golang.org/x/sync/semaphore"
)

var ErrSemaphoreAcquire = errors.New("failed to acquire the semaphore")

type (
	IDGenerator      func(m FileMeta) (string, error)
	PreviewGenerator interface {
		Pull(path string) (preview.Data, error)
		ContentType(path string) int
	}
)

type FileMeta interface {
	Path() string
	RelativePath() string
	Name() string
	ModTime() time.Time
	IsDir() bool
	Size() int64
}

type indexBuilderParams struct {
	preview          PreviewGenerator
	id               IDGenerator
	files            []FileMeta
	progress         func()
	newFiles         func()
	workers          int
	maxNewImageItems int64
	maxNewVideoItems int64
	imageProcessing  bool
	videoProcessing  bool
}

type indexBuilder struct {
	index  *Index
	params indexBuilderParams
}

func newBuilder(index *Index) indexBuilder {
	return indexBuilder{index: index, params: indexBuilderParams{
		files:            []FileMeta{},
		progress:         func() {},
		newFiles:         func() {},
		workers:          1,
		maxNewImageItems: -1,
		maxNewVideoItems: -1,
		imageProcessing:  true,
		videoProcessing:  true,
		id:               func(m FileMeta) (string, error) { return m.Path(), nil },
	}}
}

func (ib *indexBuilder) run(ctx context.Context, r io.Reader) error {
	if err := ib.index.Decode(r); err != nil {
		if errors.Is(err, io.EOF) {
			slog.Warn("The index file could not be fully read, it may be corrupted or empty")
		} else {
			return err
		}
	}

	if err := ib.loadFiles(ctx); err != nil {
		return err
	}

	if err := ib.loadTree(); err != nil {
		return err
	}

	if err := ib.loadPaths(); err != nil {
		return err
	}

	return nil
}

func ifMaxPass(maxNewItems *int64) bool {
	if atomic.LoadInt64(maxNewItems) == -1 {
		return true
	}

	if atomic.LoadInt64(maxNewItems) == 0 {
		return false
	}

	atomic.AddInt64(maxNewItems, -1)

	return true
}

func (ib *indexBuilder) filePass(file FileMeta) (string, error) {
	id, err := ib.params.id(file)
	if err != nil {
		return "", err
	}

	if _, ok := ib.index.meta[id]; ok {
		return "", nil
	}

	contentType := ib.params.preview.ContentType(file.Path())
	if contentType == ContentTypeVideo && ib.params.videoProcessing && !ifMaxPass(&ib.params.maxNewVideoItems) {
		return "", nil
	}

	if contentType == ContentTypeImage && ib.params.imageProcessing && !ifMaxPass(&ib.params.maxNewImageItems) {
		return "", nil
	}

	return id, nil
}

func (ib *indexBuilder) loadFiles(ctx context.Context) error {
	waitGroup := new(sync.WaitGroup)
	sem := semaphore.NewWeighted(int64(ib.params.workers))
	resultChannel := make(chan fileProcessorResult, len(ib.params.files))
	processor := newFileProcessor(
		withPreview(ib.params.preview),
		withChan(resultChannel),
		withSemaphore(sem),
		withWaitGroup(waitGroup))

	for _, file := range ib.params.files {
		id, err := ib.filePass(file)
		if err != nil {
			return err
		}

		if id == "" {
			ib.params.progress()

			continue
		}

		if err := sem.Acquire(ctx, 1); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return fmt.Errorf("%w:%w", ErrSemaphoreAcquire, err)
		}

		waitGroup.Add(1)
		ib.params.progress()

		go processor.run(file, id)
	}

	waitGroup.Wait()
	close(resultChannel)

	for result := range resultChannel {
		ib.params.newFiles()

		if result.meta.Preview.Length != 0 {
			result.meta.Preview.Offset = uint32(len(ib.index.data))
			ib.index.data = append(ib.index.data, result.data...)
		}

		if result.meta.IsDir {
			result.meta.Type = ContentTypeDir
		}

		ib.index.meta[result.meta.ID] = result.meta
		ib.index.outDated = true
	}

	return nil
}

func (ib *indexBuilder) loadTree() error {
	ib.index.tree["root"] = make([]*Meta, 0)
	for _, meta := range ib.index.meta {
		if filepath.Dir(meta.RelativePath) == "." {
			ib.index.tree["root"] = append(ib.index.tree["root"], meta)
		}

		if !meta.IsDir {
			continue
		}

		for _, possibleChild := range ib.index.meta {
			if meta.RelativePath != filepath.Dir(possibleChild.RelativePath) {
				continue
			}

			if _, ok := ib.index.tree[meta.ID]; !ok {
				ib.index.tree[meta.ID] = make([]*Meta, 0)
			}

			ib.index.tree[meta.ID] = append(ib.index.tree[meta.ID], possibleChild)
		}
	}

	return nil
}

func (ib *indexBuilder) loadPaths() error {
	for _, v := range ib.index.meta {
		ib.index.paths[v.RelativePath] = v
	}

	return nil
}
