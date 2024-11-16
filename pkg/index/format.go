package index

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"encoding/gob"
	"encoding/json"
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

	ErrJSONEncode       = errors.New("failed to JSON encode item")
	ErrGZIPDecode       = errors.New("failed to decode gzip")
	ErrGZIPDecoderClose = errors.New("failed to close gzip decoder")
	ErrGZIPEncode       = errors.New("failed to encode gzip")
	ErrGZIPEncoderClose = errors.New("failed to close gzip encoder")

	ErrMetaDecode = errors.New("failed to decode meta data")
	ErrMetaEncode = errors.New("failed to encode meta data")
)

const indexHeader = "TINYTUNE_INDEX"
const metaItemsCountSize = 4

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

	if err := index.metaDecode(bytes.NewReader(metaPartBuffer), metaItemsCount); err != nil {
		return fmt.Errorf("%w: %w", ErrMetaDecode, err)
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

	if err := index.metaEncode(metaBuffer); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrMetaDecode, err)
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

func (index *Index) metaEncode(w io.Writer) error {
	gzipEncoder := gzip.NewWriter(w)
	jsonEncoder := json.NewEncoder(gzipEncoder)

	for _, v := range index.meta {
		err := jsonEncoder.Encode(v)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrJSONEncode, err)
		}
	}

	err := gzipEncoder.Close()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGZIPEncoderClose, err)
	}

	return nil
}

func (index *Index) metaDecode(r io.Reader, count uint32) error {
	gzipDecoder, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGZIPDecode, err)
	}

	jsonDecoder := json.NewDecoder(gzipDecoder)

	for range count {
		m := Meta{}
		if err := jsonDecoder.Decode(&m); err != nil {
			return fmt.Errorf("%w: %w", ErrMetaItemDecode, err)
		}

		index.meta[m.ID] = &m
	}

	err = gzipDecoder.Close()
	if err != nil {
		return fmt.Errorf("%w: %w", ErrGZIPDecoderClose, err)
	}

	return nil
}
