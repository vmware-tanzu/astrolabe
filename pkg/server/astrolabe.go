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
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"
)

type Astrolabe struct {
	petm         *DirectProtectedEntityManager
	api_services map[string]*ServiceAPI
	s3_services  map[string]*ServiceS3
	s3URLBase    string
}

func NewAstrolabe(confDirPath string, port int) *Astrolabe {

	s3URLBase, pem := NewProtectedEntityManager(confDirPath, port)
	api_services := make(map[string]*ServiceAPI)
	s3_services := make(map[string]*ServiceS3)
	for _, curService := range pem.ListEntityTypeManagers() {
		serviceName := curService.GetTypeName()
		api_services[serviceName] = NewServiceAPI(curService)
		s3_services[serviceName] = NewServiceS3(curService)
	}

	retAstrolabe := Astrolabe{
		api_services: api_services,
		s3_services:  s3_services,
		s3URLBase:    s3URLBase,
	}

	return &retAstrolabe
}

func NewProtectedEntityManager(confDirPath string, port int) (string, astrolabe.ProtectedEntityManager) {
	s3URLBase, err := configS3URL(port)
	if err != nil {
		log.Fatal("Could not get host IP address", err)
	}
	dpem := NewDirectProtectedEntityManagerFromConfigDir(confDirPath, s3URLBase)
	var pem astrolabe.ProtectedEntityManager
	pem = dpem
	return s3URLBase, pem
}

func NewAstrolabeRepository() *Astrolabe {
	return nil
}

func (this *Astrolabe) Get(c echo.Context) error {
	var servicesList strings.Builder
	needsComma := false
	for serviceName := range this.api_services {
		if needsComma {
			servicesList.WriteString(",")
		}
		servicesList.WriteString(serviceName)
		needsComma = true
	}
	return c.String(http.StatusOK, servicesList.String())
}

func (this *Astrolabe) ConnectAstrolabeAPIToEcho(echo *echo.Echo) error {
	echo.GET("/astrolabe", this.Get)

	for serviceName, service := range this.api_services {
		echo.GET("/astrolabe/"+serviceName, service.listObjects)
		echo.POST("/astrolabe/"+serviceName, service.handleCopyObject)
		echo.GET("/astrolabe/"+serviceName+"/:id", service.handleObjectRequest)
		echo.GET("/astrolabe/"+serviceName+"/:id/snapshots", service.handleSnapshotListRequest)

	}
	return nil
}

func configS3URL(port int) (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return "http://" + ipnet.IP.String() + ":" + strconv.Itoa(port) + "/s3/", nil
			}
		}
	}
	return "", nil
}

func (this *Astrolabe) ConnectMiniS3ToEcho(echo *echo.Echo) error {
	echo.GET("/s3", this.Get)

	for serviceName, service := range this.s3_services {
		echo.GET("/s3/"+serviceName, service.listObjects)
		echo.GET("/s3/"+serviceName+"/:objectKey", service.handleObjectRequest)
	}
	return nil
}

const fileSuffix = ".pe.json"

func getProtectedEntityForIDStr(petm astrolabe.ProtectedEntityTypeManager, idStr string,
	echoContext echo.Context) (astrolabe.ProtectedEntityID, astrolabe.ProtectedEntity, error) {
	var id astrolabe.ProtectedEntityID
	var pe astrolabe.ProtectedEntity
	var err error

	id, err = astrolabe.NewProtectedEntityIDFromString(idStr)
	if err != nil {
		echoContext.String(http.StatusBadRequest, "id = "+idStr+" is invalid "+err.Error())
		return id, pe, err
	}
	if id.GetPeType() != (petm).GetTypeName() {
		echoContext.String(http.StatusBadRequest, "id = "+idStr+" is not type "+petm.GetTypeName())
		return id, pe, err
	}
	pe, err = (petm).GetProtectedEntity(context.Background(), id)
	if err != nil {
		echoContext.String(http.StatusNotFound, "Could not retrieve id "+id.String()+" error = "+err.Error())
		return id, pe, err
	}
	if pe == nil {
		echoContext.String(http.StatusInternalServerError, "pe was nil for "+id.String())
		return id, pe, err
	}
	return id, pe, nil
}
