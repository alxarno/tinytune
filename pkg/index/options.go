package index

type Option func(*indexBuilder)

func WithPreview(gen PreviewGenerator) Option {
	return func(i *indexBuilder) {
		i.params.preview = gen
	}
}

func WithID(f func(m FileMeta) (string, error)) Option {
	return func(i *indexBuilder) {
		i.params.id = f
	}
}

func WithFiles(files []FileMeta) Option {
	return func(i *indexBuilder) {
		i.params.files = files
	}
}

func WithProgress(f func()) Option {
	return func(i *indexBuilder) {
		i.params.progress = f
	}
}

func WithNewFiles(f func()) Option {
	return func(i *indexBuilder) {
		i.params.newFiles = f
	}
}

func WithMaxNewImageItems(param int64) Option {
	return func(i *indexBuilder) {
		i.params.maxNewImageItems = param
	}
}

func WithMaxNewVideoItems(param int64) Option {
	return func(i *indexBuilder) {
		i.params.maxNewVideoItems = param
	}
}

func WithVideo(param bool) Option {
	return func(i *indexBuilder) {
		i.params.videoProcessing = param
	}
}

func WithImage(param bool) Option {
	return func(i *indexBuilder) {
		i.params.imageProcessing = param
	}
}

func WithWorkers(w int) Option {
	return func(i *indexBuilder) {
		i.params.workers = w
	}
}
