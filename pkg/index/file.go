package index

import (
	"log/slog"
	"sync"

	"golang.org/x/sync/semaphore"
)

type fileProcessor struct {
	preview   PreviewGenerator
	id        IDGenerator
	idPass    fileIDPass
	semaphore *semaphore.Weighted
	waitGroup *sync.WaitGroup
	ch        chan *fileProcessorResult
}
type fileProcessorResult struct {
	meta IndexMeta
	data []byte
}

type fileProcessorOption func(*fileProcessor)
type fileIDPass func(string) bool

func withPreview(f PreviewGenerator) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.preview = f
	}
}

func withID(f IDGenerator) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.id = f
	}
}

func withIDCheck(f fileIDPass) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.idPass = f
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

func withChan(ch chan *fileProcessorResult) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.ch = ch
	}
}

func processFile(f FileMeta, opts ...fileProcessorOption) (*fileProcessorResult, error) {
	processor := fileProcessor{}
	for _, opt := range opts {
		opt(&processor)
	}
	result, err := processor.run(f)
	if err != nil {
		slog.Error("processFile", err.Error(), slog.String("file", f.Path()))
	}
	if processor.ch != nil && result != nil {
		processor.ch <- result
	}
	if processor.semaphore != nil {
		processor.semaphore.Release(1)
	}
	if processor.waitGroup != nil {
		processor.waitGroup.Done()
	}
	return result, err
}

func (fp *fileProcessor) run(f FileMeta) (*fileProcessorResult, error) {
	id, err := fp.id(f)
	if err != nil {
		return nil, err
	}
	if !fp.idPass(id) {
		return nil, nil
	}
	meta := IndexMeta{
		Path:         f.Path(),
		RelativePath: f.RelativePath(),
		Name:         f.Name(),
		ModTime:      f.ModTime(),
		IsDir:        f.IsDir(),
		ID:           id,
	}
	duration, t, data, err := fp.preview(meta.Path)
	if err != nil {
		return nil, err
	}
	meta.Duration = duration
	meta.Type = t
	meta.Preview = IndexMetaPreview{
		Length: uint32(len(data)),
	}
	return &fileProcessorResult{meta, data}, nil
}
