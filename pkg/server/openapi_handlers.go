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
	"github.com/go-openapi/runtime/middleware"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/gen/models"
	"github.com/vmware-tanzu/astrolabe/gen/restapi/operations"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
)

type OpenAPIAstrolabeHandler struct {
	pem astrolabe.ProtectedEntityManager
	tm  *TaskManager
	logger logrus.FieldLogger
}

func NewOpenAPIAstrolabeHandler(pem astrolabe.ProtectedEntityManager, tm *TaskManager) OpenAPIAstrolabeHandler {
	return OpenAPIAstrolabeHandler{
		pem: pem,
		tm:  tm,
		logger: logrus.StandardLogger(),
	}
}
func (this OpenAPIAstrolabeHandler) AttachHandlers(api *operations.AstrolabeAPI) {
	api.ListServicesHandler = operations.ListServicesHandlerFunc(this.ListServices)
	api.ListProtectedEntitiesHandler = operations.ListProtectedEntitiesHandlerFunc(this.ListProtectedEntities)
	api.GetProtectedEntityInfoHandler = operations.GetProtectedEntityInfoHandlerFunc(this.GetProtectedEntityInfo)
	api.CreateSnapshotHandler = operations.CreateSnapshotHandlerFunc(this.CreateSnapshot)
	api.ListSnapshotsHandler = operations.ListSnapshotsHandlerFunc(this.ListSnapshots)
	api.CopyProtectedEntityHandler = operations.CopyProtectedEntityHandlerFunc(this.CopyProtectedEntity)
}

func (this OpenAPIAstrolabeHandler) ListServices(params operations.ListServicesParams) middleware.Responder {
	etms := this.pem.ListEntityTypeManagers()
	var serviceList = models.ServiceList{
		Services: make([]string, len(etms)),
	}
	for curETMNum, curETM := range etms {
		serviceList.Services[curETMNum] = curETM.GetTypeName()
	}
	return operations.NewListServicesOK().WithPayload(&serviceList)
}

func (this OpenAPIAstrolabeHandler) ListProtectedEntities(params operations.ListProtectedEntitiesParams) middleware.Responder {
	petm := this.pem.GetProtectedEntityTypeManager(params.Service)
	if petm == nil {
		return operations.NewListProtectedEntitiesNotFound()
	}
	peids, err := petm.GetProtectedEntities(context.Background())
	if err != nil {
		errStr := err.Error()
		return operations.NewListProtectedEntitiesInternalServerError().WithPayload(&models.ServerError{
			Error:            models.Error{
				Message: &errStr,
			},
		})
	}
	mpeids := make([]models.ProtectedEntityID, len(peids))
	for peidNum, peid := range peids {
		mpeids[peidNum] = models.ProtectedEntityID(peid.String())
	}
	peList := models.ProtectedEntityList{
		List:      mpeids,
		Truncated: false,
	}
	return operations.NewListProtectedEntitiesOK().WithPayload(&peList)
}

func (this OpenAPIAstrolabeHandler) GetProtectedEntityInfo(params operations.GetProtectedEntityInfoParams) middleware.Responder {
	this.logger.WithField("service", params.Service)
	this.logger.WithField("peid", params.ProtectedEntityID)
	petm := this.pem.GetProtectedEntityTypeManager(params.Service)
	if petm == nil {
		this.logger.Error("Could not find Protected Entity Type Manager")
		return operations.NewGetProtectedEntityInfoNotFound()
	}
	peid, err := astrolabe.NewProtectedEntityIDFromString(params.ProtectedEntityID)
	if err != nil {
		this.logger.Error("Could not find Protected Entity ID")
		return operations.NewGetProtectedEntityInfoNotFound()
	}
	pe, err := petm.GetProtectedEntity(context.Background(), peid)
	if err != nil {
		this.logger.Errorf("Could not retrieve Protected Entity, err:%v", err)
		return operations.NewGetProtectedEntityInfoNotFound()
	}
	peInfo, err := pe.GetInfo(context.Background())
	if err != nil {
		this.logger.Errorf("Could not retrieve Protected Entity, err:%v", err)
		return operations.NewGetProtectedEntityInfoNotFound()
	}
	peInfoResponse := peInfo.GetModelProtectedEntityInfo()
	return operations.NewGetProtectedEntityInfoOK().WithPayload(&peInfoResponse)
}

func (this OpenAPIAstrolabeHandler) CreateSnapshot(params operations.CreateSnapshotParams) middleware.Responder {
	petm := this.pem.GetProtectedEntityTypeManager(params.Service)
	if petm == nil {

	}
	peid, err := astrolabe.NewProtectedEntityIDFromString(params.ProtectedEntityID)
	if err != nil {

	}

	snapshotParams := make(map[string]map[string]interface{})

	if params.Params != nil {
		for _, curPEParams := range params.Params {
			if curPEParams.Value != nil {
				curPEParamsMap := make(map[string]interface{})
				for _, curParam := range curPEParams.Value {
					curPEParamsMap[curParam.Key] = curParam.Value
				}
				snapshotParams[curPEParams.Key] = curPEParamsMap
			}
		}
	}
	pe, err := petm.GetProtectedEntity(context.Background(), peid)
	if err != nil {

	}
	snapshotID, err := pe.Snapshot(context.Background(), snapshotParams)
	if err != nil {

	}

	return operations.NewCreateSnapshotOK().WithPayload(snapshotID.GetModelProtectedEntitySnapshotID())
}

func (this OpenAPIAstrolabeHandler) ListSnapshots(params operations.ListSnapshotsParams) middleware.Responder {
	petm := this.pem.GetProtectedEntityTypeManager(params.Service)
	if petm == nil {

	}
	peid, err := astrolabe.NewProtectedEntityIDFromString(params.ProtectedEntityID)
	if err != nil {

	}
	pe, err := petm.GetProtectedEntity(context.Background(), peid)
	if err != nil {

	}
	snapshots, err := pe.ListSnapshots(context.Background())
	mpeids := make([]models.ProtectedEntityID, len(snapshots))
	for snidNum, snid := range snapshots {
		mpeids[snidNum] = models.ProtectedEntityID(peid.IDWithSnapshot(snid).String())
	}

	peList := models.ProtectedEntityList{
		List:      mpeids,
		Truncated: false,
	}
	return operations.NewListSnapshotsOK().WithPayload(&peList)

}

func (this OpenAPIAstrolabeHandler) CopyProtectedEntity(params operations.CopyProtectedEntityParams) middleware.Responder {
	petm := this.pem.GetProtectedEntityTypeManager(params.Service)
	if petm == nil {

	}
	pei, err := astrolabe.NewProtectedEntityInfoFromModel(params.Body.ProtectedEntityInfo)
	if err != nil {

	}
	copyParams := make(map[string]map[string]interface{})

	if params.Body.CopyParams != nil {
		for _, curPEParams := range params.Body.CopyParams {
			if curPEParams.Value != nil {
				curPEParamsMap := make(map[string]interface{})
				for _, curParam := range curPEParams.Value {
					curPEParamsMap[curParam.Key] = curParam.Value
				}
				copyParams[curPEParams.Key] = curPEParamsMap
			}
		}
	}
	//startedTime := time.Now()
	newPE, err := petm.CopyFromInfo(context.Background(), pei, copyParams, astrolabe.AllocateNewObject)
	if err != nil {
		// TODO - create an error path
		//return NewCopyProtectedEntity
	}
	/*
	var taskStatus astrolabe.TaskStatus
	if err != nil {
		taskStatus = astrolabe.Failed
	} else {
		taskStatus = astrolabe.Success
	}
	// Fake a task for now
	task := astrolabe.NewGenericTask()
	task.Completed = true
	task.StartedTime = startedTime
	task.FinishedTime = time.Now()
	task.Progress = 100
	task.TaskStatus = taskStatus
	task.Result = newPE.GetID().GetModelProtectedEntityID()
	 */
	/*
	For now we are not actually implementing tasks.  Instead we are abusing the TaskID field and returning the ID of the PE
	that was created (if any)
	 */
	ncpea := operations.NewCopyProtectedEntityAccepted()
	payload := models.CreateInProgressResponse{
		TaskID: models.TaskID(newPE.GetID().String()),
	}
	ncpea.Payload = &payload
	return ncpea
}
