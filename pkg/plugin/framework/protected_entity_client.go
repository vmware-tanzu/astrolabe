package framework

import (
	"bufio"
	"context"
	json "github.com/json-iterator/go"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
	"io"
)

type protectedEntityClient struct {
	id     astrolabe.ProtectedEntityID
	client generated.ProtectedEntityTypeManagerClient
	petm   *protectedEntityTypeManagerClient
}

func (recv protectedEntityClient) GetInfo(ctx context.Context) (astrolabe.ProtectedEntityInfo, error) {
	protoPEID := ProtoFromProtectedEntityID(recv.id)
	protoPEInfo, err := recv.client.GetInfo(ctx, protoPEID)
	if err != nil {
		return nil, err
	}
	return ProtectedEntityInfoFromProto(protoPEInfo), nil
}

func (recv protectedEntityClient) GetCombinedInfo(ctx context.Context) ([]astrolabe.ProtectedEntityInfo, error) {
	protoPEID := ProtoFromProtectedEntityID(recv.id)
	gcir, err := recv.client.GetCombinedInfo(ctx, protoPEID)
	if err != nil {
		return nil, err
	}
	retPEInfo := make([]astrolabe.ProtectedEntityInfo, len(gcir.Info))
	for index, curProtoPEInfo := range gcir.Info {
		retPEInfo[index] = ProtectedEntityInfoFromProto(curProtoPEInfo)
	}
	return retPEInfo, nil
}

func (recv protectedEntityClient) Snapshot(ctx context.Context, params map[string]map[string]interface{}) (astrolabe.ProtectedEntitySnapshotID, error) {
	paramBytes, err := json.Marshal(params)
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, err
	}

	req := generated.SnapshotRequest{
		Id:     ProtoFromProtectedEntityID(recv.id),
		Params: string(paramBytes),
	}

	resp, err := recv.client.Snapshot(ctx, &req)
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, err
	}

	return ProtectedEntitySnapshotIDFromProto(resp), nil
}

func (recv protectedEntityClient) ListSnapshots(ctx context.Context) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	resp, err := recv.client.ListSnapshots(ctx, ProtoFromProtectedEntityID(recv.id))
	if err != nil {
		return nil, err
	}

	retIDs := make([]astrolabe.ProtectedEntitySnapshotID, len(resp.SnapshotIDs))
	for index, curProtoSnapshotID := range resp.SnapshotIDs {
		retIDs[index] = ProtectedEntitySnapshotIDFromProto(curProtoSnapshotID)
	}
	return retIDs, nil
}

func (recv protectedEntityClient) DeleteSnapshot(ctx context.Context, snapshotToDelete astrolabe.ProtectedEntitySnapshotID, params map[string]map[string]interface{}) (bool, error) {
	paramBytes, err := json.Marshal(params)
	if err != nil {
		return false, err
	}

	req := generated.DeleteSnapshotRequest{
		Id:     ProtoFromProtectedEntityID(recv.id),
		Params: string(paramBytes),
	}

	resp, err := recv.client.DeleteSnapshot(ctx, &req)
	if err != nil {
		return false, err
	}

	return resp.Success, nil
}

func (recv protectedEntityClient) GetInfoForSnapshot(ctx context.Context, snapshotID astrolabe.ProtectedEntitySnapshotID) (*astrolabe.ProtectedEntityInfo, error) {
	req := generated.GetInfoForSnapshotRequest{
		Id:         ProtoFromProtectedEntityID(recv.id),
		SnapshotID: ProtoFromProtectedEntitySnapshotID(snapshotID),
	}
	resp, err := recv.client.GetInfoForSnapshot(ctx, &req)
	if err != nil {
		return nil, err
	}
	peInfo := ProtectedEntityInfoFromProto(resp)
	return &peInfo, nil
}

func (recv protectedEntityClient) GetComponents(ctx context.Context) ([]astrolabe.ProtectedEntity, error) {
	// TODO - should probably change APIs here to return component IDs
	panic("implement me")
}

func (recv protectedEntityClient) GetID() astrolabe.ProtectedEntityID {
	return recv.id
}

func (recv protectedEntityClient) getGRPCReader(resp *generated.ReaderResponse) (io.ReadCloser, error) {
	readCloser := grpcReadCloser{
		readerID: resp.ReaderID,
		client:   recv.client,
	}
	if resp.IsReaderAt {
		return grpcReadAtCloser{readCloser}, nil
	}
	return readCloser, nil
}

type bufferedReadCloser struct {
	io.Reader
	io.Closer
}

func (recv protectedEntityClient) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	resp, err := recv.client.GetDataReader(ctx, ProtoFromProtectedEntityID(recv.id))
	if err != nil {
		return nil, err
	}
	grpcReader, err := recv.getGRPCReader(resp)
	if err != nil {
		return nil, err
	}
	return &bufferedReadCloser{bufio.NewReaderSize(grpcReader, 1024*1024), grpcReader}, nil
}

func (recv protectedEntityClient) GetMetadataReader(ctx context.Context) (io.ReadCloser, error) {
	resp, err := recv.client.GetDataReader(ctx, ProtoFromProtectedEntityID(recv.id))
	if err != nil {
		return nil, err
	}
	return recv.getGRPCReader(resp)
}

func (recv protectedEntityClient) Overwrite(ctx context.Context, sourcePE astrolabe.ProtectedEntity, params map[string]map[string]interface{}, overwriteComponents bool) error {
	panic("implement me")
}
