package internal

import (
	"crypto/sha256"
	"fmt"
	"io"
)

func SHA256Hash(reader io.Reader) (string, error) {
	hashAlgorithm := sha256.New()
	buf := make([]byte, 65536)
	for {
		switch n, err := reader.Read(buf); err {
		case nil:
			hashAlgorithm.Write(buf[:n])
		case io.EOF:
			return fmt.Sprintf("%x", hashAlgorithm.Sum(nil)), nil
		default:
			return "", err
		}
	}
}
