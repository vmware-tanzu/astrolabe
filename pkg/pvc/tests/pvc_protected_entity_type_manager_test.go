package tests

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/magiconair/properties/assert"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/ivd"
	astrolabe_pvc "github.com/vmware-tanzu/astrolabe/pkg/pvc"
	"github.com/vmware-tanzu/astrolabe/pkg/s3repository"
	"github.com/vmware-tanzu/astrolabe/pkg/server"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
	"time"
)

func TestGetPVCComponents(t *testing.T) {
	path := os.Getenv("KUBECONFIG")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/whatever does not exist
		t.Skipf("The KubeConfig file, %v, is not exist", path)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		t.Fatalf("Failed to build k8s config from kubeconfig file: %+v ", err)
	}

	logger := logrus.New()
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = time.RFC3339Nano
	formatter.FullTimestamp = true
	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)

	// get ivd params about VC credential
	ivdParams := make(map[string]interface{})
	if err = ivd.RetrievePlatformInfoFromConfig(config, ivdParams); err != nil {
		t.Fatalf("Failed to retrieve VC config secret: %+v", err)
	}

	// get pvc params
	pvcParams := make(map[string]interface{})
	pvcParams["kubeconfigPath"] = path

	configParams := make(map[string]map[string]interface{})
	configParams["pvc"] = pvcParams
	configParams["ivd"] = ivdParams

	configInfo := server.NewConfigInfo(configParams, astrolabe.S3Config{
		URLBase: "VOID_URL",
	})

	pem := server.NewDirectProtectedEntityManagerFromParamMap(configInfo, logger)

	pvc_petm := pem.GetProtectedEntityTypeManager("pvc")
	if pvc_petm == nil {
		t.Fatal(err)
	}

	ctx := context.Background()
	peids, err := pvc_petm.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, curPEID := range peids {
		logger.Infof("curPEID = %v", curPEID.String())
		curPE, err := pvc_petm.GetProtectedEntity(ctx, curPEID)
		if err != nil {
			logger.Errorf("Failed to get PVC PE with PEID = %v", curPEID.String())
			t.Fatal(err)
		}
		componentPEs, err := curPE.GetComponents(ctx)
		if err != nil {
			logger.Errorf("Failed to get components for PVC PE with PEID = %v", curPEID.String())
			t.Fatal(err)
		}
		for _, componentPE := range componentPEs {
			logger.Infof("component PE ID = %v", componentPE.GetID().String())
		}
	}
}

func TestSnapshotOps(t *testing.T) {
	path := os.Getenv("KUBECONFIG")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/whatever does not exist
		t.Skipf("The KubeConfig file, %v, is not exist", path)
	}

	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		t.Fatalf("Failed to build k8s config from kubeconfig file: %+v ", err)
	}

	logger := logrus.New()
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = time.RFC3339Nano
	formatter.FullTimestamp = true
	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)

	// get ivd params about VC credential
	ivdParams := make(map[string]interface{})
	if err = ivd.RetrievePlatformInfoFromConfig(config, ivdParams); err != nil {
		t.Fatalf("Failed to retrieve VC config secret: %+v", err)
	}

	// get pvc params
	pvcParams := make(map[string]interface{})
	pvcParams["kubeconfigPath"] = path

	configParams := make(map[string]map[string]interface{})
	configParams["pvc"] = pvcParams
	configParams["ivd"] = ivdParams

	configInfo := server.NewConfigInfo(configParams, astrolabe.S3Config{
		URLBase: "VOID_URL",
	})

	pem := server.NewDirectProtectedEntityManagerFromParamMap(configInfo, logger)

	pvc_petm := pem.GetProtectedEntityTypeManager("pvc")
	if pvc_petm == nil {
		t.Fatal("Failed to get PVC ProtectedEntityTypeManager")
	}

	ctx := context.Background()
	peids, err := pvc_petm.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(peids) <= 0 {
		t.Skip("No PVC can be found in the cluster")
	}
	selectedPEID := peids[0]
	logger.Infof("Picked up the first available PVC PE, %v", selectedPEID.String())

	pvcPE, err := pvc_petm.GetProtectedEntity(ctx, selectedPEID)
	if err != nil {
		logger.Errorf("Failed to get PVC PE with PEID = %v", selectedPEID.String())
		t.Fatal(err)
	}

	peSnapshotIDs, err := pvcPE.ListSnapshots(ctx)
	if err != nil {
		logger.Errorf("Failed to list snapshots for the PVC PE with PEID = %v", pvcPE.GetID().String())
		t.Fatal(err)
	}

	prevSnapshotsNum := len(peSnapshotIDs)
	logger.Infof("There are %v snapshots for the PVC PE, %v, before snapshotting it", prevSnapshotsNum, pvcPE.GetID().String())

	logger.Infof("Snapshotting the PVC PE, %v", selectedPEID.String())
	peSnapshotID, err := pvcPE.Snapshot(ctx, make(map[string]map[string]interface{}))
	if err != nil {
		logger.Errorf("Failed to snapshot PVC PE with PEID = %v", pvcPE.GetID().String())
		t.Fatal(err)
	}
	logger.Infof("Snapshotted the PVC PE, %v with the snapshot ID as, %v", pvcPE.GetID().String(), peSnapshotID.String())

	defer func() {
		logger.Infof("Deleting snapshot, %v, for the PVC PE, %v", peSnapshotID.String(), pvcPE.GetID().String())
		success, err := pvcPE.DeleteSnapshot(ctx, peSnapshotID, make(map[string]map[string]interface{}))
		if !success || err != nil {
			logger.Errorf("Failed to delete snapshot, %v, for PVC PE with PEID = %v", peSnapshotID.String(), pvcPE.GetID().String())
		}
		logger.Infof("Deleted snapshot, %v, for the PVC PE, %v", peSnapshotID.String(), pvcPE.GetID().String())
	}()

	peSnapshotIDs, err = pvcPE.ListSnapshots(ctx)
	if err != nil {
		logger.Errorf("Failed to list snapshots for the PVC PE with PEID = %v", pvcPE.GetID().String())
		t.Fatal(err)
	}

	curSnapshotsNum := len(peSnapshotIDs)
	logger.Infof("There are %v snapshots for the PVC PE, %v, after snapshotting it", curSnapshotsNum, pvcPE.GetID().String())

	assert.Equal(t, curSnapshotsNum-prevSnapshotsNum, 1, "there should be one more snapshot available")
}

func TestCreateVolumeFromMetadata(t *testing.T) {
	path := os.Getenv("KUBECONFIG")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/whatever does not exist
		t.Skipf("The KubeConfig file, %v, is not exist", path)
	}

	logger := logrus.New()
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = time.RFC3339Nano
	formatter.FullTimestamp = true
	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)

	// get pvc params
	pvcParams := make(map[string]interface{})
	pvcParams["kubeconfigPath"] = path

	configParams := make(map[string]map[string]interface{})
	configParams["pvc"] = pvcParams

	configInfo := server.NewConfigInfo(configParams, astrolabe.S3Config{
		URLBase: "VOID_URL",
	})

	pem := server.NewDirectProtectedEntityManagerFromParamMap(configInfo, logger)

	pvc_petm := pem.GetProtectedEntityTypeManager("pvc")
	if pvc_petm == nil {
		t.Fatal("Failed to get PVC ProtectedEntityTypeManager")
	}

	ctx := context.Background()
	peids, err := pvc_petm.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(peids) <= 0 {
		t.Skip("No PVC can be found in the cluster")
	}
	selectedPEID := peids[0]
	logger.Infof("Picked up the first available PVC PE, %v", selectedPEID.String())

	pvcPE, err := pvc_petm.GetProtectedEntity(ctx, selectedPEID)
	if err != nil {
		logger.Errorf("Failed to get PVC PE with PEID = %v", selectedPEID.String())
		t.Fatal(err)
	}

	pvc, err := pvcPE.(astrolabe_pvc.PVCProtectedEntity).GetPVC()
	if err != nil {
		logger.Errorf("Failed to get PVC with PEID = %v", selectedPEID.String())
		t.Fatal(err)
	}

	pvc.Name = "new-test-pvc"
	pvcbytes, err := pvc.Marshal()
	if err != nil {
		logger.Errorf("Failed to marshal PVC %s/%s", pvc.Name, pvc.Namespace)
		t.Fatal(err)
	}

	newPE, err := pvc_petm.(*astrolabe_pvc.PVCProtectedEntityTypeManager).CreateFromMetadata(ctx, pvcbytes, astrolabe.ProtectedEntityID{}, nil,
		"", "")
	if err != nil {
		logger.Errorf("Failed to create volume from metadata: %s/%s", pvc.Name, pvc.Namespace)
		t.Fatal(err)
	}

	logger.Infof("Created new PVC PE: %s", newPE.GetID().String())
}

func TestCreateVolumeFromMetadataAndS3(t *testing.T) {
	path := os.Getenv("KUBECONFIG")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// path/to/whatever does not exist
		t.Skipf("The KubeConfig file, %v, is not exist", path)
	}
	config, err := clientcmd.BuildConfigFromFlags("", path)
	if err != nil {
		t.Fatalf("Failed to build k8s config from kubeconfig file: %+v ", err)
	}
	logger := logrus.New()
	formatter := new(logrus.TextFormatter)
	formatter.TimestampFormat = time.RFC3339Nano
	formatter.FullTimestamp = true
	logger.SetFormatter(formatter)
	logger.SetLevel(logrus.DebugLevel)

	// get pvc params
	pvcParams := make(map[string]interface{})
	pvcParams["kubeconfigPath"] = path

	configParams := make(map[string]map[string]interface{})
	configParams["pvc"] = pvcParams

	// Step 1: To create the IVD PETM, get all PEs and select one as the reference.
	ivdParams := make(map[string]interface{})
	if err = ivd.RetrievePlatformInfoFromConfig(config, ivdParams); err != nil {
		t.Fatalf("Failed to retrieve VC config secret: %+v", err)
	}
	configParams["ivd"] = ivdParams
	configInfo := server.NewConfigInfo(configParams, astrolabe.S3Config{
		URLBase: "VOID_URL",
	})

	pem := server.NewDirectProtectedEntityManagerFromParamMap(configInfo, logger)

	pvc_petm := pem.GetProtectedEntityTypeManager("pvc")
	if pvc_petm == nil {
		t.Fatal("Failed to get PVC ProtectedEntityTypeManager")
	}

	ctx := context.Background()
	s3petm, err := setupS3PETM(t, "ivd", logger)
	if err != nil {
		t.Fatal(err)
	}

	peids, err := pvc_petm.GetProtectedEntities(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(peids) <= 0 {
		t.Skip("No PVC can be found in the cluster")
	}
	selectedPEID := peids[0]
	logger.Infof("Picked up the first available PVC PE, %v", selectedPEID.String())

	pvcPE, err := pvc_petm.GetProtectedEntity(ctx, selectedPEID)
	if err != nil {
		logger.Errorf("Failed to get PVC PE with PEID = %v", selectedPEID.String())
		t.Fatal(err)
	}

	snapshotID, err := pvcPE.Snapshot(ctx, make(map[string]map[string]interface{}))
	if err != nil {
		logger.Errorf("Snapshot failed for PEID = %v", selectedPEID.String())
		t.Fatal(err)
	}

	snapshotPEID := pvcPE.GetID().IDWithSnapshot(snapshotID)
	snapshotPE, err := pvc_petm.GetProtectedEntity(ctx, snapshotPEID)
	if err != nil {
		logger.Errorf("Snapshot failed for PEID = %v", snapshotPEID.String())
		t.Fatal(err)
	}
	mdReader, err := snapshotPE.GetMetadataReader(ctx)
	if err != nil {
		logger.Errorf("Could not get MD reader for PEID = %v", snapshotPEID.String())
		t.Fatal(err)
	}
	pvcbytes, err := ioutil.ReadAll(mdReader)
	if err != nil {
		logger.Errorf("Could not read MD for PEID = %v", snapshotPEID.String())
		t.Fatal(err)
	}
	snapshotComponents, err := snapshotPE.GetComponents(ctx)
	if err != nil {
		logger.Errorf("Could not get components for PEID = %v", snapshotPEID.String())
		t.Fatal(err)
	}
	_, err = s3petm.Copy(ctx, snapshotComponents[0], make(map[string]map[string]interface{}), astrolabe.AllocateNewObject)
	if err != nil {
		logger.Errorf("Copy failed for PEID = %v", snapshotComponents[0].GetID())
		t.Fatal(err)
	}

	pvc_petm.(*astrolabe_pvc.PVCProtectedEntityTypeManager).CreateFromMetadata(ctx, pvcbytes, snapshotPEID, s3petm, "", "")
}
func setupS3PETM(t *testing.T, typeName string, logger *logrus.Logger) (*s3repository.ProtectedEntityTypeManager, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1")},
	)
	if err != nil {
		return nil, err
	}
	s3petm, err := s3repository.NewS3RepositoryProtectedEntityTypeManager(typeName, *sess, "velero-plugin-s3-repo",
		"backups/vsphere-volumes-repo/", logger)
	if err != nil {
		return nil, err
	}
	return s3petm, err
}