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
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/soap"
	"github.com/vmware/govmomi/vim25/types"
	"math"
	"math/rand"
	"net/url"
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

func TestSnapshotOpsUnderRaceCondition(t *testing.T) {
	// #0: Setup the environment
	// Prerequisite: export ASTROLABE_VC_URL='https://<VC USER>:<VC USER PASSWORD>@<VC IP>/sdk'
	u, exist := os.LookupEnv("ASTROLABE_VC_URL")
	if !exist {
		t.Skipf("ASTROLABE_VC_URL is not set")
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
	ivdPETM, err := NewIVDProtectedEntityTypeManagerFromURL(vcUrl, "/ivd", true, logger)
	if err != nil {
		t.Skipf("Failed to get a new ivd PETM: %v", err)
	}

	// #1: Create a few of IVDs
	nIVDs := 5
	datastoreType := types.HostFileSystemVolumeFileSystemTypeVsan
	datastores, err := findAllAccessibleDatastoreByType(ctx, ivdPETM.client.Client, datastoreType)
	if err != nil || len(datastores) <= 0 {
		t.Skipf("Failed to find any all accessible datastore with type, %v", datastoreType)
	}

	logger.Infof("Step 1: Creating %v IVDs", nIVDs)
	ivdDs := datastores[0]
	var ivdIds []types.ID
	for i := 0; i < nIVDs; i++ {
		createSpec := getCreateSpec(getRandomName("ivd", 5), 10, ivdDs)
		vslmTask, err := ivdPETM.vsom.CreateDisk(ctx, createSpec)
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
			vslmTask, err := ivdPETM.vsom.Delete(ctx, ivdIds[i])
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
	hosts, err := findAllHosts(ctx, ivdPETM.client.Client)
	if err != nil || len(hosts) <= 0 {
		t.Skipf("Failed to find all available hosts")
	}
	vmHost := hosts[0]

	pc := property.DefaultCollector(ivdPETM.client.Client)
	var ivdDsMo mo.Datastore
	err = pc.RetrieveOne(ctx, ivdDs.Reference(), []string{"name"}, &ivdDsMo)
	if err != nil {
		t.Skipf("Failed to get datastore managed object with err: %v", err)
	}

	logger.Debugf("Creating VM on host: %v, and datastore: %v", vmHost.Reference(), ivdDsMo.Name)
	vmName := getRandomName("vm", 5)
	vmMo, err := vmCreate(ctx, ivdPETM.client.Client, vmHost.Reference(), vmName, ivdDsMo.Name, logger)
	if err != nil {
		t.Skipf("Failed to create a VM with err: %v", err)
	}
	vmRef := vmMo.Reference()
	logger.Debugf("VM, %v(%v), created on host: %v, and datastore: %v", vmRef, vmName, vmHost, ivdDsMo.Name)
	defer func () {
		vimTask, err := vmMo.Destroy(ctx)
		if err != nil {
			t.Skipf("Failed to destroy the VM %v with err: %v", vmName, err)
		}
		err = vimTask.Wait(ctx)
		if err != nil {
			t.Skipf("Failed at waiting for the destroy of VM %v with err: %v", vmName, err)
		}
		logger.Debugf("VM, %v(%v), destroyed", vmRef, vmName)
	} ()

	// #3: Attach those IVDs to the VM
	logger.Infof("Step 3: Attaching IVDs to VM %v", vmName)
	for i := 0; i < nIVDs; i++ {
		err = vmAttachDiskWithWait(ctx, ivdPETM.client.Client, vmRef.Reference(), ivdIds[i], ivdDs.Reference())
		if err != nil {
			t.Skipf("Failed to attach ivd, %v, to, VM, %v with err: %v", ivdIds[i].Id, vmName, err)
		}

		logger.Debugf("IVD, %v, attached to VM, %v", ivdIds[i].Id, vmName)
	}

	defer func() {
		for i := 0; i < nIVDs; i++ {
			err = vmDetachDiskWithWait(ctx, ivdPETM.client.Client, vmRef.Reference(), ivdIds[i])
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
		go worker(&wg, &mutex, logger, vcUrl, i, ivdIds[i], ivdDs, errChannels)
	}
	wg.Wait()

	defer func() {
		logger.Debugf("Always clean up snapshots created in the test")
		for i := 0; i < nIVDs; i++ {
			logger.Debugf("Cleaning up snapshots for IVD %v", ivdIds[i].Id)
			snapshotInfos, err := ivdPETM.vsom.RetrieveSnapshotInfo(ctx, ivdIds[i])
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
				_, err = ivdPE.DeleteSnapshot(ctx, peSnapID)
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

func worker(wg *sync.WaitGroup, mutex *sync.Mutex, logger logrus.FieldLogger, vcUrl *url.URL, id int, diskId types.ID, datastore types.ManagedObjectReference, errChans []chan error) {
	log := logger.WithFields(logrus.Fields{
		"WorkerID": id,
		"IvdID": diskId.Id,
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

	ivdPETM, err := NewIVDProtectedEntityTypeManagerFromURL(vcUrl, "/ivd", true, log)
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
	_, err = ivdPETM.vsom.RetrieveSnapshotDetails(ctx, diskId, NewIDFromString(peSnapID.String()))
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
	_, err = ivdPE.DeleteSnapshot(ctx, peSnapID)
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
	peSnapID, err := ivdPE.Snapshot(ctx)
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

func getCreateSpec(name string, capacity int64, datastore types.ManagedObjectReference) (types.VslmCreateSpec) {
	keepAfterDeleteVm := true
	return types.VslmCreateSpec{
		Name:              name,
		KeepAfterDeleteVm: &keepAfterDeleteVm,
		BackingSpec:       &types.VslmCreateSpecDiskFileBackingSpec{
			VslmCreateSpecBackingSpec: types.VslmCreateSpecBackingSpec{
				Datastore:   datastore,
			},
		},
		CapacityInMB:      capacity,
	}
}

func getRandomName(prefix string, nDigits int) string {
	num := rand.Int63n(int64(math.Pow10(nDigits)))
	numstr := strconv.FormatInt(num, 10)
	return fmt.Sprintf("%s-%s", prefix, numstr)
}

func vmAttachDisk(ctx context.Context, client *vim25.Client, vm types.ManagedObjectReference, diskId types.ID, datastore types.ManagedObjectReference) (*object.Task, error) {
	req := types.AttachDisk_Task{
		This: vm.Reference(),
		DiskId: diskId,
		Datastore: datastore.Reference(),
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
		This: vm.Reference(),
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

func vmCreate(ctx context.Context, client *vim25.Client, vmHost types.ManagedObjectReference, vmName string, dsName string, logger logrus.FieldLogger) (*object.VirtualMachine, error) {
	finder := find.NewFinder(client)
	virtualMachineConfigSpec := types.VirtualMachineConfigSpec{
		Name: vmName,
		Files: &types.VirtualMachineFileInfo{
			VmPathName: "[" + dsName + "]",
		},
		NumCPUs:  1,
		MemoryMB: 4,
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