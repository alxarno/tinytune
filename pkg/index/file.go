package index

import (
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

var ErrPreviewPull = errors.New("failed to pull preview")

type fileProcessor struct {
	preview   PreviewGenerator
	semaphore *semaphore.Weighted
	waitGroup *sync.WaitGroup
	ch        chan fileProcessorResult
	excluded  map[RelativePath]struct{}
	image     bool
	video     bool
	maxImage  int64
	maxVideo  int64
}

type fileProcessorResult struct {
	meta *Meta
	data []byte
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

func (fp *fileProcessor) imageProcess() bool {
	return fp.image && ifMaxPass(&fp.maxImage)
}

func (fp *fileProcessor) videoProcess() bool {
	return fp.video && ifMaxPass(&fp.maxVideo)
}

func (fp *fileProcessor) run(meta *Meta) {
	defer fp.semaphore.Release(1)
	defer fp.waitGroup.Done()

	_, shouldExclude := fp.excluded[meta.RelativePath]
	skipPreview := fp.preview == nil ||
		shouldExclude ||
		meta.IsDir ||
		meta.IsImage() && !fp.imageProcess() ||
		meta.IsVideo() && !fp.videoProcess()

	if skipPreview {
		fp.ch <- fileProcessorResult{meta, nil}

		return
	}

	preview, err := fp.preview(meta)
	if err != nil {
		slog.Error(fmt.Errorf("%w: %w", ErrPreviewPull, err).Error())

		return
	}

	meta.Duration = preview.Duration()
	meta.Resolution = preview.Resolution()
	meta.Preview = PreviewLocation{
		Length: uint32(len(preview.Data())),
	}

	fp.ch <- fileProcessorResult{meta, preview.Data()}
}
