package util

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"sync"
)

/*
 A BlockSource is a data source that reads/writes in block size increments.
 */
type BlockSource interface {
	ReadAt(startSector uint64, numSectors uint64, buffer []byte) (sectorsRead uint64, err error)
	Capacity() int64
	BlockSize() int
	Close() error
}

/*
 Provides a ReaderAt stream from a block device
*/
type BlockSourceReader struct {
	offset      *int64
	offsetMutex *sync.Mutex // Lock to ensure that multiple-threads do not break offset or see the same data twice
	misalignedMutex *sync.Mutex	// Lock to make sure the multiple misaligned reads are not happening simultaneously
	logger      logrus.FieldLogger
	blockSource BlockSource
}

func NewBlockSourceReader(blockSource BlockSource, logger logrus.FieldLogger) BlockSourceReader {
	var offset int64
	offset = 0
	var offsetMutex, misalignedMutex sync.Mutex
	return BlockSourceReader{
		offset:      &offset,
		offsetMutex: &offsetMutex,
		misalignedMutex: &misalignedMutex,
		logger:      logger,
		blockSource: blockSource,
	}
}

func (recv BlockSourceReader) Read(p []byte) (n int, err error) {
	recv.offsetMutex.Lock()
	defer recv.offsetMutex.Unlock()
	bytesRead, err := recv.ReadAt(p, *recv.offset)
	*recv.offset += int64(bytesRead)
	recv.logger.Infof("Read returning %d, len(p) = %d, offset=%d\n", bytesRead, len(p), *recv.offset)
	return bytesRead, err
}

func (recv BlockSourceReader) ReadAt(p []byte, off int64) (n int, err error) {
	capacity := recv.blockSource.Capacity()
	if off >= capacity {
		return 0, io.EOF
	}
	// If we're being asked for a read beyond the end of the disk, slice the buffer down
	if off + int64(len(p)) > capacity {
		readLen := int32(capacity - off)
		p = p[0:readLen]
	}
	blockSize := recv.blockSource.BlockSize()
	blockSize64 := int64(blockSize)
	startSector := off / blockSize64
	var total = 0

	if !aligned(len(p), off, blockSize) {
		// Lock versus read and write of misaligned data so that read/modify/write cycle always gives correct
		// behavior (read/write is atomic even though misaligned)
		recv.misalignedMutex.Lock()
		defer recv.misalignedMutex.Unlock()
	}
	// Start missing aligned part
	if off%blockSize64 != 0 {
		tmpBuf := make([]byte, recv.blockSource.BlockSize())
		//err := disklib.Read(this.dli, (uint64)(startSector), 1, tmpBuf)
		sectorsRead, err := recv.blockSource.ReadAt((uint64)(startSector), 1, tmpBuf)
		if err != nil {
			return 0, err
		}
		if sectorsRead != 1 {
			return 0, errors.Errorf("Expected 1 sector, got %d", sectorsRead)
		}
		srcOff := int(off % blockSize64)
		count := recv.blockSource.BlockSize() - srcOff
		if count > len(p) {
			count = len(p)
		}
		srcEnd := srcOff + count
		tmpSlice := tmpBuf[srcOff:srcEnd]
		copy(p[:count], tmpSlice)
		startSector = startSector + 1
		total = total + count
	}
	// Middle aligned part
	numAlignedSectors := (len(p) - total) / blockSize
	if numAlignedSectors > 0 {
		desOff := total
		desEnd := total + numAlignedSectors*blockSize
		//err := disklib.Read(this.dli, (uint64)(startSector), (uint64)(numAlignedSectors), p[desOff:desEnd])
		sectorsRead, err := recv.blockSource.ReadAt((uint64)(startSector), (uint64)(numAlignedSectors), p[desOff:desEnd])
		if err != nil {
			return total, err
		}
		if sectorsRead != 1 {
			return total, errors.Errorf("Expected %d sector, got %d", numAlignedSectors, sectorsRead)
		}
		startSector = startSector + int64(numAlignedSectors)
		total = total + numAlignedSectors*blockSize
	}
	// End missing aligned part
	if (len(p) - total) > 0 {
		tmpBuf := make([]byte, blockSize)
		//err := disklib.Read(this.dli, (uint64)(startSector), 1, tmpBuf)
		sectorsRead, err := recv.blockSource.ReadAt((uint64)(startSector), 1, tmpBuf)
		if sectorsRead != 1 {
			return 0, errors.Errorf("Expected 1 sector, got %d", sectorsRead)
		}
		if err != nil {
			return total, err
		}
		count := len(p) - total
		srcEnd := count
		tmpSlice := tmpBuf[0:srcEnd]
		copy(p[total:], tmpSlice)
		total = total + count
	}
	return total, nil
}

func (recv BlockSourceReader) Close() error {
	return recv.blockSource.Close()
}

func aligned(len int, off int64, blockSize int) bool {
	return len % blockSize == 0 && off % int64(blockSize) == 0
}
