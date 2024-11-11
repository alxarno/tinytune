package index

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/alxarno/tinytune/pkg/bytesutil"
)

func init() {
	gob.Register(IndexMeta{})
}

const INDEX_HEADER = "TINYTUNE_INDEX"

func (index *Index) Decode(r io.Reader) error {
	if r == nil {
		return nil
	}
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
	// read meta items count
	bs := make([]byte, 4)
	if _, err := r.Read(bs); err != nil {
		return err
	}
	metaItemsCount := binary.LittleEndian.Uint32(bs)

	// read meta part size
	if _, err := r.Read(bs); err != nil {
		return err
	}
	metaItemsSize := binary.LittleEndian.Uint32(bs)

	// read meta
	metaPartBuffer := make([]byte, metaItemsSize)
	if _, err = r.Read(metaPartBuffer); err != nil {
		return err
	}
	decoder := gob.NewDecoder(bytes.NewReader(metaPartBuffer))
	for i := 0; i < int(metaItemsCount); i++ {
		m := IndexMeta{}
		if err := decoder.Decode(&m); err != nil {
			return err
		}
		index.meta[m.ID] = &m
	}
	// read binary data
	if index.data, err = io.ReadAll(r); err != nil {
		return err
	}
	return nil
}

func (index *Index) Encode(w io.Writer) (uint64, error) {
	// write header
	wc := bytesutil.NewWriterCounter(w)
	if _, err := wc.Write([]byte(INDEX_HEADER)); err != nil {
		return 0, err
	}
	// write meta items count
	bs := make([]byte, 4)
	binary.LittleEndian.PutUint32(bs, uint32(len(index.meta)))
	if _, err := wc.Write(bs); err != nil {
		return 0, err
	}
	// prepare meta items
	metaBuffer := bytes.NewBuffer(make([]byte, 0))
	enc := gob.NewEncoder(metaBuffer)
	for _, v := range index.meta {
		if err := enc.Encode(v); err != nil {
			return 0, err
		}
	}

	// write meta part size
	binary.LittleEndian.PutUint32(bs, uint32(metaBuffer.Len()))
	if _, err := wc.Write(bs); err != nil {
		return 0, err
	}

	// write meta part
	if _, err := wc.Write(metaBuffer.Bytes()); err != nil {
		return 0, err
	}
	// write binary data
	if _, err := wc.Write(index.data); err != nil {
		return 0, err
	}
	return wc.Count(), nil
}
