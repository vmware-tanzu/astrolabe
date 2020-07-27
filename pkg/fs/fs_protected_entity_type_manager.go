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

package fs

import (
	"context"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"io/ioutil"
	"path/filepath"
)

type FSProtectedEntityTypeManager struct {
	root     string
	s3Config astrolabe.S3Config
	logger   logrus.FieldLogger
}

const kTYPE_NAME = "fs"

func NewFSProtectedEntityTypeManagerFromConfig(params map[string]interface{}, s3Config astrolabe.S3Config,
	logger logrus.FieldLogger) (*FSProtectedEntityTypeManager, error) {
	root := params["root"].(string)

	returnTypeManager := FSProtectedEntityTypeManager{
		root:     root,
		s3Config: s3Config,
		logger:   logger,
	}
	return &returnTypeManager, nil
}

func (this *FSProtectedEntityTypeManager) GetTypeName() string {
	return kTYPE_NAME
}

func (this *FSProtectedEntityTypeManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (
	astrolabe.ProtectedEntity, error) {
	return newFSProtectedEntity(this, id, id.GetID(), filepath.Join(this.root, id.GetID()))
}

func (this *FSProtectedEntityTypeManager) GetProtectedEntities(ctx context.Context) ([]astrolabe.ProtectedEntityID, error) {
	files, err := ioutil.ReadDir(this.root)
	if err != nil {
		return nil, err
	}

	var retVal = make([]astrolabe.ProtectedEntityID, len(files))
	for index, curFile := range files {
		peid := astrolabe.NewProtectedEntityID("fs", curFile.Name())
		retVal[index] = peid
	}
	return retVal, nil
}

func (this *FSProtectedEntityTypeManager) Copy(ctx context.Context, pe astrolabe.ProtectedEntity,
	options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {

	sourcePEInfo, err := pe.GetInfo(ctx)
	if err != nil {
		return nil, err
	}
	dataReader, err := pe.GetDataReader(nil)
	if dataReader != nil {
		defer func() {
			if err := dataReader.Close(); err != nil {
				this.logger.Errorf("The deferred data reader is closed with error, %v", err)
			}
		}()
	}

	if err != nil {
		return nil, err
	}

	metadataReader, err := pe.GetMetadataReader(nil)
	if err != nil {
		return nil, err
	}
	return this.copyInt(ctx, sourcePEInfo, options, dataReader, metadataReader)
}

func (this *FSProtectedEntityTypeManager) CopyFromInfo(ctx context.Context, pe astrolabe.ProtectedEntityInfo,
	options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	return nil, nil
}

func (this *FSProtectedEntityTypeManager) copyInt(ctx context.Context, sourcePEInfo astrolabe.ProtectedEntityInfo,
	options astrolabe.CopyCreateOptions, dataReader io.Reader, metadataReader io.Reader) (astrolabe.ProtectedEntity, error) {
	id := sourcePEInfo.GetID()
	if id.GetPeType() != kTYPE_NAME {
		return nil, errors.New(id.GetPeType() + " is not of type fs")
	}
	if options == astrolabe.AllocateObjectWithID {
		return nil, errors.New("AllocateObjectWithID not supported")
	}

	if options == astrolabe.UpdateExistingObject {
		return nil, errors.New("UpdateExistingObject not supported")
	}

	fsUUID, err := uuid.NewRandom()
	if err != nil {
		panic("uuid.NewRandom return err ")
	}
	newPEID := astrolabe.NewProtectedEntityID(kTYPE_NAME, fsUUID.String())
	newPE, err := newFSProtectedEntity(this, newPEID, sourcePEInfo.GetName(), filepath.Join(this.root, newPEID.GetID()))
	if err != nil {
		return nil, err
	}
	err = newPE.createDir()
	if err != nil {
		return nil, err
	}
	err = newPE.copy(ctx, dataReader, metadataReader)
	if err != nil {
		return nil, err
	}
	return newPE, nil
}

func (this *FSProtectedEntityTypeManager) getDataTransports(id astrolabe.ProtectedEntityID) ([]astrolabe.DataTransport,
	[]astrolabe.DataTransport,
	[]astrolabe.DataTransport, error) {

	dataS3Transport, err := astrolabe.NewS3DataTransportForPEID(id, this.s3Config)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "Could not create S3 data transport")
	}

	data := []astrolabe.DataTransport{
		dataS3Transport,
	}

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
