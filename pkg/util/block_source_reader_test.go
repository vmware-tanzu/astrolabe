package util

import (
	"encoding/binary"
	"github.com/sirupsen/logrus"
	"testing"
)

type mockBlockSource struct {
	blockCapacity uint64
	blockSize int
}

func (recv mockBlockSource) ReadAt(startBlock uint64, numBlocks uint64, buffer []byte) (sectorsRead uint64, err error) {
	bufferOffset := 0
	for curBlock := 0; curBlock < int(numBlocks); curBlock++ {
		for curIntOffset := 0; curIntOffset < recv.blockSize/4; curIntOffset ++ {
			curByteOffset := curIntOffset * 4 + bufferOffset
			binary.BigEndian.PutUint32(buffer[curByteOffset:curByteOffset+4], uint32(curBlock))
		}
	}
	return numBlocks, nil
}

func (recv mockBlockSource) Capacity() int64 {
	return int64(recv.blockCapacity * uint64(recv.blockSize))
}

func (recv mockBlockSource) BlockSize() int {
	return recv.blockSize
}

func (recv mockBlockSource) Close() error {
	// no-op
	return nil
}

func TestSingleCharRead(t *testing.T) {
	mbs := mockBlockSource{
		blockCapacity: 1,
		blockSize:     512,
	}

	bsr := NewBlockSourceReader(mbs, logrus.New())
	for byteNum := 0; byteNum < 512; byteNum++ {
		buf := make([]byte, 1)
		bytesRead, err := bsr.Read(buf)
		if err != nil {
			t.Fatal(err)
		}
		if bytesRead != 1 {
			t.Fatalf("Expected 1 byte, %d", bytesRead)
		}
		if buf[0] != 0 { // mockBlockSource rwrites the block numbers as the data, so will always be 0
			t.Fatalf("Expected data to be 0 was, %d", buf[0])
		}
	}
}

