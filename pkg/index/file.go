package index

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"golang.org/x/sync/semaphore"
)

var ErrPreviewPull = errors.New("failed to pull preview")

type fileProcessor struct {
	preview   PreviewGenerator
	semaphore *semaphore.Weighted
	waitGroup *sync.WaitGroup
	ch        chan fileProcessorResult
}

type fileProcessorResult struct {
	meta *Meta
	data []byte
}

type fileProcessorOption func(*fileProcessor)

func withPreview(f PreviewGenerator) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.preview = f
	}
}

func withSemaphore(sem *semaphore.Weighted) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.semaphore = sem
	}
}

func withWaitGroup(wg *sync.WaitGroup) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.waitGroup = wg
	}
}

func withChan(ch chan fileProcessorResult) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.ch = ch
	}
}

func newFileProcessor(opts ...fileProcessorOption) fileProcessor {
	processor := fileProcessor{}
	for _, opt := range opts {
		opt(&processor)
	}

	return processor
}

func (fp *fileProcessor) run(file FileMeta, id string) {
	meta := Meta{
		Path:         file.Path(),
		RelativePath: file.RelativePath(),
		Name:         file.Name(),
		ModTime:      file.ModTime(),
		IsDir:        file.IsDir(),
		ID:           id,
	}

	preview, err := fp.preview.Pull(meta.Path)
	if err != nil {
		slog.Error(fmt.Errorf("%w: %w", ErrPreviewPull, err).Error())

		return
	}

	meta.Duration = preview.Duration
	meta.Type = preview.ContentType
	meta.Resolution = preview.Resolution
	meta.Preview = Preview{
		Length: uint32(len(preview.Data)),
	}
	result := &fileProcessorResult{&meta, preview.Data}

	if fp.ch != nil {
		fp.ch <- *result
	}

	if fp.semaphore != nil {
		fp.semaphore.Release(1)
	}

	if fp.waitGroup != nil {
		fp.waitGroup.Done()
	}
}
