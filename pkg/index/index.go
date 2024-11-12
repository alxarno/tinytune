package index

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var ErrNotFound = errors.New("not found")

type Index struct {
	meta     map[string]*Meta
	tree     map[string][]*Meta
	paths    map[string]*Meta
	data     []byte
	outDated bool
}

func NewIndex(ctx context.Context, r io.Reader, opts ...Option) (Index, error) {
	index := Index{
		data:     []byte{},
		meta:     map[string]*Meta{},
		tree:     map[string][]*Meta{},
		paths:    map[string]*Meta{},
		outDated: false,
	}
	builder := newBuilder(&index)

	for _, opt := range opts {
		opt(&builder)
	}

	if err := builder.run(ctx, r); err != nil {
		return index, err
	}

	return index, nil
}

func (index *Index) OutDated() bool {
	return index.outDated
}

func (index *Index) Pull(id string) (*Meta, error) {
	m, ok := index.meta[id]
	if !ok {
		return nil, ErrNotFound
	}

	return m, nil
}

func (index *Index) PullPreview(id string) ([]byte, error) {
	meta, ok := index.meta[id]
	if !ok {
		return nil, ErrNotFound
	}

	if meta.Preview.Length == 0 {
		return nil, nil
	}

	return index.data[meta.Preview.Offset : meta.Preview.Offset+meta.Preview.Length], nil
}

func (index *Index) PullChildren(id string) ([]*Meta, error) {
	result := make([]*Meta, 0)

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
		return children, nil
	}

	return nil, ErrNotFound
}

func (index *Index) PullPaths(id string) ([]*Meta, error) {
	result := []*Meta{}
	if id == "" {
		return result, nil
	}

	m, ok := index.meta[id]
	if !ok || !m.IsDir {
		return nil, ErrNotFound
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

func (index *Index) Search(query string, dir string) []*Meta {
	result := []*Meta{}
	query = strings.ToLower(query)
	filter := func(v *Meta) {
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

	children, ok := index.tree[dir]
	if !ok {
		return result
	}

	for _, v := range children {
		filter(v)
	}

	return result
}

func (index *Index) FilesWithPreviewStat() (int, uint32) {
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
