package internal

import (
	"bytes"
	"crypto/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexParseDump(t *testing.T) {
	// Dump
	indexOriginal := NewIndex(nil)
	indexOriginal.meta = []IndexMeta{
		{
			Path:  "/home/test/",
			Name:  "test",
			Hash:  "5762029e772e6587ddd90f08c1bf374486436eb11f81dd6a8f03bcd82d335a7f",
			IsDir: true,
		},
		{
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
	wrote, err := indexOriginal.Dump(buff)
	assert.NoError(t, err)
	assert.EqualValues(t, 458, wrote)
	assert.EqualValues(t, 458, buff.Len())
	// Parse
	indexDerivative := NewIndex(buff)
	assert.Equal(t, len(indexOriginal.meta), len(indexDerivative.meta))
	assert.ElementsMatch(t, indexOriginal.meta, indexDerivative.meta)
	assert.Equal(t, len(indexOriginal.data), len(indexDerivative.data))
	assert.ElementsMatch(t, indexOriginal.data, indexDerivative.data)
}
