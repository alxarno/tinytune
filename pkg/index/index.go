package index

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/semaphore"
)

var NotFoundError = fmt.Errorf("not found")

type PreviewGenerator func(string) (time.Duration, int, []byte, error)
type IDGenerator func(m FileMeta) (string, error)

type Index struct {
	meta     map[string]*IndexMeta
	tree     map[string][]*IndexMeta
	paths    map[string]*IndexMeta
	data     []byte
	outDated bool
	preview  PreviewGenerator
	id       IDGenerator
	files    []FileMeta
	context  context.Context
	progress func()
	newFiles func()
	workers  int
}

type IndexOption func(*Index)

type FileMeta interface {
	Path() string
	RelativePath() string
	Name() string
	ModTime() time.Time
	IsDir() bool
	Size() int64
}

func NewIndex(r io.Reader, opts ...IndexOption) (Index, error) {
	index := Index{
		data:     []byte{},
		meta:     map[string]*IndexMeta{},
		preview:  func(string) (time.Duration, int, []byte, error) { return 0, ContentTypeOther, nil, nil },
		id:       func(m FileMeta) (string, error) { return m.Path(), nil },
		files:    []FileMeta{},
		tree:     map[string][]*IndexMeta{},
		paths:    map[string]*IndexMeta{},
		context:  context.Background(),
		outDated: false,
		progress: func() {},
		newFiles: func() {},
		workers:  4,
	}
	for _, opt := range opts {
		opt(&index)
	}

	if err := index.Decode(r); err != nil {
		if errors.Is(err, io.EOF) {
			slog.Warn("The index file could not be fully read, it may be corrupted or empty")
		} else {
			return index, err
		}
	}
	if err := index.loadFiles(); err != nil {
		return index, err
	}

	if err := index.loadTree(); err != nil {
		return index, err
	}

	if err := index.loadPaths(); err != nil {
		return index, err
	}
	return index, nil
}

func WithPreview(f func(path string) (time.Duration, int, []byte, error)) IndexOption {
	return func(i *Index) {
		i.preview = f
	}
}

func WithID(f func(m FileMeta) (string, error)) IndexOption {
	return func(i *Index) {
		i.id = f
	}
}

func WithFiles(files []FileMeta) IndexOption {
	return func(i *Index) {
		i.files = files
	}
}

func WithContext(ctx context.Context) IndexOption {
	return func(i *Index) {
		i.context = ctx
	}
}

func WithProgress(f func()) IndexOption {
	return func(i *Index) {
		i.progress = f
	}
}

func WithNewFiles(f func()) IndexOption {
	return func(i *Index) {
		i.newFiles = f
	}
}

func WithWorkers(w int) IndexOption {
	return func(i *Index) {
		i.workers = w
	}
}
func (index Index) OutDated() bool {
	return index.outDated
}

func (index Index) PullPreview(id string) ([]byte, error) {
	m, ok := index.meta[id]
	if !ok {
		return nil, NotFoundError
	}
	if m.Preview.Length == 0 {
		return nil, nil
	}
	return index.data[m.Preview.Offset : m.Preview.Offset+m.Preview.Length], nil
}

func (index Index) PullChildren(id string) ([]*IndexMeta, error) {
	result := make([]*IndexMeta, 0)

	// return root children
	if id == "" {
		for _, m := range index.meta {
			if !strings.Contains(m.RelativePath, "/") {
				result = append(result, m)
			}
		}
		return result, nil
	}

	if children, ok := index.tree[id]; ok {
		result = children
	} else {
		return nil, NotFoundError
	}
	return result, nil
}

func (index Index) PullPaths(id string) ([]*IndexMeta, error) {
	result := []*IndexMeta{}
	if id == "" {
		return result, nil
	}
	m, ok := index.meta[id]
	if !ok || !m.IsDir {
		return nil, NotFoundError
	}
	paths := strings.Split(m.RelativePath, string(os.PathSeparator))
	subDirs := []string{}
	slices.Reverse(paths)
	for i, v := range paths {
		buff := paths[i+1:]
		subDirectory := filepath.Join(append(buff, v)...)
		subDirs = append(subDirs, subDirectory)
	}
	slices.Reverse(subDirs)
	for _, v := range subDirs {
		result = append(result, index.paths[v])
	}
	return result, nil
}

func (index Index) Search(query string, dir string) []*IndexMeta {
	result := []*IndexMeta{}
	query = strings.ToLower(query)
	filter := func(v *IndexMeta) {
		if strings.Contains(strings.ToLower(v.Name), query) {
			result = append(result, v)
		}
	}
	if dir == "" {
		for _, v := range index.meta {
			filter(v)
		}
		return result
	}
	if children, ok := index.tree[dir]; !ok {
		return result
	} else {
		for _, v := range children {
			filter(v)
		}
	}

	return result
}

func (index Index) FilesWithPreviewStat() (int, uint32) {
	count := 0
	size := uint32(0)
	for _, v := range index.meta {
		if v.Preview.Length != 0 {
			count++
			size += v.Preview.Length
		}
	}
	return count, size
}

func (index *Index) loadFiles() error {
	wg := new(sync.WaitGroup)
	sem := semaphore.NewWeighted(int64(index.workers))
	resultChannel := make(chan fileProcessorResult, len(index.files))

	for _, file := range index.files {
		if err := sem.Acquire(index.context, 1); err != nil {
			if errors.Is(err, context.Canceled) {
				return nil
			}
			return err
		}
		wg.Add(1)
		index.progress()
		go processFile(
			file,
			withID(index.id),
			withPreview(index.preview),
			withIDCheck(func(id string) bool { _, ok := index.meta[id]; return !ok }),
			withChan(resultChannel),
			withSemaphore(sem),
			withWaitGroup(wg))
	}

	wg.Wait()
	close(resultChannel)

	for r := range resultChannel {
		index.newFiles()
		if r.meta.Preview.Length != 0 {
			r.meta.Preview.Offset = uint32(len(index.data))
			index.data = append(index.data, r.data...)
		}
		if r.meta.IsDir {
			r.meta.Type = ContentTypeDir
		}
		index.meta[r.meta.ID] = r.meta
		index.outDated = true
	}

	return nil
}

func (index *Index) loadTree() error {
	index.tree["root"] = make([]*IndexMeta, 0)
	for _, meta := range index.meta {
		if filepath.Dir(meta.RelativePath) == "." {
			index.tree["root"] = append(index.tree["root"], meta)
		}
		if !meta.IsDir {
			continue
		}

		for _, possibleChild := range index.meta {
			if meta.RelativePath != filepath.Dir(possibleChild.RelativePath) {
				continue
			}
			if _, ok := index.tree[meta.ID]; !ok {
				index.tree[meta.ID] = make([]*IndexMeta, 0)
			}
			index.tree[meta.ID] = append(index.tree[meta.ID], possibleChild)
		}
	}
	return nil
}

func (index *Index) loadPaths() error {
	for _, v := range index.meta {
		index.paths[v.RelativePath] = v
	}
	return nil
}
