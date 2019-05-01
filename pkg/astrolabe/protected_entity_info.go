/*
 * Copyright 2019 VMware, Inc..
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
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/vmware-tanzu/astrolabe/gen/models"
	"net/url"
)

type ProtectedEntityInfo interface {
	GetID() ProtectedEntityID
	GetName() string
	GetDataTransports() [] DataTransport
	GetMetadataTransports() [] DataTransport
	GetCombinedTransports() [] DataTransport
	GetComponentIDs() []ProtectedEntityID
	GetModelProtectedEntityInfo() models.ProtectedEntityInfo
}

type ProtectedEntityInfoImpl struct {
	id                 ProtectedEntityID
	name               string
	dataTransports     []DataTransport
	metadataTransports []DataTransport
	combinedTransports []DataTransport
	componentIDs       []ProtectedEntityID
}

func NewProtectedEntityInfo(id ProtectedEntityID, name string, dataTransports []DataTransport, metadataTransports []DataTransport,
	combinedTransports []DataTransport, componentIDs []ProtectedEntityID) ProtectedEntityInfo {
	return ProtectedEntityInfoImpl{
		id:                 id,
		name:               name,
		dataTransports:     dataTransports,
		metadataTransports: metadataTransports,
		combinedTransports: combinedTransports,
		componentIDs:       componentIDs,
	}
}

func NewProtectedEntityInfoFromModel(mpei models.ProtectedEntityInfo) (ProtectedEntityInfo, error) {
	pei := ProtectedEntityInfoImpl{
	}
	err := pei.FillFromModel(mpei)
	if err != nil {
		return nil, err
	}
	return pei, nil
}

func (this ProtectedEntityInfoImpl) GetID() ProtectedEntityID {
	return this.id
}

func (this ProtectedEntityInfoImpl) GetName() string {
	return this.name
}

func stringsToURLs(urlStrs []string) ([]url.URL, error) {
	retList := []url.URL{}
	for _, curURLStr := range urlStrs {
		curURL, err := url.Parse(curURLStr)
		if err != nil {
			return nil, err
		}
		retList = append(retList, *curURL)
	}
	return retList, nil
}

func urlsToStrings(urls []url.URL) []string {
	retList := []string{}
	for _, curURL := range urls {
		curURLStr := curURL.String()
		retList = append(retList, curURLStr)
	}
	return retList
}

func (this ProtectedEntityInfoImpl) MarshalJSON() ([]byte, error) {
	jsonStruct := this.GetModelProtectedEntityInfo()

	return json.Marshal(jsonStruct)
}

func (this ProtectedEntityInfoImpl) GetModelProtectedEntityInfo() models.ProtectedEntityInfo {
	componentSpecs := make([]*models.ComponentSpec, len(this.componentIDs))
	for curComponentNum, curComponentID := range this.componentIDs {
		componentSpecs[curComponentNum] = &models.ComponentSpec{
			ID:     models.ProtectedEntityID(curComponentID.String()),
			Server: "", // TODO - convert to component specs throughout
		}
	}
	jsonStruct := models.ProtectedEntityInfo{
		ID:                 models.ProtectedEntityID(this.id.String()),
		Name:               &this.name,
		DataTransports:     convertToModelTransports(this.dataTransports),
		MetadataTransports: convertToModelTransports(this.metadataTransports),
		CombinedTransports: convertToModelTransports(this.combinedTransports),
		ComponentSpecs:     componentSpecs,
	}
	return jsonStruct
}

func convertToModelTransports(transports []DataTransport) []*models.DataTransport {
	retTransports := make([]*models.DataTransport, len(transports))
	for transportNum, curTransport := range transports {
		curModelDataTransport := curTransport.getModelDataTransport()
		retTransports[transportNum] = &curModelDataTransport
	}
	return retTransports
}

func convertToTransports(transports []*models.DataTransport) []DataTransport {
	retTransports := make([]DataTransport, len(transports))
	for transportNum, curTransport := range transports {
		retTransports[transportNum] = newDataTransportForModelTransport(*curTransport)
	}
	return retTransports
}

func appendJSON(buffer *bytes.Buffer, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return err
	}
	buffer.WriteString(fmt.Sprintf("\"%s\":%s", key, string(jsonValue)))
	return nil
}

func (this *ProtectedEntityInfoImpl) UnmarshalJSON(data []byte) error {
	jsonStruct := models.ProtectedEntityInfo{}
	err := json.Unmarshal(data, &jsonStruct)
	if err != nil {
		return err
	}
	err = this.FillFromModel(jsonStruct)
	if err != nil {
		return err
	}
	return nil
}

func (this *ProtectedEntityInfoImpl) FillFromModel(jsonStruct models.ProtectedEntityInfo) error {
	var err error
	this.id, err = NewProtectedEntityIDFromString(string(jsonStruct.ID))
	if err != nil {
		return err
	}
	this.name = *jsonStruct.Name
	this.dataTransports = convertToTransports(jsonStruct.DataTransports)
	this.metadataTransports = convertToTransports(jsonStruct.MetadataTransports)
	this.combinedTransports = convertToTransports(jsonStruct.CombinedTransports)
	componentIDs := make([]ProtectedEntityID, len(jsonStruct.ComponentSpecs))
	for curComponentNum, curComponentSpec := range componentIDs {
		componentID, err := NewProtectedEntityIDFromString(curComponentSpec.id)
		if err != nil {
			return err
		}
		componentIDs[curComponentNum] = componentID
	}
	this.componentIDs = componentIDs
	return nil
}

func (this ProtectedEntityInfoImpl) GetDataTransports() []DataTransport {
	return this.dataTransports
}

func (this ProtectedEntityInfoImpl) GetMetadataTransports() []DataTransport {
	return this.metadataTransports
}

func (this ProtectedEntityInfoImpl) GetCombinedTransports() []DataTransport {
	return this.dataTransports
}

func (this ProtectedEntityInfoImpl) GetComponentIDs() []ProtectedEntityID {
	return this.componentIDs
}
