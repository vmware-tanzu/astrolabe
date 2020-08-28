package pvc

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"io/ioutil"
	core_v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"strings"
)

/*
PVCProtectedEntity implements a Protected Entity interface to PVCs and a basic local snapshot facility.  The PVC
has a component PE which will be the PE that maps to the volume referenced by the PVC.  Currently IVD and para-virt PV
are available

The PVC at the time of the snapshot is serialized and stored in a config map in the namespace named
pvc-snap.<pvc name>.  Each snapshot has an entry in the binary data map for the serialized PVC data, named by the
snapshot ID (this is the snapshot ID returned by the subcomponent)
*/

const (
	VSphereCSIProvisioner = "csi.vsphere.vmware.com"
	PEInfoPrefix          = "peinfo"
)

type PVCProtectedEntity struct {
	ppetm    *PVCProtectedEntityTypeManager
	id       astrolabe.ProtectedEntityID
	data     []astrolabe.DataTransport
	metadata []astrolabe.DataTransport
	combined []astrolabe.DataTransport
	logger   logrus.FieldLogger
}

func newPVCProtectedEntity(ppetm *PVCProtectedEntityTypeManager, peid astrolabe.ProtectedEntityID) (PVCProtectedEntity, error) {
	if peid.GetPeType() != astrolabe.PvcPEType {
		return PVCProtectedEntity{}, errors.Errorf("%s is not a PVC PEID", peid.String())
	}
	data, metadata, combined, err := ppetm.getDataTransports(peid)
	if err != nil {
		return PVCProtectedEntity{}, errors.Wrap(err, "Failed to get data transports")
	}

	returnPE := PVCProtectedEntity{
		ppetm:    ppetm,
		id:       peid,
		data:     data,
		metadata: metadata,
		combined: combined,
		logger:   ppetm.logger,
	}
	return returnPE, nil
}

func (this PVCProtectedEntity) GetInfo(ctx context.Context) (astrolabe.ProtectedEntityInfo, error) {
	pvc, err := this.GetPVC()
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve PVC")
	}
	components, err := this.GetComponents(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve component")
	}
	retPEInfo := astrolabe.NewProtectedEntityInfo(this.id, pvc.Name, this.data, this.metadata, this.combined,
		[]astrolabe.ProtectedEntityID{components[0].GetID()})
	return retPEInfo, nil
}

func (this PVCProtectedEntity) GetCombinedInfo(ctx context.Context) ([]astrolabe.ProtectedEntityInfo, error) {
	panic("implement me")
}

func (this PVCProtectedEntity) Snapshot(ctx context.Context, params map[string]map[string]interface{}) (astrolabe.ProtectedEntitySnapshotID, error) {
	if this.id.HasSnapshot() {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.New("Cannot create snapshot of snapshot")
	}
	pvc, err := this.GetPVC()
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrap(err, "Could not retrieve pvc")
	}
	components, err := this.GetComponents(ctx)
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrap(err, "Could not retrieve components")
	}
	if len(components) != 1 {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.New(fmt.Sprintf("Expected 1 component, %s has %d", this.id.String(), len(components)))
	}
	// PVC will always have only one component, i.e., the PV it claims.
	subSnapshotID, err := components[0].Snapshot(ctx, params)
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrapf(err, "Subcomponent peid %s snapshot failed", components[0].GetID())
	}
	this.logger.Infof("PVCProtectedEntity: Received snapshotID : %s after component snapshot of pe-id %s", subSnapshotID.String(), components[0].GetID().String())
	// TODO: add configurable flag here.
	if false {
		snapConfigMapName := GetSnapConfigMapName(pvc)
		snapConfigMap, err := this.ppetm.clientSet.CoreV1().ConfigMaps(pvc.Namespace).Get(ctx, snapConfigMapName, metav1.GetOptions{})
		var binaryData map[string][]byte
		var create bool
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrapf(err, "Could not retrieve snapshot configmap %s for %s", snapConfigMapName,
					this.id.String())
			}

			snapConfigMap = &core_v1.ConfigMap{
				ObjectMeta: metav1.ObjectMeta{
					Name:      snapConfigMapName,
					Namespace: pvc.Namespace,
				},
			}
			create = true
		} else {
			binaryData = snapConfigMap.BinaryData
		}
		pvcData, err := pvc.Marshal()
		if err != nil {
			return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrapf(err, "Could not marshal PVC data for %v",
				this.id.String())
		}

		if binaryData == nil {
			binaryData = make(map[string][]byte)
		}
		binaryData[subSnapshotID.String()] = pvcData
		peInfo, err := this.GetInfo(ctx)
		if err != nil {
			return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrapf(err, "Could not retrieve PE info for %v",
				this.id.String())
		}
		peInfoData, err := json.Marshal(peInfo)
		if err != nil {
			return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrapf(err, "Marshal peid info %s failed", components[0].GetID())
		}
		binaryData[PEInfoPrefix+"-"+subSnapshotID.String()] = peInfoData
		snapConfigMap.BinaryData = binaryData
		if create {
			_, err = this.ppetm.clientSet.CoreV1().ConfigMaps(pvc.Namespace).Create(ctx, snapConfigMap, metav1.CreateOptions{})
		} else {
			_, err = this.ppetm.clientSet.CoreV1().ConfigMaps(pvc.Namespace).Update(ctx, snapConfigMap, metav1.UpdateOptions{})
		}
		if err != nil {
			return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrapf(err, "PVC Snapshot ConfigMap write peid %s failed", components[0].GetID())
		}
		this.logger.Infof("PVCProtectedEntity: Updated the config map : %s with snap-id : %s", snapConfigMapName, subSnapshotID.String())
	}
	subSnapshotPEID := components[0].GetID().IDWithSnapshot(subSnapshotID)
	subSnapshotStr := base64.RawStdEncoding.EncodeToString([]byte(subSnapshotPEID.String()))
	returnSnapshotID := astrolabe.NewProtectedEntitySnapshotID(subSnapshotStr)
	return returnSnapshotID, nil
}

func (this PVCProtectedEntity) ListSnapshots(ctx context.Context) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	this.logger.Warnf("PVC PE ListSnapshots invoked, returning empty results as snapshot persistence in config map is not available.")
	returnIDs := make([]astrolabe.ProtectedEntitySnapshotID, 0)
	if false {
		pvc, err := this.GetPVC()
		if err != nil {
			return nil, errors.Wrap(err, fmt.Sprintf("Could not retrieve pvc peid=%s", this.id.String()))
		}
		snapConfigMapName := GetSnapConfigMapName(pvc)
		snapConfigMap, err := this.ppetm.clientSet.CoreV1().ConfigMaps(pvc.Namespace).Get(ctx, snapConfigMapName, metav1.GetOptions{})
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return nil, errors.Wrapf(err, "Could not retrieve snapshot configmap %s for %s", snapConfigMapName,
					this.id.String())
			}
			// No configmap so no snapshots
			return []astrolabe.ProtectedEntitySnapshotID{}, nil
		}
		for snapshotIDStr, _ := range snapConfigMap.BinaryData {
			if strings.Contains(snapshotIDStr, PEInfoPrefix) {
				continue
			}
			returnIDs = append(returnIDs, astrolabe.NewProtectedEntitySnapshotID(snapshotIDStr))
		}
	}
	return returnIDs, nil
}

func GetSnapConfigMapName(pvc *core_v1.PersistentVolumeClaim) string {
	return "pvc-snap." + pvc.Name
}

func (this PVCProtectedEntity) DeleteSnapshot(ctx context.Context, snapshotToDelete astrolabe.ProtectedEntitySnapshotID, params map[string]map[string]interface{}) (bool, error) {
	this.logger.Infof("PVCProtectedEntity: DeleteSnapshot request received on snapshot: %s with parameters: %v on the PVC PE: %s", snapshotToDelete.String(), params, this.GetID().String())
	pvc, err := this.GetPVC()
	if err != nil {
		return false, errors.Wrap(err, fmt.Sprintf("Could not retrieve pvc peid=%s", this.id.String()))
	}

	components, err := this.GetComponents(ctx)
	if err != nil {
		return false, errors.Wrap(err, "Could not retrieve components")
	}
	if len(components) != 1 {
		return false, errors.New(fmt.Sprintf("Expected 1 component, %s has %d", this.id.String(), len(components)))
	}
	this.logger.Infof("PVCProtectedEntity: Retrieved components for the pvc protected entity id: %s", components[0].GetID().String())
	componentSnapshotID := components[0].GetID().GetSnapshotID()
	this.logger.Infof("snapshotToDelete: %s  componentSnapshotID: %s", snapshotToDelete.String(), componentSnapshotID.String())
	deleteSnapStatus := false
	this.logger.Infof("PVCProtectedEntity: Triggering DeleteSnapshot with component snapshotID: %s for the component pe-id: %s", componentSnapshotID.String(), components[0].GetID().String())
	_, err = components[0].DeleteSnapshot(ctx, componentSnapshotID, params)
	if err != nil {
		this.logger.Warnf("Subcomponent peid %s delete snapshot failed error: %v", components[0].GetID(), err)
	} else {
		deleteSnapStatus = true
		this.logger.Infof("PVCProtectedEntity: Completed DeleteSnapshot with snapshotID: %s for the component: %s", snapshotToDelete.String(), components[0].GetID().String())
	}

	if false {
		snapConfigMapName := GetSnapConfigMapName(pvc)
		snapConfigMap, err := this.ppetm.clientSet.CoreV1().ConfigMaps(pvc.Namespace).Get(ctx, snapConfigMapName, metav1.GetOptions{})
		if err != nil {
			if !k8serrors.IsNotFound(err) {
				return false, errors.Wrapf(err, "Could not retrieve snapshot configmap %s for %s", snapConfigMapName,
					this.id.String())
			}
			this.logger.Infof("PVCProtectedEntity: No snapshot config map avaiable with name: %s", snapConfigMapName)
		} else {
			_, snapPVCInfoExists := snapConfigMap.BinaryData[componentSnapshotID.String()]
			if snapPVCInfoExists {
				delete(snapConfigMap.BinaryData, componentSnapshotID.String())
				delete(snapConfigMap.BinaryData, "peinfo-"+componentSnapshotID.String())
				updatedSnapConfigMap, err := this.ppetm.clientSet.CoreV1().ConfigMaps(pvc.Namespace).Update(ctx, snapConfigMap, metav1.UpdateOptions{})
				if err != nil {
					return false, errors.Wrapf(err, "Could not update snapshot configmap %s for %s", snapConfigMapName,
						this.id.String())
				}
				if updatedSnapConfigMap.BinaryData == nil {
					// no snapshot after the update. So, clean up the config map
					var zeroSecondsGracePeriod int64
					zeroSecondsGracePeriod = 0
					err := this.ppetm.clientSet.CoreV1().ConfigMaps(pvc.Namespace).Delete(ctx, updatedSnapConfigMap.Name, metav1.DeleteOptions{GracePeriodSeconds: &zeroSecondsGracePeriod})
					if err != nil {
						return false, errors.Wrapf(err, "Could not delete snapshot configmap %s for %s", snapConfigMapName,
							this.id.String())
					}
				}
				this.logger.Infof("PVCProtectedEntity: Successfully deleted snapshot entries in the config map: %s for snapshot: %s", snapConfigMapName, snapshotToDelete.String())
			} else {
				this.logger.Infof("PVCProtectedEntity: No entries found in config map : %s for the snapshot-id: %s", snapConfigMapName, snapshotToDelete.String())
			}
		}
	}
	// If the subsnapshot existed, then we successfully removed, otherwise
	// there was no work (errors would have exited us already) so we return false
	return deleteSnapStatus, nil
}

func (this PVCProtectedEntity) GetInfoForSnapshot(ctx context.Context, snapshotID astrolabe.ProtectedEntitySnapshotID) (*astrolabe.ProtectedEntityInfo, error) {
	panic("implement me")
}

func (this PVCProtectedEntity) GetComponents(ctx context.Context) ([]astrolabe.ProtectedEntity, error) {
	this.logger.Infof("PVCProtectedEntity: Attempting to retrieve components for pe-id: %s", this.GetID().String())
	pvc, err := this.GetPVC()
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve PVC")
	}

	if pvc.Status.Phase != core_v1.ClaimBound {
		this.logger.Infof("No bound PV for the PVC, %s. So, there is no component for the PVC PE, %s", pvc.Name, this.id.String())
		return []astrolabe.ProtectedEntity{}, nil
	}

	pv, err := this.ppetm.clientSet.CoreV1().PersistentVolumes().Get(ctx, pvc.Spec.VolumeName, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrap(err, "Could not retrieve Persistent Volume")
	}

	pvPE, err := this.getProtectedEntityForPV(ctx, pv)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not find subcomponent PEID for %s %s %s", this.id.String(), pvc.Namespace, pvc.Name)
	}
	return []astrolabe.ProtectedEntity{pvPE}, nil
}

/*
Get ProtectedEntity for Persistent Volume referenced by this PVC ProtectedEntity.
Candidates are, IVD ProtectedEntity(non-GuestCluster) and ParaVirt ProtectedEntity(GuestCluster).
*/
func (this PVCProtectedEntity) getProtectedEntityForPV(ctx context.Context, pv *core_v1.PersistentVolume) (astrolabe.ProtectedEntity, error) {
	if pv.Spec.CSI != nil {
		if pv.Spec.CSI.Driver == VSphereCSIProvisioner {
			if pv.Spec.AccessModes[0] == core_v1.ReadWriteOnce {
				var pvIDstr string
				if this.ppetm.isGuest {
					pvIDstr = pv.Name // use pv name rather than pv volume handle as the ID of paravirt PE, since it is easier to retrieve pv volume handle from pv name
				} else {
					pvIDstr = pv.Spec.CSI.VolumeHandle
				}
				pvPEType := this.getComponentPEType()
				var pvPEID astrolabe.ProtectedEntityID
				var err error
				if this.GetID().HasSnapshot() {
					pvPEID, err = getPEIDForComponentSnapshot(this.GetID(), this.logger)
					if err != nil {
						return nil, errors.Wrapf(err, "Could not decode component snapshot ID for %s", this.GetID().String())
					}
					this.logger.Infof("The pvPEID: %s with snapshot was decoded", pvPEID.String())
				} else {
					pvPEID = astrolabe.NewProtectedEntityID(pvPEType, pvIDstr)
				}
				pvPE, err := this.ppetm.pem.GetProtectedEntity(ctx, pvPEID)
				if err != nil {
					return nil, errors.Wrapf(err, "Could not get Protected Entity for PV %s", pvPEID.String())
				}
				this.logger.Infof("getProtectedEntityForPV: type: %s, pvIDstr: %s, snapshot: %s", pvPEType, pvIDstr, this.id.GetSnapshotID())
				return pvPE, nil
			} else {
				return nil, errors.Errorf("Unexpected access mode, %v, for Persistent Volume %s", pv.Spec.AccessModes[0], pv.Name)
			}
		}
	}
	return nil, errors.Errorf("Could not find PE for Persistent Volume %s", pv.Name)
}

func (this PVCProtectedEntity) GetID() astrolabe.ProtectedEntityID {
	return this.id
}

func (this PVCProtectedEntity) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (this PVCProtectedEntity) GetMetadataReader(ctx context.Context) (io.ReadCloser, error) {
	if this.id.HasSnapshot() {
		// NOTE(xyang): Remove panic here so that we can call
		// GetMetadataReader in createSnapshot in backupdriver
		//panic("Fix me - snapshot MD retrieval not implemented")
		this.logger.Infof("Fix me - snapshot MD retrieval not implemented")
	}
	pvc, err := this.GetPVC()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get pvc")
	}
	pvcBytes, err := pvc.Marshal()
	return ioutil.NopCloser(bytes.NewReader(pvcBytes)), nil
}

func (this PVCProtectedEntity) GetPVC() (*core_v1.PersistentVolumeClaim, error) {
	namespace, name, err := astrolabe.GetNamespaceAndNameFromPEID(this.id)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get namespace and id from PEID %s", this.id.String())
	}
	pvc, err := this.ppetm.clientSet.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get retrieve pvc with namespace %s, id %s", namespace, name)
	}
	return pvc, nil
}

func (this PVCProtectedEntity) getComponentPEType() string {
	var peType string
	typeManagers := this.ppetm.pem.ListEntityTypeManagers()
	// always return the first non-pvc type as PVC PE will always have only one component, either ivd or paravirt-pv
	for _, typeManager := range typeManagers {
		peType := typeManager.GetTypeName()
		if peType == astrolabe.PvcPEType {
			continue
		}
		return peType
	}

	return peType
}

func (this PVCProtectedEntity) Overwrite(ctx context.Context, sourcePE astrolabe.ProtectedEntity, params map[string]map[string]interface{},
	overwriteComponents bool) error {

	return nil
}
