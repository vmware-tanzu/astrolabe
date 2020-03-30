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
	"context"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/cns"
	"github.com/vmware/govmomi/session"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"github.com/vmware/govmomi/vslm"
	vslmtypes "github.com/vmware/govmomi/vslm/types"
	"github.com/vmware/gvddk/gDiskLib"
	"io"
	"net/url"
	"time"
)

type IVDProtectedEntityTypeManager struct {
	client    *govmomi.Client
	vsom      *vslm.GlobalObjectManager
	cnsClient *cns.Client
	s3URLBase string
	user      string // These are being kept so we can open VDDK connections, may be able to open a VDDK connection
	password  string // in IVDProtectedEntityTypeManager instead
	logger    logrus.FieldLogger
}

func NewIVDProtectedEntityTypeManagerFromConfig(params map[string]interface{}, s3URLBase string,
	logger logrus.FieldLogger) (*IVDProtectedEntityTypeManager, error) {
	var vcURL url.URL
	vcHostStr, ok := params["vcHost"].(string)
	if !ok {
		return nil, errors.New("Missing vcHost param, cannot initialize IVDProtectedEntityTypeManager")
	}
	vcURL.Scheme = "https"
	vcURL.Host = vcHostStr
	insecure := false
	insecureStr, ok := params["insecureVC"].(string)
	if ok && (insecureStr == "Y" || insecureStr == "y") {
		insecure = true
	}
	vcUser, ok := params["vcUser"].(string)
	if !ok {
		return nil, errors.New("Missing vcUser param, cannot initialize IVDProtectedEntityTypeManager")
	}
	vcPassword, ok := params["vcPassword"].(string)
	if !ok {
		return nil, errors.New("Missing vcPassword param, cannot initialize IVDProtectedEntityTypeManager")
	}
	vcURL.User = url.UserPassword(vcUser, vcPassword)
	vcURL.Path = "/sdk"
	return NewIVDProtectedEntityTypeManagerFromURL(&vcURL, s3URLBase, insecure, logger)
}

func newKeepAliveClient(ctx context.Context, u *url.URL, insecure bool) (*govmomi.Client, error) {
	soapClient := soap.NewClient(u, insecure)
	vimClient, err := vim25.NewClient(ctx, soapClient)
	if err != nil {
		return nil, err
	}

	vimClient.RoundTripper = session.KeepAlive(vimClient.RoundTripper, 10*time.Minute)

	c := &govmomi.Client{
		Client:         vimClient,
		SessionManager: session.NewManager(vimClient),
	}

	// Only login if the URL contains user information.
	if u.User != nil {
		err = c.Login(ctx, u.User)
		if err != nil {
			return nil, err
		}
	}

	return c, nil
}

func NewIVDProtectedEntityTypeManagerFromURL(url *url.URL, s3URLBase string, insecure bool, logger logrus.FieldLogger) (*IVDProtectedEntityTypeManager, error) {
	ctx := context.Background()
	client, err := newKeepAliveClient(ctx, url, insecure)

	if err != nil {
		return nil, err
	}

	vslmClient, err := vslm.NewClient(ctx, client.Client)
	if err != nil {
		return nil, err
	}

	err = client.UseServiceVersion("vsan")
	if err != nil {
		return nil, err
	}
	cnsClient, err := cns.NewClient(ctx, client.Client)
	if err != nil {
		return nil, err
	}

	retVal, err := newIVDProtectedEntityTypeManagerWithClient(client, s3URLBase, vslmClient, cnsClient, logger)
	if err == nil {

		retVal.user = url.User.Username()
		password, hasPassword := url.User.Password()
		if !hasPassword {
			return nil, errors.New("No VC Password specified")
		}
		retVal.password = password
	}
	return retVal, err
}

const vsphereMajor = 6
const vSphereMinor = 7
const disklibLib64 = "/usr/lib/vmware-vix-disklib/lib64"

func newIVDProtectedEntityTypeManagerWithClient(client *govmomi.Client, s3URLBase string, vslmClient *vslm.Client,
	cnsClient *cns.Client, logger logrus.FieldLogger) (*IVDProtectedEntityTypeManager, error) {

	vsom := vslm.NewGlobalObjectManager(vslmClient)

	err := gDiskLib.Init(vsphereMajor, vSphereMinor, disklibLib64)
	if err != nil {
		return nil, errors.Wrap(err, "Could not initialize VDDK")
	}
	retVal := IVDProtectedEntityTypeManager{
		client:    client,
		vsom:      vsom,
		cnsClient: cnsClient,
		s3URLBase: s3URLBase,
		logger:    logger,
	}
	return &retVal, nil
}

func (this *IVDProtectedEntityTypeManager) GetTypeName() string {
	return "ivd"
}

func (this *IVDProtectedEntityTypeManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	retIPE, err := newIVDProtectedEntity(this, id)
	if err != nil {
		return nil, err
	}
	return retIPE, nil
}

func (this *IVDProtectedEntityTypeManager) GetProtectedEntities(ctx context.Context) ([]astrolabe.ProtectedEntityID, error) {
	// Kludge because of PR
	spec := vslmtypes.VslmVsoVStorageObjectQuerySpec{
		QueryField:    "createTime",
		QueryOperator: "greaterThan",
		QueryValue:    []string{"0"},
	}
	res, err := this.vsom.ListObjectsForSpec(ctx, []vslmtypes.VslmVsoVStorageObjectQuerySpec{spec}, 1000)
	if err != nil {
		return nil, err
	}
	retIDs := make([]astrolabe.ProtectedEntityID, len(res.Id))
	for idNum, curVSOID := range res.Id {
		retIDs[idNum] = newProtectedEntityID(curVSOID)
	}
	return retIDs, nil
}

func (this *IVDProtectedEntityTypeManager) Copy(ctx context.Context, sourcePE astrolabe.ProtectedEntity, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	sourcePEInfo, err := sourcePE.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	dataReader, err := sourcePE.GetDataReader(ctx)
	if dataReader != nil {
		defer func() {
			if err := dataReader.Close(); err != nil {
				this.logger.Errorf("The deferred data reader is closed with error, %v", err)
			}
		}()
	}

	if err != nil {
		return nil, err
	}

	metadataReader, err := sourcePE.GetMetadataReader(ctx)
	if err != nil {
		return nil, err
	}
	return this.copyInt(ctx, sourcePEInfo, options, dataReader, metadataReader)
}

func (this *IVDProtectedEntityTypeManager) CopyFromInfo(ctx context.Context, peInfo astrolabe.ProtectedEntityInfo, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	return nil, nil
}

type backingSpec struct {
	createSpec *types.VslmCreateSpecBackingSpec
}

func (this backingSpec) GetVslmCreateSpecBackingSpec() *types.VslmCreateSpecBackingSpec {
	return this.createSpec
}

func (this *IVDProtectedEntityTypeManager) copyInt(ctx context.Context, sourcePEInfo astrolabe.ProtectedEntityInfo,
	options astrolabe.CopyCreateOptions, dataReader io.Reader, metadataReader io.Reader) (astrolabe.ProtectedEntity, error) {
	this.logger.Debug("ivd PETM copyInt called")
	if (sourcePEInfo.GetID().GetPeType() != "ivd") {
		return nil, errors.New("Copy source must be an ivd")
	}
	ourVC := false
	existsInOurVC := false
	for _, checkData := range sourcePEInfo.GetDataTransports() {
		vcenterURL, ok := checkData.GetParam("vcenter")

		if checkData.GetTransportType() == "vadp" && ok && vcenterURL == this.client.URL().Host {
			ourVC = true
			existsInOurVC = true
			break
		}
	}

	if (ourVC) {
		_, err := this.vsom.Retrieve(ctx, NewVimIDFromPEID(sourcePEInfo.GetID()))
		if err != nil {
			if soap.IsSoapFault(err) {
				fault := soap.ToSoapFault(err).Detail.Fault
				if _, ok := fault.(types.NotFound); ok {
					// Doesn't exist in our local system, we can't just clone it
					existsInOurVC = false
				} else {
					return nil, err
				}
			}
		}
	}

	this.logger.WithField("ourVC", ourVC).WithField("existsInOurVC", existsInOurVC).Debug("Ready to restore from snapshot")

	var retPE IVDProtectedEntity
	var createTask *vslm.Task
	var err error

	md, err := readMetadataFromReader(ctx, metadataReader)
	if err != nil {
		return nil, err
	}

	if ourVC && existsInOurVC {
		md, err := FilterLabelsFromMetadataForVslmAPIs(md, this.logger)
		if err != nil {
			return nil, err
		}
		hasSnapshot := sourcePEInfo.GetID().HasSnapshot()
		if (hasSnapshot) {
			createTask, err = this.vsom.CreateDiskFromSnapshot(ctx, NewVimIDFromPEID(sourcePEInfo.GetID()), NewVimSnapshotIDFromPEID(sourcePEInfo.GetID()),
				sourcePEInfo.GetName(), nil, nil, "")
		} else {
			keepAfterDeleteVm := true
			cloneSpec := types.VslmCloneSpec{
				Name:              sourcePEInfo.GetName(),
				KeepAfterDeleteVm: &keepAfterDeleteVm,
				Metadata:          md.ExtendedMetadata,
			}
			createTask, err = this.vsom.Clone(ctx, NewVimIDFromPEID(sourcePEInfo.GetID()), cloneSpec)
		}

		if err != nil {
			this.logger.WithError(err).WithField("HasSnapshot", hasSnapshot).Error("Failed at creating volume from local snapshot/volume")
			return nil, err
		}

		retVal, err := createTask.WaitNonDefault(ctx, time.Hour*24, time.Second*10, true, time.Second*30);
		if err != nil {
			this.logger.WithError(err).Error("Failed at waiting for the task of creating volume")
			return nil, err
		}
		newVSO := retVal.(types.VStorageObject)
		retPE, err = newIVDProtectedEntity(this, newProtectedEntityID(newVSO.Config.Id))

		// if there is any local snasphot, we need to call updateMetadata explicitly
		// since CreateDiskFromSnapshot doesn't accept metadata as a param. The API need to be changed accordingly.
		if (hasSnapshot) {
			updateTask, err := this.vsom.UpdateMetadata(ctx, newVSO.Config.Id, md.ExtendedMetadata, []string{})
			if err != nil {
				this.logger.WithError(err).Error("Failed at calling UpdateMetadata")
				return nil, err
			}
			_, err = updateTask.WaitNonDefault(ctx, time.Hour*24, time.Second*10, true, time.Second*30);
			if err != nil {
				this.logger.WithError(err).Error("Failed at waiting for the UpdateMetadata task")
				return nil, err
			}
		}

	} else {
		// To enable cross-cluster restore, need to filter out the cns specific labels, i.e. prefix: cns, in md
		md = FilterLabelsFromMetadataForCnsAPIs(md, "cns", this.logger)

		this.logger.Debugf("Ready to provision a new volume with the source metadata: %v", md)
		volumeVimID, err := CreateCnsVolumeInCluster(ctx, this.client, this.cnsClient, md, this.logger)
		retPE, err = newIVDProtectedEntity(this, newProtectedEntityID(volumeVimID))
		if err != nil {
			return nil, err
		}
		err = retPE.copy(ctx, dataReader, md)
		if err != nil {
			this.logger.Errorf("Failed to copy data from data source to newly-provisioned IVD protected entity")
			return nil, err
		}
		this.logger.WithField("volumeId", volumeVimID.Id).WithField("volumeName", md.VirtualStorageObject.Config.Name).Debug("Copied snapshot data to newly-provisioned IVD protected entity")
	}

	if err != nil {
		return nil, err
	}
	return retPE, nil
}

func (this *IVDProtectedEntityTypeManager) getDataTransports(id astrolabe.ProtectedEntityID) ([]astrolabe.DataTransport,
	[]astrolabe.DataTransport,
	[]astrolabe.DataTransport, error) {
	vadpParams := make(map[string]string)
	vadpParams["id"] = id.GetID()
	if id.GetSnapshotID().String() != "" {
		vadpParams["snapshotID"] = id.GetSnapshotID().String()
	}
	vadpParams["vcenter"] = this.client.URL().Host

	dataS3URL := this.s3URLBase + "ivd/" + id.String()
	data := []astrolabe.DataTransport{
		astrolabe.NewDataTransport("vadp", vadpParams),
		astrolabe.NewDataTransportForS3URL(dataS3URL),
	}

	mdS3URL := dataS3URL + ".md"

	md := []astrolabe.DataTransport{
		astrolabe.NewDataTransportForS3URL(mdS3URL),
	}

	combinedS3URL := dataS3URL + ".zip"
	combined := []astrolabe.DataTransport{
		astrolabe.NewDataTransportForS3URL(combinedS3URL),
	}

	return data, md, combined, nil
}