package index

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"io"

	"github.com/alxarno/tinytune/pkg/bytesutil"
)

//nolint:gochecknoinits
func init() {
	gob.Register(Meta{})
}

var (
	ErrInvalidHeader       = errors.New("invalid header")
	ErrReadHeader          = errors.New("failed to read header")
	ErrReadMetaItemsCount  = errors.New("failed to read meta items count")
	ErrReadMetaPartSize    = errors.New("failed to read meta's part size")
	ErrReadMetaPart        = errors.New("failed to read meta items")
	ErrMetaItemDecode      = errors.New("failed to decode meta's item")
	ErrReadBinaryData      = errors.New("failed to read binary data")
	ErrWriteHeader         = errors.New("failed to write header")
	ErrWriteMetaItemsCount = errors.New("failed to write meta items count")
	ErrEncodeMetaItem      = errors.New("failed to encode meta's item")
	ErrWriteMetaPartSize   = errors.New("failed to write meta's part size")
	ErrWriteMetaPart       = errors.New("failed to write meta items")
	ErrWriteBinaryData     = errors.New("failed to write binary data")
)

const indexHeader = "TINYTUNE_INDEX"
const metaItemsCountSize = 4

//nolint:cyclop
func (index *Index) Decode(r io.Reader) error {
	if r == nil {
		return nil
	}
	// read header
	header := make([]byte, len([]byte(indexHeader)))

	n, err := r.Read(header)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrReadHeader, err)
	}

	if n == 0 {
		return nil
	}

	if string(header) != indexHeader {
		return fmt.Errorf("%s is %w", string(header), ErrInvalidHeader)
	}
	// read meta items count
	buffer := make([]byte, metaItemsCountSize)
	if _, err := r.Read(buffer); err != nil {
		return fmt.Errorf("%w: %w", ErrReadMetaItemsCount, err)
	}

	metaItemsCount := binary.LittleEndian.Uint32(buffer)

	// read meta part size
	if _, err := r.Read(buffer); err != nil {
		return fmt.Errorf("%w: %w", ErrReadMetaPartSize, err)
	}

	metaItemsSize := binary.LittleEndian.Uint32(buffer)

	// read meta
	metaPartBuffer := make([]byte, metaItemsSize)
	if _, err = r.Read(metaPartBuffer); err != nil {
		return fmt.Errorf("%w: %w", ErrReadMetaPart, err)
	}

	decoder := gob.NewDecoder(bytes.NewReader(metaPartBuffer))

	for range metaItemsCount {
		m := Meta{}
		if err := decoder.Decode(&m); err != nil {
			return fmt.Errorf("%w: %w", ErrMetaItemDecode, err)
		}

		index.meta[m.ID] = &m
	}

	// read binary data
	if index.data, err = io.ReadAll(r); err != nil {
		return fmt.Errorf("%w: %w", ErrReadBinaryData, err)
	}

	return nil
}

func (index *Index) Encode(w io.Writer) (uint64, error) {
	// write header
	writer := bytesutil.NewWriterCounter(w)
	if _, err := writer.Write([]byte(indexHeader)); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrWriteHeader, err)
	}
	// write meta items count
	buffer := make([]byte, metaItemsCountSize)
	binary.LittleEndian.PutUint32(buffer, uint32(len(index.meta)))

	if _, err := writer.Write(buffer); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrWriteMetaItemsCount, err)
	}
	// prepare meta items
	metaBuffer := bytes.NewBuffer(make([]byte, 0))
	enc := gob.NewEncoder(metaBuffer)

	for _, v := range index.meta {
		if err := enc.Encode(v); err != nil {
			return 0, fmt.Errorf("%w: %w", ErrEncodeMetaItem, err)
		}
	}

	// write meta part size
	binary.LittleEndian.PutUint32(buffer, uint32(metaBuffer.Len()))

	if _, err := writer.Write(buffer); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrWriteMetaPartSize, err)
	}

	// write meta part
	if _, err := writer.Write(metaBuffer.Bytes()); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrWriteMetaPart, err)
	}
	// write binary data
	if _, err := writer.Write(index.data); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrWriteBinaryData, err)
	}

	return writer.Count(), nil
}
