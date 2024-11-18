package index

import "time"

type FileMeta interface {
	Path() string
	RelativePath() string
	Name() string
	ModTime() time.Time
	IsDir() bool
	Size() int64
}

func metaByFile(file FileMeta) *Meta {
	metaItem := &Meta{
		AbsolutePath: Path(file.Path()),
		RelativePath: RelativePath(file.RelativePath()),
		Name:         file.Name(),
		ModTime:      file.ModTime(),
		OriginSize:   file.Size(),
		IsDir:        file.IsDir(),
	}
	metaItem.generateID()
	metaItem.setContentType()

	return metaItem
}
