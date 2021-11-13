package framework

import (
	"context"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
)

type grpcReadCloser struct {
	readerID int64
	client generated.ProtectedEntityTypeManagerClient
}

func (recv grpcReadCloser) Read(buf []byte) (bytesRead int, err error) {
	readReq := generated.ReadRequest{
		ReaderID:    recv.readerID,
		BytesToRead: int64(len(buf)),
	}
	resp, err := recv.client.Read(context.Background(), &readReq)
	if err != nil {
		return 0, err
	}

	return copy(buf, resp.Data), nil
}


func (recv grpcReadCloser) Close() error {
	closeReq := generated.CloseRequest{
		ReaderID: recv.readerID,
	}
	_, err := recv.client.Close(context.Background(), &closeReq)
	return err
}

type grpcReadAtCloser struct {
	grpcReadCloser
}

func (recv grpcReadAtCloser) ReadAt(buf []byte, offset int64) (bytesRead int, err error) {
	readReq := generated.ReadAtRequest{
		ReaderID:    recv.readerID,
		BytesToRead: int64(len(buf)),
		Offset: offset,
	}
	resp, err := recv.client.ReadAt(context.Background(), &readReq)
	if err != nil {
		return 0, err
	}

	return copy(buf, resp.Data), nil
}

// Needs to match ProtectedEntityTypeManager_ReadClient and ProtectedEntityTypeManager_ReadAtClient
type receiver interface {
	Recv() (*generated.ReadResponse, error)
}