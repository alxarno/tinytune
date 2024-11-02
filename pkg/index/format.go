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

func init() {
	gob.Register(IndexMeta{})
}

const INDEX_HEADER = "TINYTUNE_INDEX"
const INDEX_DELIMITER = "TINYTUNE_DELIMITER"

func searchDelimiter(r *bufio.Reader, delimiter []byte) ([]byte, error) {
	buff := make([]byte, 0, 1024)
	possibleDelimiter := make([]byte, len(delimiter)-1)
	for {
		b, err := r.ReadBytes(delimiter[0])
		if err != nil {
			return nil, err
		}
		buff = append(buff, b...)
		_, err = r.Read(possibleDelimiter)
		if err != nil {
			return nil, err
		}
		if bytes.Equal(possibleDelimiter, delimiter[1:]) {
			return buff[:len(buff)-1], nil
		} else {
			buff = append(buff, possibleDelimiter...)
		}
	}
}

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
	// read meta items length
	bs := make([]byte, 4)
	if _, err := r.Read(bs); err != nil {
		return err
	}
	metaLength := binary.LittleEndian.Uint32(bs)
	// read meta
	metaBytes, err := searchDelimiter(bufio.NewReader(r), []byte(INDEX_DELIMITER))
	if err != nil {
		return err
	}
	decoder := gob.NewDecoder(bytes.NewReader(metaBytes))
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
