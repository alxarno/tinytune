package internal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
)

const INDEX_HEADER = "TINYTUNE_INDEX"
const INDEX_DELIMITER = "TINYTUNE_DELIMITER"

type Index struct {
	meta []IndexMeta
	data []byte
}

type IndexMeta struct {
	Path    string
	Name    string
	Hash    string
	IsDir   bool
	Preview IndexMetaPreview
}

type IndexMetaPreview struct {
	Length uint32
	Offset uint32
}

func delimiterSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Find the delimiter
	if i := bytes.Index(data, []byte(INDEX_DELIMITER)); i >= 0 {
		return i + 1, data[0:i], nil
	}
	// If at end of file and no comma found, return the entire remaining data
	if atEOF {
		return len(data), data, nil
	}
	// Request more data
	return 0, nil, nil
}

func NewIndex(r io.Reader) Index {
	gob.Register(IndexMeta{})
	index := Index{data: []byte{}, meta: []IndexMeta{}}
	if r != nil {
		index.parse(r)
	}
	return index
}

func (index *Index) parse(r io.Reader) error {
	// read header
	header := make([]byte, len([]byte(INDEX_HEADER)))
	n, err := r.Read(header)
	if err != nil {
		return err
	}
	if n == 0 {
		return nil
	}
	if string(header) != INDEX_HEADER {
		return fmt.Errorf("invalid header: %s", string(header))
	}
	// read meta items length
	bs := make([]byte, 4)
	if _, err := r.Read(bs); err != nil {
		return err
	}
	metaLength := binary.LittleEndian.Uint32(bs)
	// read meta
	scanner := bufio.NewScanner(r)
	scanner.Split(delimiterSplit)
	scanner.Scan()
	decoder := gob.NewDecoder(bytes.NewBuffer(scanner.Bytes()))
	for i := 0; i < int(metaLength); i++ {
		m := IndexMeta{}
		if err := decoder.Decode(&m); err != nil {
			return err
		}
		index.meta = append(index.meta, m)
	}
	// read binary data
	scanner.Scan()
	index.data = scanner.Bytes()[len([]byte(INDEX_DELIMITER))-1:]
	return nil
}

func (i Index) Dump(w io.Writer) (uint64, error) {
	// write header
	wc := NewWriterCounter(w)
	if _, err := wc.Write([]byte(INDEX_HEADER)); err != nil {
		return 0, err
	}
	// write meta items count
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(i.meta)))
	if _, err := wc.Write(bs); err != nil {
		return 0, err
	}
	// write meta items
	enc := gob.NewEncoder(wc)
	for _, v := range i.meta {
		if err := enc.Encode(v); err != nil {
			return 0, err
		}
	}
	if _, err := wc.Write([]byte(INDEX_DELIMITER)); err != nil {
		return 0, err
	}
	// write binary data
	if _, err := wc.Write(i.data); err != nil {
		return 0, err
	}
	return wc.Count(), nil
}
