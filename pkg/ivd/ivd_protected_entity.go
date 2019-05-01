/*
 * Copyright 2019 VMware, Inc..
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

package ivd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	vim "github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/xml"
	"github.com/vmware/gvddk/gDiskLib"
	gvddk_high "github.com/vmware/gvddk/gvddk-high"
	"io"
	"io/ioutil"
	"strings"

	//	"github.com/vmware/govmomi/vslm"
	"context"
	"time"
)

type IVDProtectedEntity struct {
	ipetm    *IVDProtectedEntityTypeManager
	id       astrolabe.ProtectedEntityID
	data     []astrolabe.DataTransport
	metadata []astrolabe.DataTransport
	combined []astrolabe.DataTransport
	logger   logrus.FieldLogger
}

type metadata struct {
	VirtualStorageObject vim.VStorageObject         `xml:"virtualStorageObject"`
	Datastore            vim.ManagedObjectReference `xml:"datastore"`
	ExtendedMetadata     []vim.KeyValue          `xml:"extendedMetadata"`
}

func (this IVDProtectedEntity) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	diskConnectParam, err := this.getDiskConnectionParams(ctx, true)
	if err != nil {
		return nil, err
	}

	diskReader, vErr := gvddk_high.Open(diskConnectParam, this.logger)
	if vErr != nil {
		return nil, errors.New(fmt.Sprintf(vErr.Error() + " with error code: %d", vErr.VixErrorCode()))
	}

	return diskReader, nil
}


func (this IVDProtectedEntity) copy(ctx context.Context, dataReader io.Reader,
	metadata metadata) error {
	// TODO - restore metadata
	dataWriter, err := this.getDataWriter(ctx)
	if dataWriter != nil {
		defer func() {
			if err := dataWriter.Close(); err != nil {
				this.logger.Errorf("The deferred data writer is closed with error, %v", err)
			}
		}()
	}

	if err != nil {
		return err
	}

	bufferedWriter := bufio.NewWriterSize(dataWriter, 1024*1024)
	buf := make([]byte, 1024*1024)
	_, err = io.CopyBuffer(bufferedWriter, dataReader, buf) // TODO - add a copy routine that we can interrupt via context

	return err
}

func (this IVDProtectedEntity) getDataWriter(ctx context.Context) (io.WriteCloser, error) {
	diskConnectParam, err := this.getDiskConnectionParams(ctx, false)
	if err != nil {
		return nil, err
	}

	diskWriter, vErr := gvddk_high.Open(diskConnectParam, this.logger)
	if vErr != nil {
		return nil, errors.New(fmt.Sprintf(vErr.Error() + " with error code: %d", vErr.VixErrorCode()))
	}

	return diskWriter, nil
}

func (this IVDProtectedEntity) getDiskConnectionParams(ctx context.Context, readOnly bool) (gDiskLib.ConnectParams, error) {
	url := this.ipetm.client.URL()
	serverName := url.Hostname()
	userName := this.ipetm.user
	password := this.ipetm.password
	fcdId := this.id.GetID()
	vso, err := this.ipetm.vsom.Retrieve(context.Background(), NewVimIDFromPEID(this.id))
	if err != nil {
		//return gDiskLib.DiskHandle{}, err
		return gDiskLib.ConnectParams{}, err
	}
	datastore := vso.Config.Backing.GetBaseConfigInfoBackingInfo().Datastore.String()
	datastore = strings.TrimPrefix(datastore, "Datastore:")

	fcdssid := ""
	if this.id.HasSnapshot() {
		fcdssid = this.id.GetSnapshotID().String()
	}
	path := ""
	var flags uint32
	if readOnly {
		flags = gDiskLib.VIXDISKLIB_FLAG_OPEN_COMPRESSION_SKIPZ | gDiskLib.VIXDISKLIB_FLAG_OPEN_READ_ONLY
	} else {
		flags = gDiskLib.VIXDISKLIB_FLAG_OPEN_UNBUFFERED
	}
	transportMode := "nbd"
	thumbPrint, err := gDiskLib.GetThumbPrintForURL(*url)
	if err != nil {
		this.logger.Errorf("Failed to get the thumb print for the URL, %s", url.String())
		return gDiskLib.ConnectParams{}, err
	}

	params := gDiskLib.NewConnectParams("",
		serverName,
		thumbPrint,
		userName,
		password,
		fcdId,
		datastore,
		fcdssid,
		"",
		"vm-example",
		path,
		flags,
		readOnly,
		transportMode)

	return params, nil
}

func (this IVDProtectedEntity) GetMetadataReader(ctx context.Context) (io.Reader, error) {
	infoBuf, err := this.getMetadataBuf(ctx)
	if err != nil {
		return nil, err
	}

	return bytes.NewReader(infoBuf), nil
}

func (this IVDProtectedEntity) getMetadataBuf(ctx context.Context) ([]byte, error) {
	md, err := this.getMetadata(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Retrieve failed")
	}
	retBuf, err := xml.MarshalIndent(md, "  ", "    ")
	if err == nil {
		fmt.Println(string(retBuf))
	}
	return retBuf, err
}

func (this IVDProtectedEntity) getMetadata(ctx context.Context) (metadata, error) {
	vsoID := vim.ID{
		Id: this.id.GetID(),
	}
	vso, err := this.ipetm.vsom.Retrieve(ctx, vsoID)
	if err != nil {
		return metadata{}, err
	}
	datastore := vso.Config.BaseConfigInfo.GetBaseConfigInfo().Backing.GetBaseConfigInfoBackingInfo().Datastore
	var ssID *vim.ID = nil
	if (this.id.HasSnapshot()) {

		ssID = &vim.ID{
			Id: this.id.GetSnapshotID().GetID(),
		}
	}
	extendedMetadata, err := this.ipetm.vsom.RetrieveMetadata(ctx, vsoID, ssID, "")

	retVal := metadata{
		VirtualStorageObject: *vso,
		Datastore:            datastore,
		ExtendedMetadata:     extendedMetadata,
	}
	return retVal, nil
}

func readMetadataFromReader(ctx context.Context, metadataReader io.Reader) (metadata, error) {
	mdBuf, err := ioutil.ReadAll(metadataReader) // TODO - limit this so it can't run us out of memory here
	if err != nil {
		return metadata{}, err
	}
	return readMetadataFromBuf(ctx, mdBuf)
}

func readMetadataFromBuf(ctx context.Context, buf []byte) (metadata, error) {
	var retVal = metadata{}
	err := xml.Unmarshal(buf, &retVal)
	return retVal, err
}

func newProtectedEntityID(id vim.ID) astrolabe.ProtectedEntityID {
	return astrolabe.NewProtectedEntityID("ivd", id.Id)
}

func newProtectedEntityIDWithSnapshotID(id vim.ID, snapshotID astrolabe.ProtectedEntitySnapshotID) astrolabe.ProtectedEntityID {
	return astrolabe.NewProtectedEntityIDWithSnapshotID("ivd", id.Id, snapshotID)
}

func newIVDProtectedEntity(ipetm *IVDProtectedEntityTypeManager, id astrolabe.ProtectedEntityID) (IVDProtectedEntity, error) {
	data, metadata, combined, err := ipetm.getDataTransports(id)
	if err != nil {
		return IVDProtectedEntity{}, err
	}
	newIPE := IVDProtectedEntity{
		ipetm:    ipetm,
		id:       id,
		data:     data,
		metadata: metadata,
		combined: combined,
		logger:   ipetm.logger,
	}
	return newIPE, nil
}
func (this IVDProtectedEntity) GetInfo(ctx context.Context) (astrolabe.ProtectedEntityInfo, error) {
	vsoID := vim.ID{
		Id: this.id.GetID(),
	}
	vso, err := this.ipetm.vsom.Retrieve(ctx, vsoID)
	if err != nil {
		return nil, errors.Wrap(err, "Retrieve failed")
	}

	retVal := astrolabe.NewProtectedEntityInfo(
		this.id,
		vso.Config.Name,
		this.data,
		this.metadata,
		this.combined,
		[]astrolabe.ProtectedEntityID{})
	return retVal, nil
}

func (this IVDProtectedEntity) GetCombinedInfo(ctx context.Context) ([]astrolabe.ProtectedEntityInfo, error) {
	ivdIPE, err := this.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	return []astrolabe.ProtectedEntityInfo{ivdIPE}, nil
}

const waitTime = 3600 * time.Second

/*
 * Snapshot APIs
 */
func (this IVDProtectedEntity) Snapshot(ctx context.Context) (astrolabe.ProtectedEntitySnapshotID, error) {
	vslmTask, err := this.ipetm.vsom.CreateSnapshot(ctx, NewVimIDFromPEID(this.GetID()), "AstrolabeSnapshot")
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrap(err, "Snapshot failed")
	}
	ivdSnapshotIDAny, err := vslmTask.Wait(ctx, waitTime)
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrap(err, "Wait failed")
	}
	ivdSnapshotID := ivdSnapshotIDAny.(vim.ID)
	/*
		ivdSnapshotStr := ivdSnapshotIDAny.(string)
		ivdSnapshotID := vim.ID{
			id: ivdSnapshotStr,
		}
	*/
	retVal := astrolabe.NewProtectedEntitySnapshotID(ivdSnapshotID.Id)
	return retVal, nil
}

func (this IVDProtectedEntity) ListSnapshots(ctx context.Context) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	snapshotInfo, err := this.ipetm.vsom.RetrieveSnapshotInfo(ctx, NewVimIDFromPEID(this.GetID()))
	if err != nil {
		return nil, errors.Wrap(err, "RetrieveSnapshotInfo failed")
	}
	peSnapshotIDs := []astrolabe.ProtectedEntitySnapshotID{}
	for _, curSnapshotInfo := range snapshotInfo {
		peSnapshotIDs = append(peSnapshotIDs, astrolabe.NewProtectedEntitySnapshotID(curSnapshotInfo.Id.Id))
	}
	return peSnapshotIDs, nil
}
func (this IVDProtectedEntity) DeleteSnapshot(ctx context.Context, snapshotToDelete astrolabe.ProtectedEntitySnapshotID) (bool, error) {
	vslmTask, err := this.ipetm.vsom.DeleteSnapshot(ctx, NewVimIDFromPEID(this.id), NewVimSnapshotIDFromPESnapshotID(snapshotToDelete))
	this.logger.Infof("IVD DeleteSnapshot API get called. Error: %v", err)
	if err != nil {
		return false, errors.Wrap(err, "DeleteSnapshot failed")
	}
	_, err = vslmTask.Wait(ctx, waitTime)
	this.logger.Infof("IVD DeleteSnapshot Task is reported to be completed. Error: %v", err)
	if err != nil {
		return false, errors.Wrap(err, "Wait failed")
	}
	return true, nil
}

func (this IVDProtectedEntity) GetInfoForSnapshot(ctx context.Context, snapshotID astrolabe.ProtectedEntitySnapshotID) (*astrolabe.ProtectedEntityInfo, error) {
	return nil, nil
}

func (this IVDProtectedEntity) GetComponents(ctx context.Context) ([]astrolabe.ProtectedEntity, error) {
	return make([]astrolabe.ProtectedEntity, 0), nil
}

func (this IVDProtectedEntity) GetID() astrolabe.ProtectedEntityID {
	return this.id
}

func NewIDFromString(idStr string) vim.ID {
	return vim.ID{
		Id: idStr,
	}
}

func NewVimIDFromPEID(peid astrolabe.ProtectedEntityID) vim.ID {
	if peid.GetPeType() == "ivd" {
		return vim.ID{
			Id: peid.GetID(),
		}
	} else {
		return vim.ID{}
	}
}

func NewVimSnapshotIDFromPEID(peid astrolabe.ProtectedEntityID) vim.ID {
	if peid.HasSnapshot() {
		return vim.ID{
			Id: peid.GetSnapshotID().GetID(),
		}
	} else {
		return vim.ID{}
	}
}

func NewVimSnapshotIDFromPESnapshotID(peSnapshotID astrolabe.ProtectedEntitySnapshotID) vim.ID {
	return vim.ID{
		Id: peSnapshotID.GetID(),
	}
}
