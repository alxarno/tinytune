package index

import "context"

type IndexOption func(*indexBuilder)

func WithPreview(gen PreviewGenerator) IndexOption {
	return func(i *indexBuilder) {
		i.params.preview = gen
	}
}

func WithID(f func(m FileMeta) (string, error)) IndexOption {
	return func(i *indexBuilder) {
		i.params.id = f
	}
}

func WithFiles(files []FileMeta) IndexOption {
	return func(i *indexBuilder) {
		i.params.files = files
	}
}

func WithContext(ctx context.Context) IndexOption {
	return func(i *indexBuilder) {
		i.params.context = ctx
	}
}

func WithProgress(f func()) IndexOption {
	return func(i *indexBuilder) {
		i.params.progress = f
	}
}

func WithNewFiles(f func()) IndexOption {
	return func(i *indexBuilder) {
		i.params.newFiles = f
	}
}

func WithMaxNewImageItems(param int64) IndexOption {
	return func(i *indexBuilder) {
		i.params.maxNewImageItems = param
	}
}

func WithMaxNewVideoItems(param int64) IndexOption {
	return func(i *indexBuilder) {
		i.params.maxNewVideoItems = param
	}
}

func WithVideo(param bool) IndexOption {
	return func(i *indexBuilder) {
		i.params.videoProcessing = param
	}
}

func WithImage(param bool) IndexOption {
	return func(i *indexBuilder) {
		i.params.imageProcessing = param
	}
}

func WithWorkers(w int) IndexOption {
	return func(i *indexBuilder) {
		i.params.workers = w
	}
}
