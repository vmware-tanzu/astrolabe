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
	"io"
	"net/http"
	"strings"
)

type ServiceS3 struct {
	petm astrolabe.ProtectedEntityTypeManager
}

func NewServiceS3(petm astrolabe.ProtectedEntityTypeManager) *ServiceS3 {
	return &ServiceS3{
		petm: petm,
	}
}

func (this *ServiceS3) listObjects(echoContext echo.Context) error {
	/*
	 * No, this is not a correct implementation of S3 list bucket.
	 * TODO - write a proper implementation
	 */
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

func (this *ServiceS3) handleObjectRequest(echoContext echo.Context) error {
	objectKey := echoContext.Param("objectKey")
	var objectStream io.Reader
	var idStr, source, contentType string
	if strings.HasSuffix(objectKey, ".md") {
		idStr = strings.TrimSuffix(objectKey, ".md")
		source = "md"
		contentType = "application/octet-stream"
	} else {
		idStr = objectKey
		source = "data"
		contentType = "application/octet-stream"
	}

	_, pe, err := getProtectedEntityForIDStr(this.petm, idStr, echoContext)
	if err != nil {

	}

	switch source {
	case "md":
		objectStream, err = pe.GetMetadataReader(nil)
	case "data":
		objectStream, err = pe.GetDataReader(nil)
	}
	if err != nil {

	}

	echoContext.Stream(http.StatusOK, contentType, objectStream)
	return nil
}
