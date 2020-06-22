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

package astrolabe

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/vmware-tanzu/astrolabe/gen/models"
	"io"
	"log"
	"strings"
)

type ProtectedEntityID struct {
	peType     string
	id         string
	snapshotID ProtectedEntitySnapshotID
}

func NewProtectedEntityID(peType string, id string) ProtectedEntityID {
	return NewProtectedEntityIDWithSnapshotID(peType, id, ProtectedEntitySnapshotID{})
}

func NewProtectedEntityIDWithSnapshotID(peType string, id string, snapshotID ProtectedEntitySnapshotID) ProtectedEntityID {
	newID := ProtectedEntityID{
		peType:     peType,
		id:         id,
		snapshotID: snapshotID,
	}
	return newID
}

func NewProtectedEntityIDFromString(peiString string) (returnPEI ProtectedEntityID, returnError error) {
	returnError = fillInProtectedEntityIDFromString(&returnPEI, peiString)
	return returnPEI, returnError
}

func NewProtectedEntityIDFromModel(mpei models.ProtectedEntityID) (ProtectedEntityID, error){
	return NewProtectedEntityIDFromString(string(mpei))
}

func fillInProtectedEntityIDFromString(pei *ProtectedEntityID, peiString string) error {
	components := strings.Split(peiString, ":")
	if len(components) > 1 {
		pei.peType = components[0]
		pei.id = components[1]
		if len(components) == 3 {
			pei.snapshotID = NewProtectedEntitySnapshotID(components[2])
		}
		log.Print("pei = " + pei.String())
	} else {
		return errors.New("astrolabe: '" + peiString + "' is not a valid protected entity ID")
	}
	return nil
}
func (this ProtectedEntityID) GetID() string {
	return this.id
}

func (this ProtectedEntityID) GetPeType() string {
	return this.peType
}

func (this ProtectedEntityID) GetSnapshotID() ProtectedEntitySnapshotID {
	return this.snapshotID

}

func (this ProtectedEntityID) HasSnapshot() bool {
	return this.snapshotID.id != ""
}

func (this ProtectedEntityID) String() string {
	var returnString string
	returnString = this.peType + ":" + this.id
	if (this.snapshotID) != (ProtectedEntitySnapshotID{}) {
		returnString += ":" + this.snapshotID.String()
	}
	return returnString
}

func (this ProtectedEntityID) GetModelProtectedEntityID() models.ProtectedEntityID {
	return models.ProtectedEntityID(this.String())
}

func (this ProtectedEntityID) MarshalJSON() ([]byte, error) {
	return json.Marshal(this.String()) // Use marshal to make sure encoding happens
}

func (this *ProtectedEntityID) UnmarshalJSON(b []byte) error {
	var idStr string
	json.Unmarshal(b, &idStr) // Use unmarshall to make sure decoding happens
	log.Print("UnmarshalJSON idStr = " + idStr)
	return fillInProtectedEntityIDFromString(this, idStr)
}

type ProtectedEntitySnapshotID struct {
	// We should move this to actually being a UUID internally
	id string
}

func NewProtectedEntitySnapshotID(pesiString string) ProtectedEntitySnapshotID {
	returnPESI := ProtectedEntitySnapshotID{
		id: pesiString,
	}
	return returnPESI
}

func (this ProtectedEntitySnapshotID) GetID() string {
	return this.id
}

func (this ProtectedEntitySnapshotID) String() string {
	return this.id
}

func (this ProtectedEntitySnapshotID) GetModelProtectedEntitySnapshotID() models.ProtectedEntitySnapshotID {
	return models.ProtectedEntitySnapshotID(this.String())
}

type ProtectedEntity interface {
	GetInfo(ctx context.Context) (ProtectedEntityInfo, error)
	GetCombinedInfo(ctx context.Context) ([]ProtectedEntityInfo, error)
	/*
	 * Snapshot APIs
	 */
	Snapshot(ctx context.Context) (ProtectedEntitySnapshotID, error)
	ListSnapshots(ctx context.Context) ([]ProtectedEntitySnapshotID, error)
	DeleteSnapshot(ctx context.Context, snapshotToDelete ProtectedEntitySnapshotID) (bool, error)
	GetInfoForSnapshot(ctx context.Context, snapshotID ProtectedEntitySnapshotID) (*ProtectedEntityInfo, error)

	GetComponents(ctx context.Context) ([]ProtectedEntity, error)
	GetID() ProtectedEntityID

	// GetDataReader returns a reader for the data of the ProtectedEntity.  The ProtectedEntity will pick the
	// best data path to provide the Reader stream.  If the ProtectedEntity does not have any data, nil will be
	// returned
	GetDataReader(ctx context.Context) (io.ReadCloser, error)

	// GetMetadataReader returns a reader for the metadata of the ProtectedEntity.  The ProtectedEntity will pick the
	// best data path to provide the Reader stream.  If the ProtectedEntity does not have any metadata, nil will be
	// returned
	GetMetadataReader(ctx context.Context) (io.ReadCloser, error)
}
