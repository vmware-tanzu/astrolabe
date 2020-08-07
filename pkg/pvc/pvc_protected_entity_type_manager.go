package pvc

import (
	"context"
	"encoding/base64"
	"fmt"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/util"
	v1 "k8s.io/api/core/v1"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type PVCProtectedEntityTypeManager struct {
	clientSet *kubernetes.Clientset
	isGuest   bool
	pem       astrolabe.ProtectedEntityManager
	s3Config  astrolabe.S3Config
	logger    logrus.FieldLogger
}

func NewProtectedEntityIDFromPVCName(namespace string, pvcName string) astrolabe.ProtectedEntityID {
	return astrolabe.NewProtectedEntityIDWithNamespace(astrolabe.PvcPEType, pvcName, namespace)
}

/*
Creates a PVC ProtectectEntityTypeManager
K8S configuration is provided through parameters
restConfig - *rest.Config - if set, this will be used
if restConfig is not set, masterURL and kubeConfigPath will be used.  Either can be set
*/
func NewPVCProtectedEntityTypeManagerFromConfig(params map[string]interface{}, s3Config astrolabe.S3Config,
	logger logrus.FieldLogger) (*PVCProtectedEntityTypeManager, error) {
	var err error
	var config *rest.Config
	config, ok := params["restConfig"].(*rest.Config)
	if !ok {
		masterURL, _ := util.GetStringFromParamsMap(params, "masterURL", logger)
		kubeconfigPath, _ := util.GetStringFromParamsMap(params, "kubeconfigPath", logger)
		config, err = clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
		if err != nil {
			return nil, err
		}
	}
	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	_, isGuest := params["svcNamespace"]
	return &PVCProtectedEntityTypeManager{
		clientSet: clientSet,
		isGuest:   isGuest,
		s3Config:  s3Config,
		logger:    logger,
	}, nil
}

func (this *PVCProtectedEntityTypeManager) SetProtectedEntityManager(pem astrolabe.ProtectedEntityManager) {
	this.pem = pem
}

func (this *PVCProtectedEntityTypeManager) GetTypeName() string {
	return astrolabe.PvcPEType
}

func (this *PVCProtectedEntityTypeManager) GetProtectedEntity(ctx context.Context, peid astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	namespace, name, err := astrolabe.GetNamespaceAndNameFromPEID(peid)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get namespace and name from peid %s", peid.String())
	}

	returnPE, err := newPVCProtectedEntity(this, peid)

	if err != nil {
		return nil, errors.Wrapf(err, "Could not create PVCProtectedEntity for namespace = %s, name = %s", namespace, name)
	}

	_, err = returnPE.GetPVC()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not retrieve PVC for namespace = %s, name = %s", namespace, name)
	}
	return returnPE, nil
}

func (this *PVCProtectedEntityTypeManager) GetProtectedEntities(ctx context.Context) ([]astrolabe.ProtectedEntityID, error) {
	nsList, err := this.clientSet.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Could not list namespaces")
	}
	retPEIDs := make([]astrolabe.ProtectedEntityID, 0)
	for _, ns := range nsList.Items {
		pvcList, err := this.clientSet.CoreV1().PersistentVolumeClaims(ns.Name).List(metav1.ListOptions{})
		if err != nil {
			return nil, errors.Wrap(err, "Could not list PVCs")
		}

		for _, curPVC := range pvcList.Items {
			retPEIDs = append(retPEIDs, NewProtectedEntityIDFromPVCName(curPVC.Namespace, curPVC.Name))
		}
	}
	return retPEIDs, nil
}

func (this *PVCProtectedEntityTypeManager) Copy(ctx context.Context, pe astrolabe.ProtectedEntity,
	params map[string]map[string]interface{}, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (this *PVCProtectedEntityTypeManager) CopyFromInfo(ctx context.Context, info astrolabe.ProtectedEntityInfo,
	params map[string]map[string]interface{}, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (this *PVCProtectedEntityTypeManager) getDataTransports(id astrolabe.ProtectedEntityID) ([]astrolabe.DataTransport,
	[]astrolabe.DataTransport,
	[]astrolabe.DataTransport, error) {

	data := []astrolabe.DataTransport{}

	mdS3Transport, err := astrolabe.NewS3MDTransportForPEID(id, this.s3Config)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Could not create S3 md transport")
	}

	md := []astrolabe.DataTransport{
		mdS3Transport,
	}

	combinedS3Transport, err := astrolabe.NewS3CombinedTransportForPEID(id, this.s3Config)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Could not create S3 combined transport")
	}

	combined := []astrolabe.DataTransport{
		combinedS3Transport,
	}

	return data, md, combined, nil
}

// CreateFromMetadata creates a new PVC (dynamic provisioning path) with serialized PVC info
func (this *PVCProtectedEntityTypeManager) CreateFromMetadata(ctx context.Context, buf []byte,
	sourceSnapshotID astrolabe.ProtectedEntityID, componentSourcePETM astrolabe.ProtectedEntityTypeManager, cloneFromSnapshotNamespace string, cloneFromSnapshotName string) (astrolabe.ProtectedEntity, error) {
	pvc := v1.PersistentVolumeClaim{}
	err := pvc.Unmarshal(buf)
	if err != nil {
		return nil, err
	}
	this.logger.Infof("CreateFromMetadata: retrieve PVC %s/%s from metadata: %v", pvc.Namespace, pvc.Name, pvc)
	peID := NewProtectedEntityIDFromPVCName(pvc.Namespace, pvc.Name)
	this.logger.Infof("CreateFromMetadata: generated peID: %s", peID.String())

	// Clears out fields before creating a new PVC in the API server
	pvc.CreationTimestamp = metav1.Time{}
	pvc.DeletionTimestamp = nil
	annotations := map[string]string{}
	pvc.Annotations = annotations
	var finalizers []string
	pvc.Finalizers = finalizers
	pvc.ResourceVersion = ""
	pvc.UID = ""
	pvc.Spec.VolumeName = ""
	pvc.Status = v1.PersistentVolumeClaimStatus{}

	var pvcPE astrolabe.ProtectedEntity
	dynamic := true
	if dynamic {

		// Creates a new PVC (dynamic provisioning path)
		if _, err = this.clientSet.CoreV1().PersistentVolumeClaims(pvc.Namespace).Create(&pvc); err == nil || apierrs.IsAlreadyExists(err) {
			// Save succeeded.
			if err != nil {
				this.logger.Infof("PVC %s/%s already exists, reusing", pvc.Namespace, pvc.Name)
				err = nil
			} else {
				this.logger.Infof("PVC %s/%s saved", pvc.Namespace, pvc.Name)
			}
		}
		if err != nil {
			return nil, fmt.Errorf("failed to create PVC %s/%s in the API server: %v", pvc.Namespace, pvc.Name, err)
		}
		this.logger.Infof("CreateFromMetadata: created PVC: %s/%s", pvc.Namespace, pvc.Name)

		err = WaitForPersistentVolumeClaimPhase(v1.ClaimBound, this.clientSet, pvc.Namespace, pvc.Name, Poll, ClaimBindingTimeout, this.logger)
		if err != nil {
			return nil, fmt.Errorf("PVC %q did not become Bound: %v", pvc.Name, err)
		}

		this.logger.Infof("CreateFromMetadata: PVC %s/%s is bound.", pvc.Namespace, pvc.Name)
		pvcPE, err = this.pem.GetProtectedEntity(ctx, peID)
		if err != nil {
			errorMsg := fmt.Sprintf("Failed to get the PVC ProtectedEntity from peID %s", peID.String())
			this.logger.WithError(err).Error(errorMsg)
			return nil, errors.Wrap(err, errorMsg)
		}

		if componentSourcePETM != nil {
			// Get the PE for the PV and overwrite it
			components, err := pvcPE.GetComponents(ctx)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to get the components from peID %s", peID.String())
				this.logger.WithError(err).Error(errorMsg)
				return nil, errors.Wrap(err, errorMsg)
			}

			// Need to extract component snapshot ID from sourceSnapshotID here
			componentSnapshotPEID, err := getPEIDForComponentSnapshot(sourceSnapshotID, this.logger)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to decode the component ID from %s", sourceSnapshotID.String())
				this.logger.WithError(err).Error(errorMsg)
				return nil, errors.Wrap(err, errorMsg)
			}
			sourcePE, err := componentSourcePETM.GetProtectedEntity(ctx, componentSnapshotPEID)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to get the source snapshot PE for peID %s", componentSnapshotPEID.String())
				this.logger.WithError(err).Error(errorMsg)
				return nil, errors.Wrap(err, errorMsg)
			}
			overwriteParams := make(map[string]map[string]interface{})
			cloneParams := make(map[string]interface{})
			cloneParams["CloneFromSnapshotNamespace"] = cloneFromSnapshotNamespace
			cloneParams["CloneFromSnapshotName"] = cloneFromSnapshotName
			overwriteParams["CloneFromSnapshotReference"] = cloneParams
			err = components[0].Overwrite(ctx, sourcePE, overwriteParams, false)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to get the source snapshot PE for peID %s", componentSnapshotPEID.String())
				this.logger.WithError(err).Error(errorMsg)
				return nil, errors.Wrap(err, errorMsg)
			}
		}
	} else {
		// Handle static provisioning path
	}

	this.logger.Infof("CreateFromMetadata: retrieved ProtectedEntity for ID %s", peID.String())
	return pvcPE, nil
}

func getPEIDForComponentSnapshot(sourceSnapshotID astrolabe.ProtectedEntityID, logger logrus.FieldLogger) (astrolabe.ProtectedEntityID, error) {
	componentID64Str := sourceSnapshotID.GetSnapshotID().String()
	componentIDBytes, err := base64.StdEncoding.DecodeString(componentID64Str)
	if err != nil {
		errorMsg := fmt.Sprintf("Could not decode snapshot ID encoded string %s", componentID64Str)
		logger.WithError(err).Error(errorMsg)
		return astrolabe.ProtectedEntityID{}, errors.Wrap(err, errorMsg)
	}

	return astrolabe.NewProtectedEntityIDFromString(string(componentIDBytes))
}
