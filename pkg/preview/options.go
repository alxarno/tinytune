package preview

import "time"

type Option func(*Previewer)

func WithImage(param bool) Option {
	return func(p *Previewer) {
		p.image = param
	}
}

func WithVideo(param bool) Option {
	return func(p *Previewer) {
		p.video = param
	}
}

func WithVideoAccel(param VideoProcessingAccelType) Option {
	return func(p *Previewer) {
		p.videoAccelType = param
	}
}

// files = map[FilePath]struct{}.
func WithExcludedFiles(files map[string]struct{}) Option {
	return func(p *Previewer) {
		p.excludedFiles = files
	}
}

func WithMaxImages(param int64) Option {
	return func(p *Previewer) {
		p.maxImages = param
	}
}

func WithMaxVideos(param int64) Option {
	return func(p *Previewer) {
		p.maxVideos = param
	}
}

func WithMaxFileSize(param int64) Option {
	return func(p *Previewer) {
		p.maxFileSize = param
	}
}

func WithTimeout(param time.Duration) Option {
	return func(p *Previewer) {
		p.timeout = param
	}
}
