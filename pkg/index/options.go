package index

type Option func(*indexBuilder)

func WithPreview(gen PreviewGenerator) Option {
	return func(i *indexBuilder) {
		i.params.preview = gen
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

func WithWorkers(w int) Option {
	return func(i *indexBuilder) {
		i.params.workers = w
	}
}
