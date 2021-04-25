package astrolabe

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"gotest.tools/assert"
	"io"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

type DummyReaderAt struct {
	dummyData []byte
	closed *bool
	size int64
	offset * int64	// Offset in our stream
	mutex * sync.Mutex
}

func NewDummyReaderAt(dummyData []byte, size int64) DummyReaderAt {
	var offset int64
	var mutex sync.Mutex
	var closed bool
	return DummyReaderAt{
		dummyData: dummyData,
		size:size,
		offset: &offset,
		mutex: &mutex,
		closed: &closed,
	}
}

func (recv DummyReaderAt) ReadAt(buf []byte, off int64) (bytesRead int, err error) {
	if *recv.closed {
		err = errors.New("Closed")
	} else {
		// Tiles recv.dummyData into the buffer.  off and dummyData may not be alig*-
		bytesRead = 0
		if off + int64(len(buf)) > recv.size {
			err = io.EOF	// We're not going to return the number of bytes asked so set EOF here
		}
		if off < recv.size {
			dummyOff := int(off % int64(len(recv.dummyData)))
			if dummyOff > 0 {
				dummyBytes := len(recv.dummyData) - dummyOff
				if int64(dummyBytes)+off > recv.size {
					dummyBytes = int(recv.size - off)
				}
				bytesRead += copy(buf, recv.dummyData[dummyOff:dummyOff+dummyBytes])
			}
			for bytesRead+len(recv.dummyData) < len(buf) && // Not going to overflow the buffer with a full write of dummyData
				len(buf)-bytesRead >= len(recv.dummyData) && // Need a full write of dummyData
				off + int64(bytesRead + len(recv.dummyData)) < recv.size /* Not going past the end*/{
				bytesRead += copy(buf[bytesRead:], recv.dummyData)
			}
			if off + int64(bytesRead) < recv.size /* Not going past the end*/ {
				dummyEnd := len(buf)-bytesRead
				if off + int64(bytesRead + dummyEnd) > recv.size {
					dummyEnd = int(recv.size - (off + int64(bytesRead)))
				}
				bytesRead += copy(buf[bytesRead:], recv.dummyData[:dummyEnd])
			}
		}
	}
	return
}

func (recv DummyReaderAt) Read(buf []byte) (bytesRead int, err error) {
	recv.mutex.Lock()
	defer recv.mutex.Unlock()
	bytesRead, err = recv.ReadAt(buf, *recv.offset)
	*recv.offset = *recv.offset + int64(bytesRead)
	return
}

func (recv DummyReaderAt) Close() (err error) {
	recv.mutex.Lock()
	defer recv.mutex.Unlock()
	*recv.closed = true
	return
}


type DummyPE struct {
	info       ProtectedEntityInfoImpl
	mdSize     int32
	components [] ProtectedEntity
}

func (recv DummyPE) GetInfo(ctx context.Context) (ProtectedEntityInfo, error) {
	return recv.info, nil
}

func (recv DummyPE) GetCombinedInfo(ctx context.Context) ([]ProtectedEntityInfo, error) {
	panic("implement me")
}

func (recv DummyPE) Snapshot(ctx context.Context, params map[string]map[string]interface{}) (ProtectedEntitySnapshotID, error) {
	panic("implement me")
}

func (recv DummyPE) ListSnapshots(ctx context.Context) ([]ProtectedEntitySnapshotID, error) {
	panic("implement me")
}

func (recv DummyPE) DeleteSnapshot(ctx context.Context, snapshotToDelete ProtectedEntitySnapshotID, params map[string]map[string]interface{}) (bool, error) {
	panic("implement me")
}

func (recv DummyPE) GetInfoForSnapshot(ctx context.Context, snapshotID ProtectedEntitySnapshotID) (*ProtectedEntityInfo, error) {
	panic("implement me")
}

func (recv DummyPE) GetComponents(ctx context.Context) ([]ProtectedEntity, error) {
	return recv.components, nil
}

func (recv DummyPE) GetID() ProtectedEntityID {
	return recv.info.id
}

func (recv DummyPE) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	return NewDummyReaderAt([]byte("data-" + recv.GetID().String()), recv.info.size), nil
}

func (recv DummyPE) GetMetadataReader(ctx context.Context) (io.ReadCloser, error) {
	return NewDummyReaderAt([]byte("md-" + recv.GetID().String()), int64(recv.mdSize)), nil

}

func (recv DummyPE) Overwrite(ctx context.Context, sourcePE ProtectedEntity, params map[string]map[string]interface{}, overwriteComponents bool) error {
	panic("implement me")
}

func compareReaders(reader1 io.Reader, reader2 io.Reader, expectedBytes int64) (err error) {
	buf1 := make([]byte, 1024*1024)
	buf2 := make([]byte, 1024*1024)
	var offset int64
	for err == nil {
		var read1, read2 int
		read1, err = io.ReadFull(reader1, buf1)
		if err == nil || err == io.EOF || err == io.ErrUnexpectedEOF {
			read2, err = io.ReadFull(reader2, buf2)
			if err == nil || err == io.EOF || err == io.ErrUnexpectedEOF {
				if read1 != read2 {
					err = errors.New(fmt.Sprintf("Got differing amounts of data, read1 = %d, read2 = %d, offset = %d", read1, read2, offset))
				} else {
					if bytes.Compare(buf1[:read1], buf2[:read2]) != 0 {
						err = errors.New(fmt.Sprintf("Got different data, offset = %d", offset))
					}
					offset += int64(read1)
				}
			}
		}
	}
	if offset != expectedBytes {
		err = errors.New(fmt.Sprintf("Expected %d bytes, read %d", expectedBytes, offset))
	}
	if err == io.EOF || err == io.ErrUnexpectedEOF {
		err = nil		// EOF is expected
	}
	return
}

func TestZipUnZipSinglePE(t *testing.T) {
	ctx := context.Background()
	id, err := uuid.NewRandom()
	if err != nil {
		t.Errorf("Got err %v creating UUID", err)
	}
	name := "dummy-" + id.String()
	srcPE := DummyPE{
		info:       ProtectedEntityInfoImpl{
			id:   NewProtectedEntityID("dmy", id.String()),
			name: name,
			size: 1024*1024,
		},
		mdSize:     1024,
		components: nil,
	}

	file, err := ioutil.TempFile("", name + ".zip")
	if err != nil {
		t.Errorf("Got err %v creating temp file", err)
	}
	defer os.Remove(file.Name()) // clean up
	err = ZipProtectedEntityToWriter(ctx, srcPE, file)
	if err != nil {
		t.Errorf("Got err %v zipping PE to temp file", err)
	}

	fileInfo, err := file.Stat()
	if err != nil {
		t.Errorf("Got err %v getting stat for temp file", err)
	}
	destPE, err := GetPEFromZipStream(ctx, file, fileInfo.Size())
	if err != nil {
		t.Errorf("Got err %v unzipping PE from temp file", err)
	}
	assert.Equal(t, srcPE.GetID(), destPE.GetID())

	srcReader, err := srcPE.GetMetadataReader(ctx)
	if err != nil {
		t.Errorf("Got err %v retrieving srcReader", err)
	}
	destMDReader, err := destPE.GetMetadataReader(ctx)
	if err != nil {
		t.Errorf("Got err %v retrieving destMDReader", err)
	}
	err = compareReaders(srcReader, destMDReader, int64(srcPE.mdSize))
	if err != nil {
		t.Errorf("Metadata streams do not match, err = %v", err)
	}

	srcDataReader, err := srcPE.GetDataReader(ctx)
	if err != nil {
		t.Errorf("Got err %v retrieving srcReader", err)
	}
	destDataReader, err := destPE.GetDataReader(ctx)
	if err != nil {
		t.Errorf("Got err %v retrieving destMDReader", err)
	}
	err = compareReaders(srcDataReader, destDataReader, srcPE.info.size)
	if err != nil {
		t.Errorf("Data streams do not match, err = %v", err)
	}
	err = file.Close()
	if err != nil {
		t.Errorf("Got err %v closing temp file", err)
	}

}

