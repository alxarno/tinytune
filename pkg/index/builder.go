package index

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"slices"
	"sync"

	"github.com/alxarno/tinytune/pkg/preview"
	"golang.org/x/sync/semaphore"
)

var (
	ErrSemaphoreAcquire = errors.New("failed to acquire the semaphore")
	ErrFileLoad         = errors.New("failed to load file")
	ErrPreviewPull      = errors.New("failed to pull preview")
)

type PreviewGenerator interface {
	Pull(item preview.Source) (preview.Data, error)
}

type indexBuilderParams struct {
	preview  PreviewGenerator
	files    []FileMeta
	progress func()
	newFiles func()
	workers  int
}

type indexBuilder struct {
	index  *Index
	params indexBuilderParams
}

func newBuilder(index *Index) indexBuilder {
	return indexBuilder{index: index, params: indexBuilderParams{
		files:    []FileMeta{},
		progress: func() {},
		newFiles: func() {},
		workers:  1,
	}}
}

func (ib *indexBuilder) run(ctx context.Context, r io.Reader) error {
	if err := ib.index.Decode(r); err != nil {
		if !errors.Is(err, io.EOF) {
			return err
		}

		slog.Warn("The index file could not be fully read, it may be corrupted or empty")
	}

	if err := ib.loadPaths(); err != nil {
		return err
	}

	if err := ib.loadFiles(ctx); err != nil {
		return err
	}

	if err := ib.loadTree(); err != nil {
		return err
	}

	// load second time for new meta items
	if err := ib.loadPaths(); err != nil {
		return err
	}

	return nil
}

type loadedFile struct {
	meta *Meta
	data []byte
}

//nolint:cyclop
func (ib *indexBuilder) loadFile(
	ctx context.Context,
	wg *sync.WaitGroup,
	sem *semaphore.Weighted,
	file FileMeta,
	dst chan loadedFile,
) error {
	metaItem := metaByFile(file)

	// if item already in map, but without preview -> create preview
	shouldSkipPreview := metaItem.IsDir || metaItem.IsOtherFile()
	if saved, ok := ib.index.meta[metaItem.ID]; ok && (saved.Preview.Length != 0 || shouldSkipPreview) {
		return nil
	}

	// needs for check if file/folder was modified, and remove old version
	if oldMeta, ok := ib.index.paths[metaItem.RelativePath]; ok {
		delete(ib.index.meta, oldMeta.ID)
	}

	if ib.params.preview == nil || metaItem.IsDir {
		dst <- loadedFile{metaItem, nil}

		return nil
	}

	if err := sem.Acquire(ctx, 1); err != nil {
		if errors.Is(err, context.Canceled) {
			return nil
		}

		return fmt.Errorf("%w: %w", ErrSemaphoreAcquire, err)
	}

	wg.Add(1)

	go func() {
		defer sem.Release(1)
		defer wg.Done()

		preview, err := ib.params.preview.Pull(metaItem)
		if err != nil {
			slog.Error(fmt.Errorf("%w: %w", ErrPreviewPull, err).Error())
			dst <- loadedFile{metaItem, nil}

			return
		}

		metaItem.Duration = preview.Duration()
		metaItem.Resolution = preview.Resolution()
		metaItem.Preview = PreviewLocation{
			Length: uint32(len(preview.Data())),
		}

		dst <- loadedFile{metaItem, preview.Data()}
	}()

	return nil
}

func (ib *indexBuilder) loadFiles(ctx context.Context) error {
	wg := new(sync.WaitGroup)
	sem := semaphore.NewWeighted(int64(ib.params.workers))
	results := make(chan loadedFile, len(ib.params.files))

	// pass biggest files first
	slices.SortStableFunc(ib.params.files, compareFileMetaSize)

	for _, file := range ib.params.files {
		ib.params.progress()

		err := ib.loadFile(ctx, wg, sem, file, results)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrFileLoad, err)
		}
	}

	wg.Wait()
	close(results)

	for result := range results {
		ib.params.newFiles()

		if result.meta.Preview.Length != 0 {
			result.meta.Preview.Offset = uint32(len(ib.index.data))
			ib.index.data = append(ib.index.data, result.data...)
		}

		ib.index.meta[result.meta.ID] = result.meta
		ib.index.outDated = true
	}

	return nil
}

func (ib *indexBuilder) loadTree() error {
	upperDir := func(path RelativePath) RelativePath {
		return RelativePath(filepath.Dir(string(path)))
	}
	rootChildren := make([]*Meta, 0)

	for _, subRoot := range ib.index.meta {
		if upperDir(subRoot.RelativePath) == "." {
			rootChildren = append(rootChildren, subRoot)
		}

		if !subRoot.IsDir {
			continue
		}

		subRootChildren := make([]*Meta, 0)

		for _, possibleChild := range ib.index.meta {
			if subRoot.RelativePath == upperDir(possibleChild.RelativePath) {
				subRootChildren = append(subRootChildren, possibleChild)
			}
		}

		if len(subRootChildren) != 0 {
			ib.index.tree[subRoot.ID] = subRootChildren
		}
	}

	ib.index.tree["root"] = rootChildren

	return nil
}

func (ib *indexBuilder) loadPaths() error {
	ib.index.paths = map[RelativePath]*Meta{}
	for _, v := range ib.index.meta {
		ib.index.paths[v.RelativePath] = v
	}

	return nil
}
