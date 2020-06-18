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
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/fs"
	"github.com/vmware-tanzu/astrolabe/pkg/ivd"
	"github.com/vmware-tanzu/astrolabe/pkg/kubernetes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

type DirectProtectedEntityManager struct {
	typeManager map[string]astrolabe.ProtectedEntityTypeManager
	s3Config    astrolabe.S3Config
}

func NewDirectProtectedEntityManager(petms []astrolabe.ProtectedEntityTypeManager, s3Config astrolabe.S3Config) (returnPEM *DirectProtectedEntityManager) {
	returnPEM = &DirectProtectedEntityManager{
		typeManager: make(map[string]astrolabe.ProtectedEntityTypeManager),
	}
	for _, curPETM := range petms {
		returnPEM.typeManager[curPETM.GetTypeName()] = curPETM
	}
	returnPEM.s3Config = s3Config
	return
}

func NewDirectProtectedEntityManagerFromConfigDir(confDirPath string) *DirectProtectedEntityManager {
	configInfo, err := readConfigFiles(confDirPath)
	if err != nil {
		log.Fatalf("Could not read config files from dir %s, err: %v", confDirPath, err)
	}
	return NewDirectProtectedEntityManagerFromParamMap(configInfo)
}

func NewDirectProtectedEntityManagerFromParamMap(configInfo ConfigInfo) *DirectProtectedEntityManager {
	petms := make([]astrolabe.ProtectedEntityTypeManager, 0) // No guarantee all configs will be valid, so don't preallocate
	var err error
	logger := logrus.New()
	for serviceName, params := range configInfo.peConfigs {
		var curService astrolabe.ProtectedEntityTypeManager
		switch serviceName {
		case "ivd":
			curService, err = ivd.NewIVDProtectedEntityTypeManagerFromConfig(params, configInfo.s3Config, logger)
		case "k8sns":
			curService, err = kubernetes.NewKubernetesNamespaceProtectedEntityTypeManagerFromConfig(params, configInfo.s3Config,
				logger)
		case "fs":
			curService, err = fs.NewFSProtectedEntityTypeManagerFromConfig(params, configInfo.s3Config, logger)
		default:

		}
		if err != nil {
			log.Printf("Could not start service %s err=%v", serviceName, err)
			continue
		}
		if curService != nil {
			petms = append(petms, curService)
		}
	}
	return NewDirectProtectedEntityManager(petms, configInfo.s3Config)
}

type ConfigInfo struct {
	peConfigs map[string]map[string]interface{}
	s3Config  astrolabe.S3Config
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

	return ConfigInfo{
		peConfigs: configMap,
		s3Config:  *s3Config,
	}, nil
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
	return this.typeManager[id.GetPeType()].GetProtectedEntity(ctx, id)
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
