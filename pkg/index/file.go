package index

import (
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

type fileProcessor struct {
	preview          PreviewGenerator
	id               IDGenerator
	idPass           fileIDPass
	semaphore        *semaphore.Weighted
	waitGroup        *sync.WaitGroup
	ch               chan fileProcessorResult
	maxNewImageItems int64
	maxNewVideoItems int64
}
type fileProcessorResult struct {
	meta *IndexMeta
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

func withChan(ch chan fileProcessorResult) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.ch = ch
	}
}

func withMaxImageItems(param int64) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.maxNewImageItems = param
	}
}

func withMaxVideoItems(param int64) fileProcessorOption {
	return func(fp *fileProcessor) {
		fp.maxNewVideoItems = param
	}
}

func newFileProcessor(opts ...fileProcessorOption) fileProcessor {
	processor := fileProcessor{}
	for _, opt := range opts {
		opt(&processor)
	}
	return processor
}

func ifMaxPass(max *int64) bool {
	if atomic.LoadInt64(max) == -1 {
		return true
	}
	if atomic.LoadInt64(max) == 0 {
		return false
	}
	atomic.AddInt64(max, -1)
	return true
}

func (fp *fileProcessor) run(f FileMeta) (*fileProcessorResult, error) {
	id, err := fp.id(f)
	if err != nil {
		return nil, err
	}
	if !fp.idPass(id) {
		return nil, nil
	}
	contentType := fp.preview.ContentType(f.Path())
	if contentType == ContentTypeVideo && !ifMaxPass(&fp.maxNewVideoItems) {
		return nil, nil
	}

	if contentType == ContentTypeImage && !ifMaxPass(&fp.maxNewImageItems) {
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
