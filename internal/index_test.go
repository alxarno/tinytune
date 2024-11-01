package internal

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexEncodeDecode(t *testing.T) {
	indexOriginal := NewIndex(nil)
	indexOriginal.meta = map[string]IndexMeta{
		"5762029e772e6587ddd90f08c1bf374486436eb11f81dd6a8f03bcd82d335a7f": {
			Path:  "/home/test/",
			Name:  "test",
			Hash:  "5762029e772e6587ddd90f08c1bf374486436eb11f81dd6a8f03bcd82d335a7f",
			IsDir: true,
		},
		"2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824": {
			Path:  "/home/test/abc.jpg",
			Name:  "abc.jpg",
			Hash:  "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824",
			IsDir: false,
			Preview: IndexMetaPreview{
				Length: 100,
				Offset: 0,
			},
		},
	}
	indexOriginal.data = make([]byte, 100)
	rand.Read(indexOriginal.data)
	buff := new(bytes.Buffer)
	wrote, err := indexOriginal.Encode(buff)
	assert.NoError(t, err)
	assert.EqualValues(t, 510, wrote)
	assert.EqualValues(t, 510, buff.Len())
	// Parse
	indexDerivative := NewIndex(buff)
	assert.Equal(t, len(indexOriginal.meta), len(indexDerivative.meta))
	assert.Equal(t, indexOriginal.meta, indexDerivative.meta)
	assert.Equal(t, len(indexOriginal.data), len(indexDerivative.data))
	assert.Equal(t, indexOriginal.data, indexDerivative.data)
}

func TestIndexFiles(t *testing.T) {
	files, err := NewCrawlerOS("../test/").Scan()
	assert.NoError(t, err)
	index := NewIndex(
		nil,
		WithFiles(files),
		WithPreview(GeneratePreview),
		WithID(func(p FileMeta) (string, error) {
			return SHA256Hash(bytes.NewReader([]byte(fmt.Sprintf("%s%s", p.Name(), p.ModTime()))))
		}))
	fmt.Println(index.meta)
	fmt.Println(len(index.data))
}
