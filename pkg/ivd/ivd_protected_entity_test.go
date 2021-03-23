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
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/common/vsphere"
	"github.com/vmware-tanzu/astrolabe/pkg/s3repository"
	"github.com/vmware/govmomi/cns"
	cnstypes "github.com/vmware/govmomi/cns/types"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/pbm"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"math"
	"math/rand"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestProtectedEntityIDFromString(t *testing.T) {

}

const (
	MaxNumOfIVDs = 15
)

func TestSnapshotOpsUnderRaceCondition(t *testing.T) {
	// #0: Setup the environment
	// Prerequisite: export ASTROLABE_VC_URL='https://<VC USER>:<VC USER PASSWORD>@<VC IP>/sdk'
	u, exist := os.LookupEnv("ASTROLABE_VC_URL")
	if !exist {
		t.Skipf("ASTROLABE_VC_URL is not set")
	}

	nIVDs := 5
	nIVDsStr, ok := os.LookupEnv("NUM_OF_IVD")
	if ok {
		nIVDsInt, err := strconv.Atoi(nIVDsStr)
		if err == nil && nIVDsInt > 0 && nIVDsInt <= MaxNumOfIVDs {
			nIVDs = nIVDsInt
		}
	}

	vcUrl, err := soap.ParseURL(u)
	if err != nil {
		t.Skipf("Failed to parse the env variable, ASTROLABE_VC_URL, with err: %v", err)
	}

	ctx := context.Background()
	logger := logrus.New()
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = time.RFC3339Nano
	formatter.FullTimestamp = true
	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)
	s3Config := astrolabe.S3Config{
		URLBase: "VOID_URL",
	}

	params := make(map[string]interface{})
	params[vsphere.HostVcParamKey] = vcUrl.Host
	params[vsphere.PortVcParamKey] = vcUrl.Port()
	params[vsphere.UserVcParamKey] = vcUrl.User.Username()
	password, _ := vcUrl.User.Password()
	params[vsphere.PasswordVcParamKey] = password
	params[vsphere.InsecureFlagVcParamKey] = true
	params[vsphere.ClusterVcParamKey] = ""

	ivdPETM, err := NewIVDProtectedEntityTypeManager(params, s3Config, logger)
	if err != nil {
		t.Skipf("Failed to get a new ivd PETM: %v", err)
	}

	virtualCenter := ivdPETM.vcenter

	// #1: Create a few of IVDs
	datastoreType := types.HostFileSystemVolumeFileSystemTypeVsan
	datastores, err := findAllAccessibleDatastoreByType(ctx, virtualCenter.Client.Client, datastoreType)
	if err != nil || len(datastores) <= 0 {
		t.Skipf("Failed to find any all accessible datastore with type, %v", datastoreType)
	}

	logger.Infof("Step 1: Creating %v IVDs", nIVDs)
	ivdDs := datastores[0]
	var ivdIds []types.ID
	for i := 0; i < nIVDs; i++ {
		createSpec := getCreateSpec(getRandomName("ivd", 5), 10, ivdDs, nil)
		vslmTask, err := ivdPETM.vslmManager.CreateDisk(ctx, createSpec)
		if err != nil {
			t.Skipf("Failed to create task for CreateDisk invocation")
		}

		taskResult, err := vslmTask.Wait(ctx, waitTime)
		if err != nil {
			t.Skipf("Failed at waiting for the CreateDisk invocation")
		}
		vStorageObject := taskResult.(types.VStorageObject)
		ivdIds = append(ivdIds, vStorageObject.Config.Id)
		logger.Debugf("IVD, %v, created", vStorageObject.Config.Id.Id)
	}

	if ivdIds == nil {
		t.Skipf("Failed to create the list of ivds as expected")
	}

	defer func() {
		for i := 0; i < nIVDs; i++ {
			vslmTask, err := ivdPETM.vslmManager.Delete(ctx, ivdIds[i])
			if err != nil {
				t.Skipf("Failed to create task for DeleteDisk invocation with err: %v", err)
			}

			_, err = vslmTask.Wait(ctx, waitTime)
			if err != nil {
				t.Skipf("Failed at waiting for the DeleteDisk invocation with err: %v", err)
			}
			logger.Debugf("IVD, %v, deleted", ivdIds[i].Id)
		}
	}()

	// #2: Create a VM
	logger.Info("Step 2: Creating a VM")
	hosts, err := findAllHosts(ctx, virtualCenter.Client.Client)
	if err != nil || len(hosts) <= 0 {
		t.Skipf("Failed to find all available hosts")
	}
	vmHost := hosts[0]

	pc := property.DefaultCollector(virtualCenter.Client.Client)
	var ivdDsMo mo.Datastore
	err = pc.RetrieveOne(ctx, ivdDs.Reference(), []string{"name"}, &ivdDsMo)
	if err != nil {
		t.Skipf("Failed to get datastore managed object with err: %v", err)
	}

	logger.Debugf("Creating VM on host: %v, and datastore: %v", vmHost.Reference(), ivdDsMo.Name)
	vmName := getRandomName("vm", 5)
	vmMo, err := vmCreate(ctx, virtualCenter.Client.Client, vmHost.Reference(), vmName, ivdDsMo.Name, nil, logger)
	if err != nil {
		t.Skipf("Failed to create a VM with err: %v", err)
	}
	vmRef := vmMo.Reference()
	logger.Debugf("VM, %v(%v), created on host: %v, and datastore: %v", vmRef, vmName, vmHost, ivdDsMo.Name)
	defer func() {
		vimTask, err := vmMo.Destroy(ctx)
		if err != nil {
			t.Skipf("Failed to destroy the VM %v with err: %v", vmName, err)
		}
		err = vimTask.Wait(ctx)
		if err != nil {
			t.Skipf("Failed at waiting for the destroy of VM %v with err: %v", vmName, err)
		}
		logger.Debugf("VM, %v(%v), destroyed", vmRef, vmName)
	}()

	// #3: Attach those IVDs to the VM
	logger.Infof("Step 3: Attaching IVDs to VM %v", vmName)
	for i := 0; i < nIVDs; i++ {
		err = vmAttachDiskWithWait(ctx, virtualCenter.Client.Client, vmRef.Reference(), ivdIds[i], ivdDs.Reference())
		if err != nil {
			t.Skipf("Failed to attach ivd, %v, to, VM, %v with err: %v", ivdIds[i].Id, vmName, err)
		}

		logger.Debugf("IVD, %v, attached to VM, %v", ivdIds[i].Id, vmName)
	}

	defer func() {
		for i := 0; i < nIVDs; i++ {
			err = vmDetachDiskWithWait(ctx, virtualCenter.Client.Client, vmRef.Reference(), ivdIds[i])
			if err != nil {
				t.Skipf("Failed to detach ivd, %v, to, VM, %v with err: %v", ivdIds[i].Id, vmName, err)
			}

			logger.Debugf("IVD, %v, detached from VM, %v", ivdIds[i].Id, vmName)
		}
	}()

	// #4: Mimic the race condition by running the concurrent CreateSnapshot and DeleteSnapshot operations
	logger.Info("Step 4: Testing the API behavior under concurrent snapshot invocations")
	errChannels := make([]chan error, nIVDs)
	var wg sync.WaitGroup
	var mutex sync.Mutex
	for i := 0; i < nIVDs; i++ {
		wg.Add(1)
		go worker(&wg, &mutex, logger, params, i, ivdIds[i], ivdDs, errChannels)
	}
	wg.Wait()

	defer func() {
		logger.Debugf("Always clean up snapshots created in the test")
		for i := 0; i < nIVDs; i++ {
			logger.Debugf("Cleaning up snapshots for IVD %v", ivdIds[i].Id)
			snapshotInfos, err := ivdPETM.vslmManager.RetrieveSnapshotInfo(ctx, ivdIds[i])
			if err != nil {
				t.Fatalf("Failed at retrieving snapshot info from IVD %v with err: %v", ivdIds[i].Id, err)
			}

			if len(snapshotInfos) == 0 {
				logger.Debugf("No unexpected snasphot left behind for IVD %v", ivdIds[i].Id)
				continue
			}

			for _, snapshotInfo := range snapshotInfos {
				logger.Debugf("Cleaning up snapshot %v created for IVD %v but failed to be deleted", snapshotInfo.Id.Id, ivdIds[i].Id)
				ivdPE, err := ivdPETM.GetProtectedEntity(ctx, newProtectedEntityID(ivdIds[i]))
				if err != nil {
					t.Fatalf("[Cleanup] Failed to get IVD protected entity at the cleanup phase with err: %v", err)
				}
				peSnapID := astrolabe.NewProtectedEntitySnapshotID(snapshotInfo.Id.Id)
				_, err = ivdPE.DeleteSnapshot(ctx, peSnapID, make(map[string]map[string]interface{}))
				if err != nil {
					t.Fatalf("[Cleanup] Failed to DeleteSnapshot, %v, on IVD protected entity, %v with err: %v", peSnapID.GetID(), ivdPE.GetID().GetID(), err)
				}
			}
		}
	}()

	// Error Handling
	var result bool
	result = true
	for i := 0; i < nIVDs; i++ {
		err := <-errChannels[i]
		if err != nil {
			result = false
			t.Errorf("Worker %v on IVD %v failed with err: %v", i, ivdIds[i].Id, err)
		}
	}

	if !result {
		t.Fatal("Test Failed")
	}

}

func worker(wg *sync.WaitGroup, mutex *sync.Mutex, logger logrus.FieldLogger, params map[string]interface{}, id int, diskId types.ID, datastore types.ManagedObjectReference, errChans []chan error) {
	log := logger.WithFields(logrus.Fields{
		"WorkerID": id,
		"IvdID":    diskId.Id,
	})
	var err error
	log.Debugf("Worker starting")
	defer func() {
		log.Debugf("Worker completed with err: %v", err)
	}()

	errChans[id] = make(chan error)
	defer func() {
		errChans[id] <- err
		close(errChans[id])
	}()

	defer wg.Done()

	ctx := context.Background()

	s3Config := astrolabe.S3Config{
		URLBase: "VOID_URL",
	}
	ivdPETM, err := NewIVDProtectedEntityTypeManager(params, s3Config, logger)
	if err != nil {
		log.Error("Failed to get a new ivd PETM")
		return
	}

	ivdPE, err := ivdPETM.GetProtectedEntity(ctx, newProtectedEntityID(diskId))
	if err != nil {
		log.Error("Failed to get IVD protected entity")
		return
	}

	log.Debugf("Creating a snapshot on IVD protected entity")
	peSnapID, err := createSnapshotLocked(mutex, ctx, ivdPE, log)
	if err != nil {
		log.Error("Failed to snapshot the IVD protected entity")
		return
	}

	log.Debugf("Retrieving the newly created snapshot, %v, on IVD protected entity, %v", peSnapID.GetID(), ivdPE.GetID().GetID())
	_, err = ivdPETM.vslmManager.RetrieveSnapshotDetails(ctx, diskId, NewIDFromString(peSnapID.String()))
	if err != nil {
		if soap.IsSoapFault(err) {
			soapFault := soap.ToSoapFault(err)
			soapType := reflect.TypeOf(soapFault)
			log.WithError(err).Errorf("soap fault type: %v, err: %v", soapType, soapFault)
			faultMsg := soap.ToSoapFault(err).String
			if strings.Contains(faultMsg, "A specified parameter was not correct: snapshotId") {
				log.WithError(err).Error("Unexpected InvalidArgument soap fault due to race condition")
				return
			}
			log.WithError(err).Error("Unexpected soap fault")
		} else {
			log.WithError(err).Error("Unexpected other fault")
		}

		return
	}

	log.Debugf("Deleting the newly created snapshot, %v, on IVD protected entity, %v", peSnapID.GetID(), ivdPE.GetID().GetID())
	_, err = ivdPE.DeleteSnapshot(ctx, peSnapID, make(map[string]map[string]interface{}))
	if err != nil {
		log.WithError(err).Errorf("Failed to DeleteSnapshot, %v, on IVD protected entity, %v", peSnapID.GetID(), ivdPE.GetID().GetID())
	}
}

func createSnapshotLocked(mutex *sync.Mutex, ctx context.Context, ivdPE astrolabe.ProtectedEntity, log logrus.FieldLogger) (astrolabe.ProtectedEntitySnapshotID, error) {
	log.Debugf("Acquiring the lock on CreateSnapshot")
	mutex.Lock()
	log.Debugf("Acquired the lock on CreateSnapshot")
	defer func() {
		mutex.Unlock()
		log.Debugf("Released the lock on CreateSnapshot")
	}()
	peSnapID, err := ivdPE.Snapshot(ctx, nil)
	if err != nil {
		log.Error("Failed to snapshot the IVD protected entity")
		return astrolabe.ProtectedEntitySnapshotID{}, err
	}
	return peSnapID, nil
}

func findAllHosts(ctx context.Context, client *vim25.Client) ([]types.ManagedObjectReference, error) {
	finder := find.NewFinder(client)

	hosts, err := finder.HostSystemList(ctx, "*")
	if err != nil {
		return nil, err
	}

	var hostList []types.ManagedObjectReference
	for _, host := range hosts {
		hostList = append(hostList, host.Reference())
	}

	return hostList, nil
}

func findAllAccessibleDatastoreByType(ctx context.Context, client *vim25.Client, datastoreType types.HostFileSystemVolumeFileSystemType) ([]types.ManagedObjectReference, error) {
	finder := find.NewFinder(client)

	hosts, err := findAllHosts(ctx, client)
	if err != nil {
		return nil, err
	}
	nHosts := len(hosts)

	dss, err := finder.DatastoreList(ctx, "*")
	if err != nil {
		return nil, err
	}

	var dsList []types.ManagedObjectReference
	for _, ds := range dss {
		attachedHosts, err := ds.AttachedHosts(ctx)
		if err != nil {
			fmt.Printf("Failed to get all the attached hosts of datastore %v\n", ds.Name())
			continue
		}
		if nHosts != len(attachedHosts) {
			continue
		}

		dsType, err := ds.Type(ctx)
		if err != nil {
			fmt.Printf("Failed to get type of datastore %v\n", ds.Name())
			continue
		}
		if dsType == datastoreType {
			dsList = append(dsList, ds.Reference())
			break
		}
	}

	return dsList, nil
}

func getCreateSpec(name string, capacity int64, datastore types.ManagedObjectReference, profile []types.BaseVirtualMachineProfileSpec) types.VslmCreateSpec {
	keepAfterDeleteVm := true
	return types.VslmCreateSpec{
		Name:              name,
		KeepAfterDeleteVm: &keepAfterDeleteVm,
		BackingSpec: &types.VslmCreateSpecDiskFileBackingSpec{
			VslmCreateSpecBackingSpec: types.VslmCreateSpecBackingSpec{
				Datastore: datastore,
			},
		},
		CapacityInMB: capacity,
		Profile:      profile,
	}
}

func getRandomName(prefix string, nDigits int) string {
	rand.Seed(time.Now().UnixNano())
	num := rand.Int63n(int64(math.Pow10(nDigits)))
	numstr := strconv.FormatInt(num, 10)
	return fmt.Sprintf("%s-%s", prefix, numstr)
}

func vmAttachDisk(ctx context.Context, client *vim25.Client, vm types.ManagedObjectReference, diskId types.ID, datastore types.ManagedObjectReference) (*object.Task, error) {
	req := types.AttachDisk_Task{
		This:       vm.Reference(),
		DiskId:     diskId,
		Datastore:  datastore.Reference(),
		UnitNumber: nil,
	}

	res, err := methods.AttachDisk_Task(ctx, client, &req)
	if err != nil {
		return nil, err
	}

	return object.NewTask(client, res.Returnval), nil
}

func vmAttachDiskWithWait(ctx context.Context, client *vim25.Client, vm types.ManagedObjectReference, diskId types.ID, datastore types.ManagedObjectReference) error {
	vimTask, err := vmAttachDisk(ctx, client, vm, diskId, datastore)
	if err != nil {
		return err
	}

	err = vimTask.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func vmDetachDisk(ctx context.Context, client *vim25.Client, vm types.ManagedObjectReference, diskId types.ID) (*object.Task, error) {
	req := types.DetachDisk_Task{
		This:   vm.Reference(),
		DiskId: diskId,
	}

	res, err := methods.DetachDisk_Task(ctx, client, &req)
	if err != nil {
		return nil, err
	}

	return object.NewTask(client, res.Returnval), nil
}

func vmDetachDiskWithWait(ctx context.Context, client *vim25.Client, vm types.ManagedObjectReference, diskId types.ID) error {
	vimTask, err := vmDetachDisk(ctx, client, vm, diskId)
	if err != nil {
		return err
	}

	err = vimTask.Wait(ctx)
	if err != nil {
		return err
	}

	return nil
}

func vmCreate(ctx context.Context, client *vim25.Client, vmHost types.ManagedObjectReference, vmName string, dsName string, vmProfile []types.BaseVirtualMachineProfileSpec, logger logrus.FieldLogger) (*object.VirtualMachine, error) {
	finder := find.NewFinder(client)
	virtualMachineConfigSpec := types.VirtualMachineConfigSpec{
		Name: vmName,
		Files: &types.VirtualMachineFileInfo{
			VmPathName: "[" + dsName + "]",
		},
		Annotation: "Quick Dummy",
		GuestId:    "otherLinux64Guest",
		NumCPUs:    1,
		MemoryMB:   128,
		DeviceChange: []types.BaseVirtualDeviceConfigSpec{
			&types.VirtualDeviceConfigSpec{
				Operation: types.VirtualDeviceConfigSpecOperationAdd,
				Device: &types.ParaVirtualSCSIController{
					VirtualSCSIController: types.VirtualSCSIController{
						SharedBus: types.VirtualSCSISharingNoSharing,
						VirtualController: types.VirtualController{
							BusNumber: 0,
							VirtualDevice: types.VirtualDevice{
								Key: 1000,
							},
						},
					},
				},
			},
		},
		VmProfile: vmProfile,
	}
	defaultFolder, err := finder.DefaultFolder(ctx)
	defaultResourcePool, err := finder.DefaultResourcePool(ctx)
	vmHostSystem := object.NewHostSystem(client, vmHost)
	task, err := defaultFolder.CreateVM(ctx, virtualMachineConfigSpec, defaultResourcePool, vmHostSystem)
	if err != nil {
		logger.Errorf("Failed to create VM. Error: %v", err)
		return nil, err
	}

	vmTaskInfo, err := task.WaitForResult(ctx, nil)
	if err != nil {
		logger.Errorf("Error occurred while waiting for create VM task result. Error: %v", err)
		return nil, err
	}

	vmRef := vmTaskInfo.Result.(object.Reference)
	nodeVM := object.NewVirtualMachine(client, vmRef.Reference())
	return nodeVM, nil
}

func TestBackupIVDs(t *testing.T) {
	// #0: Setup the environment
	// Prerequisite: export ASTROLABE_VC_URL='https://<VC USER>:<VC USER PASSWORD>@<VC IP>/sdk'
	u, exist := os.LookupEnv("ASTROLABE_VC_URL")
	if !exist {
		t.Skipf("ASTROLABE_VC_URL is not set")
	}

	enableDebugLog := false
	enableDebugLogStr, ok := os.LookupEnv("ENABLE_DEBUG_LOG")
	if ok {
		if res, _ := strconv.ParseBool(enableDebugLogStr); res {
			enableDebugLog = true
		}
	}

	ctx := context.Background()
	logger := logrus.New()
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = time.RFC3339Nano
	formatter.FullTimestamp = true
	logger.SetFormatter(formatter)
	if enableDebugLog {
		logger.SetLevel(logrus.DebugLevel)
	}

	// BEGIN: configuration options
	// Use two IVDs in this test case if not explicitly specified
	nIVDs := 2
	nIVDsStr, ok := os.LookupEnv("NUM_OF_IVD")
	if ok {
		nIVDsInt, err := strconv.Atoi(nIVDsStr)
		if err == nil && nIVDsInt > 0 && nIVDsInt <= MaxNumOfIVDs {
			nIVDs = nIVDsInt
		}
	}

	// Use Encryption
	useEncryption := false
	useEncryptionStr, ok := os.LookupEnv("USE_ENCRYPTION")
	if ok {
		if res, _ := strconv.ParseBool(useEncryptionStr); res {
			useEncryption = true
		}
	}

	// Enable Remote Copy of IVD Snapshot
	enableRemoteCopy := false
	enableRemoteCopyStr, ok := os.LookupEnv("ENABLE_REMOTE_COPY")
	if ok {
		if res, _ := strconv.ParseBool(enableRemoteCopyStr); res {
			enableRemoteCopy = true
		}
	}
	logger.Debugf("Configuration options: numOfIVDs = %v, useEncryption=%v, enableRemoteCopy=%v", nIVDs, useEncryption, enableRemoteCopy)
	// END: configuration options

	vcUrl, err := soap.ParseURL(u)
	if err != nil {
		t.Skipf("Failed to parse the env variable, ASTROLABE_VC_URL, with err: %v", err)
	}

	s3Config := astrolabe.S3Config{
		URLBase: "VOID_URL",
	}
	params := make(map[string]interface{})
	params[vsphere.HostVcParamKey] = vcUrl.Host
	if vcUrl.Port() == "" {
		params[vsphere.PortVcParamKey] = "443"
	} else {
		params[vsphere.PortVcParamKey] = vcUrl.Port()
	}
	params[vsphere.UserVcParamKey] = vcUrl.User.Username()
	password, _ := vcUrl.User.Password()
	params[vsphere.PasswordVcParamKey] = password
	params[vsphere.InsecureFlagVcParamKey] = "true"
	params[vsphere.ClusterVcParamKey] = ""

	logger.Infof("params: %v", params)

	ivdPETM, err := NewIVDProtectedEntityTypeManager(params, s3Config, logger)
	if err != nil {
		t.Skipf("Failed to get a new ivd PETM: %v", err)
	}
	virtualCenter := ivdPETM.vcenter
	datastoreType := types.HostFileSystemVolumeFileSystemTypeVsan
	datastores, err := findAllAccessibleDatastoreByType(ctx, virtualCenter.Client.Client, datastoreType)
	if err != nil || len(datastores) <= 0 {
		t.Skipf("Failed to find any all accessible datastore with type, %v", datastoreType)
	}
	vmDs := datastores[0]
	ivdDs := datastores[len(datastores)-1]

	// check if use encrpytion on IVDs
	var encryptionProfileId string
	if useEncryption {
		encryptionProfileId, err = getEncryptionProfileId(ctx, virtualCenter.Client.Client)
		if err != nil {
			t.Skipf("Failed to get encryption profile ID: %v", err)
		}
	}

	// #1: Create an VM
	logger.Info("Step 1: Creating a VM")
	hosts, err := findAllHosts(ctx, virtualCenter.Client.Client)
	if err != nil || len(hosts) <= 0 {
		t.Skipf("Failed to find all available hosts")
	}
	vmHost := hosts[0]

	pc := property.DefaultCollector(virtualCenter.Client.Client)
	var vmDsMo mo.Datastore
	err = pc.RetrieveOne(ctx, vmDs.Reference(), []string{"name"}, &vmDsMo)
	if err != nil {
		t.Skipf("Failed to get datastore managed object with err: %v", err)
	}

	logger.Debugf("Creating VM on host: %v, and datastore: %v", vmHost.Reference(), vmDsMo.Name)
	vmName := getRandomName("vm", 5)
	vmProfile := getProfileSpecs(encryptionProfileId)
	vmMo, err := vmCreate(ctx, virtualCenter.Client.Client, vmHost.Reference(), vmName, vmDsMo.Name, vmProfile, logger)
	if err != nil {
		t.Skipf("Failed to create a VM with err: %v", err)
	}
	vmRef := vmMo.Reference()
	logger.Debugf("VM, %v(%v), created on host: %v, and datastore: %v", vmRef, vmName, vmHost, vmDsMo.Name)
	defer func() {
		vimTask, err := vmMo.Destroy(ctx)
		if err != nil {
			t.Skipf("Failed to destroy the VM %v with err: %v", vmName, err)
		}
		err = vimTask.Wait(ctx)
		if err != nil {
			t.Skipf("Failed at waiting for the destroy of VM %v with err: %v", vmName, err)
		}
		logger.Debugf("VM, %v(%v), destroyed", vmRef, vmName)
	}()

	// #2: Poweron the VM
	logger.Info("Step 2: Powering on a VM")
	vimTask, err := vmMo.PowerOn(ctx)
	if err != nil {
		t.Skipf("Failed to create a VM PowerOn task: %v", err)
	}
	err = vimTask.Wait(ctx)
	if err != nil {
		t.Skipf("Failed at waiting for the PowerOn of VM %v with err: %v", vmName, err)
	}
	logger.Debugf("VM, %v(%v), powered on", vmRef, vmName)
	defer func() {
		vimTask, err := vmMo.PowerOff(ctx)
		if err != nil {
			t.Skipf("Failed to create a VM PowerOff task: %v", err)
		}
		err = vimTask.Wait(ctx)
		if err != nil {
			t.Skipf("Failed at waiting for the PowerOff of VM %v with err: %v", vmName, err)
		}
		logger.Debugf("VM, %v(%v), powered off", vmRef, vmName)
	}()

	// #3: Create CNS Block Volumes
	logger.Infof("Step 3: Creating %v IVDs", nIVDs)
	ivdProfile := vmProfile

	var ivdIds []types.ID
	for i := 0; i < nIVDs; i++ {
		volumeIdStr, err := CnsCreateVolume(ctx, ivdPETM.vcenter.CnsClient, logger, getRandomName("ivd", 5), 50, ivdDs, ivdProfile)
		if err != nil {
			t.Skip(err.Error())
		}
		ivdIds = append(ivdIds, types.ID{Id:volumeIdStr})
	}

	if ivdIds == nil {
		t.Skipf("Failed to create the list of ivds as expected")
	}

	defer func() {
		for i := 0; i < nIVDs; i++ {
			err := CnsDeleteVolume(ctx, ivdPETM.vcenter.CnsClient, logger, ivdIds[i].Id)
			if err != nil {
				t.Skip(err.Error())
			}
		}
	}()

	// #4: Attach it to VM
	logger.Infof("Step 4: Attaching IVDs to VM %v", vmName)
	for i := 0; i < nIVDs; i++ {
		err = CnsAttachVolume(ctx, ivdPETM.vcenter.CnsClient, logger, ivdIds[i].Id, vmRef.Reference())
		if err != nil {
			t.Skipf("Failed to attach ivd, %v, to, VM, %v with err: %v", ivdIds[i].Id, vmName, err)
		}

		logger.Debugf("IVD, %v, attached to VM, %v", ivdIds[i].Id, vmName)
	}

	defer func() {
		for i := 0; i < nIVDs; i++ {
			err = vmDetachDiskWithWait(ctx, virtualCenter.Client.Client, vmRef.Reference(), ivdIds[i])
			if err != nil {
				t.Skipf("Failed to detach ivd, %v, to, VM, %v with err: %v", ivdIds[i].Id, vmName, err)
			}

			logger.Debugf("IVD, %v, detached from VM, %v", ivdIds[i].Id, vmName)
		}
	}()

	// #5: Backup the IVD
	logger.Infof("Step 5: Backing up IVDs")
	// #5.1: Create an IVD snapshot
	logger.Debugf("Creating a snapshot on each IVD")
	//var snapPEIDs []astrolabe.ProtectedEntityID
	snapPEIDtoIvdPEMap := make(map[astrolabe.ProtectedEntityID]astrolabe.ProtectedEntity)
	for _, ivdId := range ivdIds {
		ivdPE, err := ivdPETM.GetProtectedEntity(ctx, newProtectedEntityID(ivdId))
		if err != nil {
			t.Skipf("Failed to get IVD protected entity for the IVD, %v", ivdId)
		}

		snapID, err := ivdPE.Snapshot(ctx, nil)
		if err != nil {
			t.Fatalf("Failed to snapshot the IVD protected entity, %v: %v", ivdId, err)
		}
		snapPEID := astrolabe.NewProtectedEntityIDWithSnapshotID("ivd", ivdId.Id, snapID)
		snapPEIDtoIvdPEMap[snapPEID] = ivdPE
	}

	defer func() {
		// Delete the local IVD snapshot
		for snapPEID, ivdPE := range snapPEIDtoIvdPEMap {
			_, err := ivdPE.DeleteSnapshot(ctx, snapPEID.GetSnapshotID(), make(map[string]map[string]interface{}))
			if err != nil {
				t.Fatalf("Failed to delete local IVD snapshot, %v: %v", snapPEID.GetSnapshotID(), err)
			}
		}
	}()

	if !enableRemoteCopy {
		logger.Infof("Skipping the remote copy of IVD snapshot based on configuration options in test")
		return
	}

	// #5.2: Copy the IVD snapshot to specified object store
	logger.Debugf("Copying the IVD snapshots to object store")
	s3PETM, err := setupPETM("ivd", logger)
	if err != nil {
		t.Skipf("Failed to setup s3 PETM for the object store")
	}

	snapPEIDtos3PEMap := make(map[astrolabe.ProtectedEntityID]astrolabe.ProtectedEntity)
	for snapPEID, _ := range snapPEIDtoIvdPEMap {
		snapPE, err := ivdPETM.GetProtectedEntity(ctx, snapPEID)
		if err != nil {
			t.Fatalf("Failed to get snapshot protected entity for the IVD snapshot, %v", snapPEID.String())
		}
		s3PE, err := s3PETM.Copy(ctx, snapPE, make(map[string]map[string]interface{}), astrolabe.AllocateNewObject)
		if err != nil {
			t.Fatalf("Failed to copy snapshot PE, %v, to S3 object store: %v", snapPEID.String(), err)
		}
		snapPEIDtos3PEMap[snapPEID] = s3PE
	}

	defer func() {
		for snapPEID, _ := range snapPEIDtoIvdPEMap {
			s3PE := snapPEIDtos3PEMap[snapPEID]
			_, err := s3PE.DeleteSnapshot(ctx, snapPEID.GetSnapshotID(), make(map[string]map[string]interface{}))
			if err != nil {
				logger.Errorf("Failed to delete snapshot, %v, on object store: %v", snapPEID.GetSnapshotID().String(), err)
			}
		}
	}()
}

func CnsCreateVolume(ctx context.Context, cnsClient *cns.Client, logger logrus.FieldLogger, name string, capacity int64, datastore types.ManagedObjectReference, profile []types.BaseVirtualMachineProfileSpec) (string, error) {
	var dsList []types.ManagedObjectReference
	dsList = append(dsList, datastore.Reference())

	var containerClusterArray []cnstypes.CnsContainerCluster
	containerCluster := cnstypes.CnsContainerCluster{
		ClusterType:         string(cnstypes.CnsClusterTypeKubernetes),
		ClusterId:           "demo-cluster-id",
		VSphereUser:         "Administrator@vsphere.local",
		ClusterFlavor:       string(cnstypes.CnsClusterFlavorVanilla),
	}
	containerClusterArray = append(containerClusterArray, containerCluster)

	// Test CreateVolume API
	var cnsVolumeCreateSpecList []cnstypes.CnsVolumeCreateSpec
	cnsVolumeCreateSpec := cnstypes.CnsVolumeCreateSpec{
		Name:       name,
		VolumeType: string(cnstypes.CnsVolumeTypeBlock),
		Datastores: dsList,
		Metadata: cnstypes.CnsVolumeMetadata{
			ContainerCluster: containerCluster,
		},
		BackingObjectDetails: &cnstypes.CnsBlockBackingDetails{
			CnsBackingObjectDetails: cnstypes.CnsBackingObjectDetails{
				CapacityInMb: capacity,
			},
		},
		Profile: profile,
	}
	cnsVolumeCreateSpecList = append(cnsVolumeCreateSpecList, cnsVolumeCreateSpec)

	createTask, err := cnsClient.CreateVolume(ctx, cnsVolumeCreateSpecList)
	if err != nil {
		logger.Errorf("Failed to create volume. Error: %+v \n", err)
		return "", err
	}
	createTaskInfo, err := cns.GetTaskInfo(ctx, createTask)
	if err != nil {
		logger.Errorf("Failed to create volume. Error: %+v \n", err)
		return "", err
	}
	createTaskResult, err := cns.GetTaskResult(ctx, createTaskInfo)
	if err != nil {
		logger.Errorf("Failed to create volume. Error: %+v \n", err)
		return "", err
	}
	if createTaskResult == nil {
		errorMsg := "Empty create task results"
		return "", errors.New(errorMsg)
	}
	createVolumeOperationRes := createTaskResult.GetCnsVolumeOperationResult()
	if createVolumeOperationRes.Fault != nil {
		errorMsg := fmt.Sprintf("Failed to create volume: fault=%+v", createVolumeOperationRes.Fault)
		logger.Error(errorMsg)
		return "", errors.New(errorMsg)
	}
	volumeID := createVolumeOperationRes.VolumeId.Id
	logger.Debugf("CNS Block Volume created sucessfully. volumeId: %s", volumeID)

	return volumeID, nil
}

func CnsDeleteVolume(ctx context.Context, cnsClient *cns.Client, logger logrus.FieldLogger, volumeID string) error {
	var deleteVolumeFromSnapshotVolumeIDList []cnstypes.CnsVolumeId
	deleteVolumeFromSnapshotVolumeIDList = append(deleteVolumeFromSnapshotVolumeIDList, cnstypes.CnsVolumeId{Id: volumeID})
	logger.Debugf("Deleting volume: %+v", deleteVolumeFromSnapshotVolumeIDList)
	deleteVolumeFromSnapshotTask, err := cnsClient.DeleteVolume(ctx, deleteVolumeFromSnapshotVolumeIDList, true)
	if err != nil {
		logger.Errorf("Failed to delete volume. Error: %+v \n", err)
		return err
	}
	deleteVolumeFromSnapshotTaskInfo, err := cns.GetTaskInfo(ctx, deleteVolumeFromSnapshotTask)
	if err != nil {
		logger.Errorf("Failed to delete volume. Error: %+v \n", err)
		return err
	}
	deleteVolumeFromSnapshotTaskResult, err := cns.GetTaskResult(ctx, deleteVolumeFromSnapshotTaskInfo)
	if err != nil {
		logger.Errorf("Failed to delete volume. Error: %+v \n", err)
		return err
	}
	if deleteVolumeFromSnapshotTaskResult == nil {
		errorMsg := "Empty delete task results"
		return errors.New(errorMsg)
	}
	deleteVolumeFromSnapshotOperationRes := deleteVolumeFromSnapshotTaskResult.GetCnsVolumeOperationResult()
	if deleteVolumeFromSnapshotOperationRes.Fault != nil {
		errorMsg := fmt.Sprintf("Failed to delete volume: fault=%+v", deleteVolumeFromSnapshotOperationRes.Fault)
		return errors.New(errorMsg)
	}
	logger.Debugf("CNS Block Volume: %q deleted sucessfully", volumeID)
	return nil
}

func CnsAttachVolume(ctx context.Context, cnsClient *cns.Client, logger logrus.FieldLogger, volumeID string, vmRef types.ManagedObjectReference) error {
	var cnsVolumeAttachSpecList []cnstypes.CnsVolumeAttachDetachSpec
	cnsVolumeAttachSpec := cnstypes.CnsVolumeAttachDetachSpec{
		VolumeId: cnstypes.CnsVolumeId{
			Id: volumeID,
		},
		Vm: vmRef,
	}
	cnsVolumeAttachSpecList = append(cnsVolumeAttachSpecList, cnsVolumeAttachSpec)
	logger.Debugf("Attaching volume using the spec: %+v", cnsVolumeAttachSpec)
	attachTask, err := cnsClient.AttachVolume(ctx, cnsVolumeAttachSpecList)
	if err != nil {
		logger.Errorf("Failed to attach volume. Error: %+v \n", err)
		return err
	}
	attachTaskInfo, err := cns.GetTaskInfo(ctx, attachTask)
	if err != nil {
		logger.Errorf("Failed to attach volume. Error: %+v \n", err)
		return err
	}
	attachTaskResult, err := cns.GetTaskResult(ctx, attachTaskInfo)
	if err != nil {
		logger.Errorf("Failed to attach volume. Error: %+v \n", err)
		return err
	}
	if attachTaskResult == nil {
		errorMsg := "Empty attach task results"
		return errors.New(errorMsg)
	}
	attachVolumeOperationRes := attachTaskResult.GetCnsVolumeOperationResult()
	if attachVolumeOperationRes.Fault != nil {
		errorMsg := fmt.Sprintf("Failed to attach volume: fault=%+v", attachVolumeOperationRes.Fault)
		return errors.New(errorMsg)
	}
	diskUUID := interface{}(attachTaskResult).(*cnstypes.CnsVolumeAttachResult).DiskUUID
	logger.Debugf("CNS Block Volume attached sucessfully. Disk UUID: %s", diskUUID)

	return nil
}

func getProfileSpecs(profileId string) []types.BaseVirtualMachineProfileSpec {
	var profileSpecs []types.BaseVirtualMachineProfileSpec
	if profileId == "" {
		profileSpecs = append(profileSpecs, &types.VirtualMachineDefaultProfileSpec{})
	} else {
		profileSpecs = append(profileSpecs, &types.VirtualMachineDefinedProfileSpec{
			VirtualMachineProfileSpec: types.VirtualMachineProfileSpec{},
			ProfileId:                 profileId,
		})
	}
	return profileSpecs
}

func getEncryptionProfileId(ctx context.Context, client *vim25.Client) (string, error) {
	pbmClient, err := pbm.NewClient(ctx, client)
	if err != nil {
		return "", err
	}

	encryptionProfileName := "VM Encryption Policy"
	return pbmClient.ProfileIDByName(ctx, encryptionProfileName)
}

func setupPETM(typeName string, logger logrus.FieldLogger) (*s3repository.ProtectedEntityTypeManager, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1")},
	)
	if err != nil {
		return nil, err
	}
	s3petm, err := s3repository.NewS3RepositoryProtectedEntityTypeManager(typeName, *sess, "velero-plugin-s3-repo",
		"plugins/vsphere-volumes-repo/", logger)
	if err != nil {
		return nil, err
	}
	return s3petm, err
}
