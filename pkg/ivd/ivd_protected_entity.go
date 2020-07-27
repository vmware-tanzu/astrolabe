/*
 * Copyright 2019 the Astrolabe contributors
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
	"github.com/vmware/govmomi/vim25/soap"
	vim "github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vim25/xml"
	"github.com/vmware/gvddk/gDiskLib"
	gvddk_high "github.com/vmware/gvddk/gvddk-high"
	"io"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/wait"
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
	ExtendedMetadata     []vim.KeyValue             `xml:"extendedMetadata"`
}

func (this IVDProtectedEntity) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	diskConnectParam, err := this.getDiskConnectionParams(ctx, true)
	if err != nil {
		return nil, err
	}

	diskReader, vErr := gvddk_high.Open(diskConnectParam, this.logger)
	if vErr != nil {
		if vErr.VixErrorCode() == 20005 {
			// Try the end access once
			gDiskLib.EndAccess(diskConnectParam)
			diskReader, vErr = gvddk_high.Open(diskConnectParam, this.logger)
		}
		if vErr != nil {
			return nil, errors.New(fmt.Sprintf(vErr.Error()+" with error code: %d", vErr.VixErrorCode()))
		}
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
		return nil, errors.New(fmt.Sprintf(vErr.Error()+" with error code: %d", vErr.VixErrorCode()))
	}

	return diskWriter, nil
}

func (this IVDProtectedEntity) getDiskConnectionParams(ctx context.Context, readOnly bool) (gDiskLib.ConnectParams, error) {
	url := this.ipetm.client.URL()
	serverName := url.Hostname()
	userName, err := GetUserFromParamsMap(this.ipetm.vcParams)
	if err != nil {
		return gDiskLib.ConnectParams{}, err
	}
	password, err := GetPasswordFromParamsMap(this.ipetm.vcParams)
	if err != nil {
		return gDiskLib.ConnectParams{}, err
	}
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

func (this IVDProtectedEntity) GetMetadataReader(ctx context.Context) (io.ReadCloser, error) {
	infoBuf, err := this.getMetadataBuf(ctx)
	if err != nil {
		return nil, err
	}

	return ioutil.NopCloser(bytes.NewReader(infoBuf)), nil
}

func (this IVDProtectedEntity) getMetadataBuf(ctx context.Context) ([]byte, error) {
	md, err := this.getMetadata(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Retrieve failed")
	}
	retBuf, err := xml.MarshalIndent(md, "  ", "    ")
	if err != nil {
		return nil, err
	}
	return retBuf, nil
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
	if this.id.HasSnapshot() {

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
	if err != nil && err != io.EOF {
		return metadata{}, errors.Wrap(err, "ReadAll failed")
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
func (this IVDProtectedEntity) Snapshot(ctx context.Context, params map[string]map[string]interface{}) (astrolabe.ProtectedEntitySnapshotID, error) {
	this.logger.Infof("CreateSnapshot called on IVD Protected Entity, %v", this.id.String())
	var retVal astrolabe.ProtectedEntitySnapshotID
	err := wait.PollImmediate(time.Second, time.Hour, func() (bool, error) {
		this.logger.Infof("Retrying CreateSnapshot on IVD Protected Entity, %v, for one hour at the maximum", this.GetID().String())
		vslmTask, err := this.ipetm.vsom.CreateSnapshot(ctx, NewVimIDFromPEID(this.GetID()), "AstrolabeSnapshot")
		if err != nil {
			return false, errors.Wrapf(err, "Failed to create a task for the CreateSnapshot invocation on IVD Protected Entity, %v", this.id.String())
		}
		ivdSnapshotIDAny, err := vslmTask.Wait(ctx, waitTime)
		if err != nil {
			if soap.IsVimFault(err) {
				_, ok := soap.ToVimFault(err).(*vim.InvalidState)
				if ok {
					this.logger.WithError(err).Error("There is some operation, other than this CreateSnapshot invocation, on the VM attached still being protected by its VM state. Will retry")
					return false, nil
				}
			}
			return false, errors.Wrapf(err, "Failed at waiting for the CreateSnapshot invocation on IVD Protected Entity, %v", this.id.String())
		}
		ivdSnapshotID := ivdSnapshotIDAny.(vim.ID)
		this.logger.Debugf("A new snapshot, %v, was created on IVD Protected Entity, %v", ivdSnapshotID.Id, this.GetID().String())

		// Will try RetrieveSnapshotDetail right after the completion of CreateSnapshot to make sure there is no impact from race condition
		_, err = this.ipetm.vsom.RetrieveSnapshotDetails(ctx, NewVimIDFromPEID(this.GetID()), ivdSnapshotID)
		if err != nil {
			if soap.IsSoapFault(err) {
				faultMsg := soap.ToSoapFault(err).String
				if strings.Contains(faultMsg, "A specified parameter was not correct: snapshotId") {
					this.logger.WithError(err).Error("Unexpected InvalidArgument SOAP fault due to the known race condition. Will retry")
					return false, nil
				}
				this.logger.WithError(err).Error("Unexpected SOAP fault")
			}
			return false, errors.Wrapf(err, "Failed at retrieving the snapshot detail post the completion of CreateSnapshot invocation on, %v", this.id.String())
		}
		this.logger.Debugf("The retrieval of the newly created snapshot, %v, is completed successfully", ivdSnapshotID.Id)

		retVal = astrolabe.NewProtectedEntitySnapshotID(ivdSnapshotID.Id)
		return true, nil
	})

	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, err
	}
	this.logger.Infof("CreateSnapshot completed on IVD Protected Entity, %v, with the new snapshot, %v, being created", this.id.String(), retVal.String())
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
	this.logger.Infof("DeleteSnapshot called on IVD Protected Entity, %v, with input arg, %v", this.GetID().String(), snapshotToDelete.String())
	err := wait.PollImmediate(time.Second, time.Hour, func() (bool, error) {
		this.logger.Debugf("Retrying DeleteSnapshot on IVD Protected Entity, %v, for one hour at the maximum", this.GetID().String())
		vslmTask, err := this.ipetm.vsom.DeleteSnapshot(ctx, NewVimIDFromPEID(this.GetID()), NewVimSnapshotIDFromPESnapshotID(snapshotToDelete))
		if err != nil {
			return false, errors.Wrapf(err, "Failed to create a task for the DeleteSnapshot invocation on IVD Protected Entity, %v, with input arg, %v", this.GetID().String(), snapshotToDelete.String())
		}

		_, err = vslmTask.Wait(ctx, waitTime)
		if err != nil {
			if soap.IsVimFault(err) {
				switch soap.ToVimFault(err).(type) {
				case *vim.InvalidArgument:
					this.logger.WithError(err).Infof("Disk doesn't have given snapshot due to the snapshot stamp was removed in the previous DeleteSnapshot operation which failed with InvalidState fault. And it will be resolved by the next snapshot operation on the same VM. Will NOT retry")
					return true, nil
				case *vim.NotFound:
					this.logger.WithError(err).Infof("There is a temporary catalog mismatch due to a race condition with one another concurrent DeleteSnapshot operation. And it will be resolved by the next consolidateDisks operation on the same VM. Will NOT retry")
					return true, nil
				case *vim.InvalidState:
					this.logger.WithError(err).Error("There is some operation, other than this DeleteSnapshot invocation, on the same VM still being protected by its VM state. Will retry")
					return false, nil
				case *vim.TaskInProgress:
					this.logger.WithError(err).Error("There is some other InProgress operation on the same VM. Will retry")
					return false, nil
				case *vim.FileLocked:
					this.logger.WithError(err).Error("An error occurred while consolidating disks: Failed to lock the file. Will retry")
					return false, nil
				}
			}
			return false, errors.Wrapf(err, "Failed at waiting for the DeleteSnapshot invocation on IVD Protected Entity, %v, with input arg, %v", this.GetID().String(), snapshotToDelete.String())
		}
		return true, nil
	})

	if err != nil {
		return false, err
	}
	this.logger.Infof("DeleteSnapshot completed on IVD Protected Entity, %v, with input arg, %v", this.GetID().String(), snapshotToDelete.String())
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
