package index

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
)

var ErrNotFound = fmt.Errorf("not found")

type Index struct {
	meta     map[string]*IndexMeta
	tree     map[string][]*IndexMeta
	paths    map[string]*IndexMeta
	data     []byte
	outDated bool
}

func NewIndex(r io.Reader, opts ...IndexOption) (Index, error) {
	index := Index{
		data:     make([]byte, 0, 64*1024),
		meta:     map[string]*IndexMeta{},
		tree:     map[string][]*IndexMeta{},
		paths:    map[string]*IndexMeta{},
		outDated: false,
	}
	builder := newBuilder(&index)
	for _, opt := range opts {
		opt(&builder)
	}
	if err := builder.run(r); err != nil {
		return index, err
	}

	return index, nil
}

func (index Index) OutDated() bool {
	return index.outDated
}

func (index Index) Pull(id string) (*IndexMeta, error) {
	if m, ok := index.meta[id]; !ok {
		return nil, ErrNotFound
	} else {
		return m, nil
	}
}

func (index Index) PullPreview(id string) ([]byte, error) {
	m, ok := index.meta[id]
	if !ok {
		return nil, ErrNotFound
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
		return nil, ErrNotFound
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
