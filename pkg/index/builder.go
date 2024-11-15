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
	ErrGetExcludedFiles = errors.New("failed to get excluded files")
)

type (
	PreviewGenerator func(item preview.Source) (preview.Data, error)
)

type indexBuilderParams struct {
	preview              PreviewGenerator
	files                []FileMeta
	progress             func()
	newFiles             func()
	workers              int
	maxImageProcessCount int64
	maxVideoProcessCount int64
	imageProcessing      bool
	videoProcessing      bool
	includePatterns      string
	excludePatterns      string
}

type indexBuilder struct {
	index  *Index
	params indexBuilderParams
}

func newBuilder(index *Index) indexBuilder {
	return indexBuilder{index: index, params: indexBuilderParams{
		files:                []FileMeta{},
		progress:             func() {},
		newFiles:             func() {},
		workers:              1,
		maxImageProcessCount: -1,
		maxVideoProcessCount: -1,
		imageProcessing:      true,
		videoProcessing:      true,
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

func (ib *indexBuilder) getFileLoader(
	ctx context.Context,
	waitGroup *sync.WaitGroup,
	sem *semaphore.Weighted,
	processor fileProcessor,
) func(FileMeta) error {
	acquire := func() error {
		err := sem.Acquire(ctx, 1)
		if err == nil || errors.Is(err, context.Canceled) {
			return nil
		}

		return fmt.Errorf("%w: %w", ErrSemaphoreAcquire, err)
	}

	return func(file FileMeta) error {
		metaItem := metaByFile(file)

		// if item already in map, but without preview -> create preview
		if saved, ok := ib.index.meta[metaItem.ID]; ok && saved.Preview.Length != 0 {
			return nil
		}

		// needs for check if file/folder was modified, and remove old version
		if oldMeta, ok := ib.index.paths[metaItem.RelativePath]; ok {
			delete(ib.index.meta, oldMeta.ID)
		}

		if err := acquire(); err != nil {
			return err
		}

		waitGroup.Add(1)
		ib.params.progress()

		go processor.run(metaItem)

		return nil
	}
}

func (ib *indexBuilder) loadFiles(ctx context.Context) error {
	waitGroup := new(sync.WaitGroup)
	sem := semaphore.NewWeighted(int64(ib.params.workers))
	resultChannel := make(chan fileProcessorResult, len(ib.params.files))

	excludedFromPreview, err := ib.getExcludedFiles()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGetExcludedFiles, err)
	}

	if len(excludedFromPreview) != 0 {
		slog.Info(fmt.Sprintf("Excluded %d files from media processing", len(excludedFromPreview)))
	}

	// pass biggest files first
	slices.SortStableFunc(ib.params.files, compareFileMetaSize)

	processor := fileProcessor{
		ch:        resultChannel,
		waitGroup: waitGroup,
		semaphore: sem,
		image:     ib.params.imageProcessing,
		video:     ib.params.videoProcessing,
		maxImage:  ib.params.maxImageProcessCount,
		maxVideo:  ib.params.maxVideoProcessCount,
		preview:   ib.params.preview,
		excluded:  excludedFromPreview,
	}

	fileLoader := ib.getFileLoader(ctx, waitGroup, sem, processor)
	for _, file := range ib.params.files {
		if err := fileLoader(file); err != nil {
			return fmt.Errorf("%w: %w", ErrFileLoad, err)
		}
	}

	waitGroup.Wait()
	close(resultChannel)

	for result := range resultChannel {
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
