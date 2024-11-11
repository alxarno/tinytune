package index

import (
	"sync"

	"golang.org/x/sync/semaphore"
)

type fileProcessor struct {
	preview   PreviewGenerator
	semaphore *semaphore.Weighted
	waitGroup *sync.WaitGroup
	ch        chan fileProcessorResult
}

type fileProcessorResult struct {
	meta *IndexMeta
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

func (fp *fileProcessor) run(f FileMeta, id string) (*fileProcessorResult, error) {
	meta := IndexMeta{
		Path:         f.Path(),
		RelativePath: f.RelativePath(),
		Name:         f.Name(),
		ModTime:      f.ModTime(),
		IsDir:        f.IsDir(),
		ID:           id,
	}
	preview, err := fp.preview.Pull(meta.Path)
	if err != nil {
		return nil, err
	}
	meta.Duration = preview.Duration
	meta.Type = preview.ContentType
	meta.Resolution = preview.Resolution
	meta.Preview = IndexMetaPreview{
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

	return result, nil
}
