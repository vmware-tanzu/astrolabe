/*
 * Copyright the Astrolabe contributors
 * SPDX-License-Identifier: Apache-2.0
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */
package framework

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
	"io"
	"sync/atomic"
)

type protectedEntityTypeManagerServer struct {
	generated.UnimplementedProtectedEntityTypeManagerServer
	logger   *logrus.Logger
	initFunc astrolabe.InitFunc
	petm     astrolabe.ProtectedEntityTypeManager
	readerID int64
	readers  map[int64]io.ReadCloser
}

func NewProtectedEntityTypeManagerServer(initFunc astrolabe.InitFunc) (generated.ProtectedEntityTypeManagerServer, error) {
	logger := NewLogger()
	server := protectedEntityTypeManagerServer{
		logger:   logger,
		initFunc: initFunc,
		readers:  map[int64]io.ReadCloser{},
	}
	return &server, nil
}

func (recv *protectedEntityTypeManagerServer) Init(ctx context.Context, request *generated.InitRequest) (*generated.Empty, error) {
	var params map[string]interface{}
	err := json.Unmarshal([]byte(request.ConfigInfo), &params)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal params from string %s\n", request.ConfigInfo)
	}
	var s3Config astrolabe.S3Config
	err = json.Unmarshal([]byte(request.S3Config), &s3Config)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal s3 config from string %s\n", request.S3Config)
	}

	petm, err := recv.initFunc(params, s3Config, recv.logger)
	if err != nil {
		return nil, errors.Wrapf(err, "error initializing protected entity type manager")
	}
	recv.petm = petm
	return &generated.Empty{}, nil
}

func (recv *protectedEntityTypeManagerServer) GetTypeName(ctx context.Context, empty *generated.Empty) (*generated.GetTypeNameResponse, error) {
	gtrn := generated.GetTypeNameResponse{
		TypeName: recv.petm.GetTypeName(),
	}
	return &gtrn, nil
}

func (recv *protectedEntityTypeManagerServer) GetProtectedEntities(ctx context.Context, empty *generated.Empty) (*generated.GetProtectedEntitiesResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	peIDs, err := recv.petm.GetProtectedEntities(ctx)
	if err != nil {
		return nil, err
	}
	peIDPtrs := make([]*generated.ProtectedEntityID, len(peIDs))
	for index, curPEID := range peIDs {
		peIDPtrs[index] = &generated.ProtectedEntityID{
			PeType:     curPEID.GetPeType(),
			Id:         curPEID.GetID(),
			SnapshotID: &generated.ProtectedEntitySnapshotID{Id: curPEID.GetSnapshotID().GetID()},
		}
	}
	gper := generated.GetProtectedEntitiesResponse{
		Ids: peIDPtrs,
	}
	return &gper, nil
}

func (recv *protectedEntityTypeManagerServer) CopyFromInfo(ctx context.Context, request *generated.CopyFromInfoRequest) (*generated.CopyFromInfoResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	panic("implement me")
}

func (recv *protectedEntityTypeManagerServer) getProtectedEntity(ctx context.Context, id *generated.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	peid := ProtectedEntityIDFromProto(id)
	pe, err := recv.petm.GetProtectedEntity(ctx, peid)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve protected entity id %v", peid)
	}
	return pe, nil
}

func (recv *protectedEntityTypeManagerServer) GetInfo(ctx context.Context, id *generated.ProtectedEntityID) (*generated.ProtectedEntityInfo, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	peInfo, err := pe.GetInfo(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve protected entity info for %v", ProtectedEntityIDFromProto(id))
	}
	return ProtoFromProtectedEntityInfo(peInfo), nil
}

func (recv *protectedEntityTypeManagerServer) GetCombinedInfo(ctx context.Context, id *generated.ProtectedEntityID) (*generated.GetCombinedInfoResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	combInfo, err := pe.GetCombinedInfo(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve protected entity combined info for %v", ProtectedEntityIDFromProto(id))
	}
	protoInfo := make([]*generated.ProtectedEntityInfo, len(combInfo))
	for index, curInfo := range combInfo {
		protoInfo[index] = ProtoFromProtectedEntityInfo(curInfo)
	}
	return &generated.GetCombinedInfoResponse{Info: protoInfo}, nil
}

func (recv *protectedEntityTypeManagerServer) Snapshot(ctx context.Context, request *generated.SnapshotRequest) (*generated.ProtectedEntitySnapshotID, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	params := map[string]map[string]interface{}{}
	err = json.Unmarshal([]byte(request.Params), &params)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal params %s", request.Params)
	}
	snapshotID, err := pe.Snapshot(ctx, params)
	if err != nil {
		return nil, err
	}
	return ProtoFromProtectedEntitySnapshotID(snapshotID), nil
}

func (recv *protectedEntityTypeManagerServer) ListSnapshots(ctx context.Context, id *generated.ProtectedEntityID) (*generated.ListSnapshotsResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	snapshotIDs, err := pe.ListSnapshots(ctx)
	if err != nil {
		return nil, err
	}
	protoSnapshotIDs := make([]*generated.ProtectedEntitySnapshotID, len(snapshotIDs))
	for index, curSnapshotID := range snapshotIDs {
		protoSnapshotIDs[index] = ProtoFromProtectedEntitySnapshotID(curSnapshotID)
	}
	return &generated.ListSnapshotsResponse{SnapshotIDs: protoSnapshotIDs}, nil
}

func (recv *protectedEntityTypeManagerServer) DeleteSnapshot(ctx context.Context, request *generated.DeleteSnapshotRequest) (*generated.DeleteSnapshotResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	params := map[string]map[string]interface{}{}
	err = json.Unmarshal([]byte(request.Params), &params)
	if err != nil {
		return nil, errors.Wrapf(err, "could not unmarshal params %s", request.Params)
	}
	success, err := pe.DeleteSnapshot(ctx, ProtectedEntitySnapshotIDFromProto(request.SnapshotID), params)
	if err != nil {
		return nil, err
	}
	return &generated.DeleteSnapshotResponse{Success: success}, nil
}

func (recv *protectedEntityTypeManagerServer) GetInfoForSnapshot(ctx context.Context, request *generated.GetInfoForSnapshotRequest) (*generated.ProtectedEntityInfo, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, request.Id)
	if err != nil {
		return nil, err
	}
	peInfo, err := pe.GetInfoForSnapshot(ctx, ProtectedEntitySnapshotIDFromProto(request.SnapshotID))
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve protected entity info for id %v snapshot id %v",
			ProtectedEntityIDFromProto(request.Id), ProtectedEntitySnapshotIDFromProto(request.SnapshotID))
	}
	return ProtoFromProtectedEntityInfo(*peInfo), nil
}

func (recv *protectedEntityTypeManagerServer) GetComponents(ctx context.Context, id *generated.ProtectedEntityID) (*generated.GetComponentsResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	components, err := pe.GetComponents(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "could not retrieve components for %v", ProtectedEntityIDFromProto(id))
	}
	protoComponentIDs := make([]*generated.ProtectedEntityID, len(components))
	for index, curComponent := range components {
		protoComponentIDs[index] = ProtoFromProtectedEntityID(curComponent.GetID())
	}
	return &generated.GetComponentsResponse{Components: protoComponentIDs}, nil
}

func (recv *protectedEntityTypeManagerServer) GetMetadataReader(ctx context.Context, id *generated.ProtectedEntityID) (*generated.ReaderResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	mdReader, err := pe.GetMetadataReader(ctx)
	if err != nil {
		return nil, err
	}
	readerID := recv.allocateReaderIDForReader(mdReader)
	_, isReaderAt := mdReader.(io.ReaderAt)
	return &generated.ReaderResponse{
		ReaderID:   readerID,
		IsReaderAt: isReaderAt,
	}, nil
}

func (recv *protectedEntityTypeManagerServer) GetDataReader(ctx context.Context, id *generated.ProtectedEntityID) (*generated.ReaderResponse, error) {
	if recv.petm == nil {
		return nil, errors.New("Init must be called for plugin")
	}
	pe, err := recv.getProtectedEntity(ctx, id)
	if err != nil {
		return nil, err
	}
	mdReader, err := pe.GetDataReader(ctx)
	if err != nil {
		return nil, err
	}
	readerID := recv.allocateReaderIDForReader(mdReader)
	_, isReaderAt := mdReader.(io.ReaderAt)
	return &generated.ReaderResponse{
		ReaderID:   readerID,
		IsReaderAt: isReaderAt,
	}, nil
}

func (recv *protectedEntityTypeManagerServer) allocateReaderIDForReader(reader io.ReadCloser) int64 {
	readerID := atomic.AddInt64(&recv.readerID, 1)
	recv.readers[readerID] = reader
	return readerID
}

func (recv *protectedEntityTypeManagerServer) releaseReaderID(readerID int64) {
	delete(recv.readers, readerID)
}

func (recv *protectedEntityTypeManagerServer) Read(ctx context.Context, request *generated.ReadRequest) (*generated.ReadResponse, error) {
	reader, ok := recv.readers[request.ReaderID]
	if !ok {
		return nil, errors.Errorf("reader %d not found", request.ReaderID)
	}
	buf := make([]byte, request.BytesToRead)
	bytesRead, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}
	readResponse := generated.ReadResponse{
		Data: buf[0:bytesRead],
	}
	return &readResponse, nil
}

func (recv *protectedEntityTypeManagerServer) ReadAt(ctx context.Context, request *generated.ReadAtRequest) (*generated.ReadResponse, error) {
	reader, ok := recv.readers[request.ReaderID]
	if !ok {
		return nil, errors.Errorf("reader %d not found", request.ReaderID)
	}
	readerAt, ok := reader.(io.ReaderAt)
	if !ok {
		return nil, errors.Errorf("reader %d is not a ReaderAt")
	}
	offset := request.Offset
	buf := make([]byte, request.BytesToRead)
	bytesRead, err := readerAt.ReadAt(buf, offset)
	if err != nil {
		return nil, err
	}
	readResponse := generated.ReadResponse{
		Data: buf[0:bytesRead],
	}
	return &readResponse, nil
}

func (recv *protectedEntityTypeManagerServer) Close(ctx context.Context, request *generated.CloseRequest) (*generated.Empty, error) {
	reader, ok := recv.readers[request.ReaderID]
	if ok {
		recv.releaseReaderID(request.ReaderID)
		return &generated.Empty{}, reader.Close()
	}
	return &generated.Empty{}, nil
}
