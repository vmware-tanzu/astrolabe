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

package server

import (
	"context"
	"github.com/labstack/echo"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"net/http"
)

type ServiceAPI struct {
	petm astrolabe.ProtectedEntityTypeManager
}

func NewServiceAPI(petm astrolabe.ProtectedEntityTypeManager) *ServiceAPI {
	return &ServiceAPI{
		petm: petm,
	}
}

func (this *ServiceAPI) listObjects(echoContext echo.Context) error {

	pes, err := this.petm.GetProtectedEntities(context.Background())
	if err != nil {
		return err
	}
	var pesList []string
	for _, curPes := range pes {
		pesList = append(pesList, curPes.GetID())
	}
	echoContext.JSON(http.StatusOK, pesList)
	return nil
}

func (this *ServiceAPI) handleObjectRequest(echoContext echo.Context) error {
	idStr := echoContext.Param("id")
	id, pe, err := getProtectedEntityForIDStr(this.petm, idStr, echoContext)
	if err != nil {
		return nil
	}

	if action, ok := echoContext.Request().URL.Query()["action"]; ok {
		switch action[0] {
		case "snapshot":
			this.snapshot(echoContext, pe)
		case "deleteSnapshot":
			this.deleteSnapshot(echoContext, pe)
		default:
			echoContext.String(http.StatusBadRequest, "Action "+action[0]+" not understood")
		}
		return nil
	}
	info, err := pe.GetInfo(context.Background())
	if err != nil {
		echoContext.String(http.StatusNotFound, "Could not retrieve info for id "+id.String()+" error = "+err.Error())
		return nil
	}
	echoContext.JSON(http.StatusOK, info)

	return nil
}

func (this *ServiceAPI) snapshot(echoContext echo.Context, pe astrolabe.ProtectedEntity) {
	snapshotID, err := pe.Snapshot(context.Background(), make(map[string]map[string]interface{}))
	if err != nil {
		echoContext.String(http.StatusNotFound, "Snapshot failed for id "+pe.GetID().String()+" error = "+err.Error())
		return
	}

	echoContext.String(http.StatusOK, snapshotID.String())
}

func (this *ServiceAPI) deleteSnapshot(echoContext echo.Context, pe astrolabe.ProtectedEntity) {
	snapshotID := pe.GetID().GetSnapshotID()
	if snapshotID.GetID() == "" {
		echoContext.String(http.StatusBadRequest, "No snapshot ID specified in id "+pe.GetID().String()+" for delete")
		return
	}
	deleted, err := pe.DeleteSnapshot(context.Background(), snapshotID)
	if err != nil {
		echoContext.String(http.StatusNotFound, "Snapshot delete failed for id "+pe.GetID().String()+" error = "+err.Error())
		return
	}
	if deleted == false {
		echoContext.String(http.StatusInternalServerError, "Could not delete snapshot "+pe.GetID().String())
		return
	}
	echoContext.String(http.StatusOK, snapshotID.String())
}

func (this *ServiceAPI) handleSnapshotListRequest(echoContext echo.Context) error {
	idStr := echoContext.Param("id")
	id, pe, err := getProtectedEntityForIDStr(this.petm, idStr, echoContext)
	if err != nil {
		return nil
	}
	snapshotIDs, err := pe.ListSnapshots(context.Background())
	if pe == nil {
		echoContext.String(http.StatusInternalServerError, "Could not retrieve snapshots "+id.String())
		return nil
	}
	snapshotIDStrs := []string{}
	for _, curSnapshotID := range snapshotIDs {
		snapshotIDStrs = append(snapshotIDStrs, curSnapshotID.String())
	}
	echoContext.JSON(http.StatusOK, snapshotIDStrs)
	return nil
}

func (this *ServiceAPI) handleCopyObject(echoContext echo.Context) (err error) {
	pei := new(astrolabe.ProtectedEntityInfoImpl)
	if err = echoContext.Bind(pei); err != nil {
		return
	}
	newPE, err := this.petm.CopyFromInfo(context.Background(), pei, astrolabe.AllocateNewObject)
	if err != nil {
		return err
	}
	echoContext.String(http.StatusOK, newPE.GetID().String())
	return
}
