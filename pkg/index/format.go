package index

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"

	"github.com/alxarno/tinytune/pkg/bytesutil"
)

const INDEX_HEADER = "TINYTUNE_INDEX"
const INDEX_DELIMITER = "TINYTUNE_DELIMITER"

func indexDelimiterSplit(data []byte, atEOF bool) (advance int, token []byte, err error) {
	// Find the delimiter
	if i := bytes.Index(data, []byte(INDEX_DELIMITER)); i >= 0 {
		return i + 1 + len([]byte(INDEX_DELIMITER)), data[0:i], nil
	}
	// If at end of file and no comma found, return the entire remaining data
	if atEOF {
		return len(data), data, nil
	}
	// Request more data
	return 0, nil, nil
}

func (index *Index) Decode(r io.Reader) error {
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
	scanner.Split(indexDelimiterSplit)
	scanner.Scan()
	decoder := gob.NewDecoder(bytes.NewBuffer(scanner.Bytes()))
	for i := 0; i < int(metaLength); i++ {
		m := IndexMeta{}
		if err := decoder.Decode(&m); err != nil {
			return err
		}
		index.meta[m.ID] = m
	}
	// read binary data
	if index.data, err = io.ReadAll(r); err != nil {
		return err
	}
	return nil
}

func (i Index) Encode(w io.Writer) (uint64, error) {
	// write header
	wc := bytesutil.NewWriterCounter(w)
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
