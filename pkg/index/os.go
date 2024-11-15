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

func compareFileMetaSize(a, b FileMeta) int {
	if a.Size() == b.Size() {
		return 0
	}

	if a.Size() < b.Size() {
		return 1
	}

	return -1
}

func metaByFile(file FileMeta) *Meta {
	metaItem := &Meta{
		AbsolutePath: Path(file.Path()),
		RelativePath: RelativePath(file.RelativePath()),
		Name:         file.Name(),
		ModTime:      file.ModTime(),
		IsDir:        file.IsDir(),
	}
	metaItem.generateID()
	metaItem.setContentType()

	return metaItem
}
