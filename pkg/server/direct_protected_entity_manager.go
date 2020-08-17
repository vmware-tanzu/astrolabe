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
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/fs"
	"github.com/vmware-tanzu/astrolabe/pkg/ivd"
	"github.com/vmware-tanzu/astrolabe/pkg/kubernetes"
	"github.com/vmware-tanzu/astrolabe/pkg/pvc"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type DirectProtectedEntityManager struct {
	typeManager map[string]astrolabe.ProtectedEntityTypeManager
	s3Config    astrolabe.S3Config
	logger      logrus.FieldLogger
}

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

func NewDirectProtectedEntityManagerFromConfigDir(confDirPath string) *DirectProtectedEntityManager {
	configInfo, err := readConfigFiles(confDirPath)
	if err != nil {
		log.Fatalf("Could not read config files from dir %s, err: %v", confDirPath, err)
	}
	return NewDirectProtectedEntityManagerFromParamMap(configInfo, nil)
}

func NewDirectProtectedEntityManagerFromParamMap(configInfo ConfigInfo, logger logrus.FieldLogger) *DirectProtectedEntityManager {
	petms := make([]astrolabe.ProtectedEntityTypeManager, 0) // No guarantee all configs will be valid, so don't preallocate
	var err error
	if logger == nil {
		logger = logrus.New()
	}
	for serviceName, params := range configInfo.PEConfigs {
		var curService astrolabe.ProtectedEntityTypeManager
		switch serviceName {
		case "ivd":
			curService, err = ivd.NewIVDProtectedEntityTypeManagerFromConfig(params, configInfo.S3Config, logger)
		case "k8sns":
			curService, err = kubernetes.NewKubernetesNamespaceProtectedEntityTypeManagerFromConfig(params, configInfo.S3Config,
				logger)
		case "fs":
			curService, err = fs.NewFSProtectedEntityTypeManagerFromConfig(params, configInfo.S3Config, logger)
		case "pvc":
			curService, err = pvc.NewPVCProtectedEntityTypeManagerFromConfig(params, configInfo.S3Config, logger)
		default:
			logger.Warnf("Unknown service type, %v", serviceName)
		}
		if err != nil {
			logger.Infof("Could not start service %s err=%v", serviceName, err)
			continue
		}
		if curService != nil {
			petms = append(petms, curService)
		}
	}
	return NewDirectProtectedEntityManager(petms, configInfo.S3Config, logger)
}

type ConfigInfo struct {
	PEConfigs map[string]map[string]interface{}
	S3Config  astrolabe.S3Config
}

func NewConfigInfo(peConfigs map[string]map[string]interface{}, s3Config astrolabe.S3Config) ConfigInfo {
	return ConfigInfo{
		PEConfigs: peConfigs,
		S3Config:  s3Config,
	}
}

func readConfigFiles(confDirPath string) (ConfigInfo, error) {
	configMap := make(map[string]map[string]interface{})

	confDir, err := os.Stat(confDirPath)
	if err != nil {
		log.Panicln("Could not stat configuration directory " + confDirPath)
	}
	if !confDir.Mode().IsDir() {
		log.Panicln(confDirPath + " is not a directory")

	}

	peDirPath := confDirPath + "/pes"
	peDir, err := os.Stat(peDirPath)
	if err != nil {
		log.Panicf("Could not stat protected entity configuration directory %s, err = %v\n", peDirPath, err)
	}
	if !peDir.Mode().IsDir() {
		log.Panicf("%s is not a directory", peDirPath)

	}
	files, err := ioutil.ReadDir(peDirPath)
	if err != nil {
		log.Panicf("Could not list protected entity configuration directory %s, err = %v\n", peDirPath, err)
	}
	for _, curFile := range files {
		if !strings.HasPrefix(curFile.Name(), ".") && strings.HasSuffix(curFile.Name(), fileSuffix) {
			peTypeName := strings.TrimSuffix(curFile.Name(), fileSuffix)
			peConf, err := readConfigFile(filepath.Join(peDirPath, curFile.Name()))
			if err != nil {
				log.Panicln("Could not process conf file " + curFile.Name() + " continuing, err = " + err.Error())
			} else {
				configMap[peTypeName] = peConf
			}
		}
	}

	s3ConfFilePath := filepath.Join(confDirPath, "s3config.json")

	s3Config, err := readS3ConfigFile(s3ConfFilePath)
	if err != nil {
		log.Panicf("Could not read S3 configuration directory %s, err = %v\n", s3ConfFilePath, err)
	}

	return NewConfigInfo(configMap, *s3Config), nil
}

func readConfigFile(confFile string) (map[string]interface{}, error) {
	jsonFile, err := os.Open(confFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open conf file %s", confFile)
	}
	defer jsonFile.Close()
	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read conf file %s", confFile)
	}
	var result map[string]interface{}
	err = json.Unmarshal([]byte(jsonBytes), &result)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to unmarshal JSON from %s", confFile)
	}
	return result, nil
}

func readS3ConfigFile(s3ConfFile string) (*astrolabe.S3Config, error) {
	jsonFile, err := os.Open(s3ConfFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open conf file %s", s3ConfFile)
	}
	defer jsonFile.Close()
	jsonBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not read conf file %s", s3ConfFile)
	}
	var result astrolabe.S3Config
	err = json.Unmarshal([]byte(jsonBytes), &result)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to unmarshal JSON from %s", s3ConfFile)
	}
	return &result, nil
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