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
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/fs"
	"github.com/vmware-tanzu/astrolabe/pkg/pvc"
	"github.com/vmware-tanzu/astrolabe/pkg/util"
)

type DirectProtectedEntityManager struct {
	typeManager map[string]astrolabe.ProtectedEntityTypeManager
	s3Config    astrolabe.S3Config
	logger      logrus.FieldLogger
}

// NewDirectProtectedEntityManager - This will create a Protected Entity Manager with a set of configured
// Protected Entity Type Managers from petms.   s3Config will be used to create S3 transports for the PEs.
func NewDirectProtectedEntityManager(petms []astrolabe.ProtectedEntityTypeManager, s3Config astrolabe.S3Config, logger logrus.FieldLogger) (returnPEM *DirectProtectedEntityManager) {
	returnPEM = &DirectProtectedEntityManager{
		typeManager: make(map[string]astrolabe.ProtectedEntityTypeManager),
		logger:      logger,
	}
	for _, curPETM := range petms {
		returnPEM.typeManager[curPETM.GetTypeName()] = curPETM
		switch curPETM.(type) {
		case *pvc.PVCProtectedEntityTypeManager:
			curPETM.(*pvc.PVCProtectedEntityTypeManager).SetProtectedEntityManager(returnPEM)
		}
	}
	returnPEM.s3Config = s3Config
	return
}

// NewDirectProtectedEntityManagerFromConfigDir - This reads an Astrolabe configuration directory and creates and configures
// Protected Entity Type Managers with the information from the config directory confDirPath using the addOnInits map to
// find InitFuncs to create the PETMS as needed.
func NewDirectProtectedEntityManagerFromConfigDir(confDirPath string, addonInits map[string]astrolabe.InitFunc, logger logrus.FieldLogger) *DirectProtectedEntityManager {
	configInfo, err := util.ReadConfigFiles(confDirPath)
	if err != nil {
		logger.Fatalf("Could not read config files from dir %s, err: %v", confDirPath, err)
	}
	return NewDirectProtectedEntityManagerFromParamMap(configInfo, addonInits, logger)
}

// NewDirectProtectedEntityManagerFromParamMap - Creates and configures Protected Entity Type Managers with the
// configuration in configInfo using the addOnInits map to find InitFuncs to create the PETMS as needed.
func NewDirectProtectedEntityManagerFromParamMap(configInfo util.ConfigInfo, addonInits map[string]astrolabe.InitFunc, logger logrus.FieldLogger) *DirectProtectedEntityManager {
	initFuncs := make(map[string]astrolabe.InitFunc)
	initFuncs["fs"] = fs.NewFSProtectedEntityTypeManagerFromConfig
	initFuncs["pvc"] = pvc.NewPVCProtectedEntityTypeManagerFromConfig
	if addonInits != nil {
		for peType, initFunc := range addonInits {
			initFuncs[peType] = initFunc
		}
	}
	petms := make([]astrolabe.ProtectedEntityTypeManager, 0) // No guarantee all configs will be valid, so don't preallocate
	var err error
	if logger == nil {
		logger = logrus.New()
	}
	for serviceName, params := range configInfo.PEConfigs {
		var curService astrolabe.ProtectedEntityTypeManager
		initFunc := initFuncs[serviceName]
		if initFunc != nil {
			curService, err = initFunc(params, configInfo.S3Config, logger)
			if err != nil {
				logger.Infof("Could not start service %s err=%v", serviceName, err)
				continue
			}
			if curService != nil {
				petms = append(petms, curService)
			}
		} else {
			logger.Warnf("Unknown service type, %v", serviceName)
		}
	}
	return NewDirectProtectedEntityManager(petms, configInfo.S3Config, logger)
}

func (this *DirectProtectedEntityManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	typeManager, ok := this.typeManager[id.GetPeType()]
	if !ok {
		errMsg := fmt.Sprintf("PeType, %v, is not available", id.GetPeType())
		this.logger.Error(errMsg)
		return nil, errors.New(errMsg)
	}
	return typeManager.GetProtectedEntity(ctx, id)
}

func (this *DirectProtectedEntityManager) GetProtectedEntityTypeManager(peType string) astrolabe.ProtectedEntityTypeManager {
	return this.typeManager[peType]
}

func (this *DirectProtectedEntityManager) ListEntityTypeManagers() []astrolabe.ProtectedEntityTypeManager {
	returnArr := []astrolabe.ProtectedEntityTypeManager{}
	for _, curPETM := range this.typeManager {
		returnArr = append(returnArr, curPETM)
	}
	return returnArr
}

func (this *DirectProtectedEntityManager) RegisterExternalProtectedEntityTypeManagers(petms []astrolabe.ProtectedEntityTypeManager) {
	for _, curPETM := range petms {
		this.logger.Infof("Registered External ProtectedEntityTypeManager: %v", curPETM.GetTypeName())
		this.typeManager[curPETM.GetTypeName()] = curPETM
	}
}
