/*
 * Copyright the Astrolabe contributors
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
package plugin

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/plugin/framework"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
	"github.com/vmware-tanzu/astrolabe/pkg/util"
	"os"
	"os/exec"
	"path/filepath"
)

type pluginProtectedEntityManager struct {
	logger logrus.FieldLogger
	pluginsDir string
	pluginPETMs map[string]*pluginPETM
}

type pluginPETM struct {
	client *plugin.Client
	petm framework.ProtectedEntityTypeManagerClient
}

func NewPluginProtectedEntityManagerFromConfigDir(confDirPath string, pluginsDir string, logger logrus.FieldLogger) (astrolabe.ProtectedEntityManager, error) {
	configInfo, err := util.ReadConfigFiles(confDirPath)
	if err != nil {
		logger.Fatalf("Could not read config files from dir %s, err: %v", confDirPath, err)
	}
	return NewPluginProtectedEntityManagerFromParamMap(configInfo, pluginsDir, logger)
}

func NewPluginProtectedEntityManagerFromParamMap(configInfo util.ConfigInfo, pluginsDir string, logger logrus.FieldLogger) (astrolabe.ProtectedEntityManager, error) {
	pluginsDirInfo, err := os.Stat(pluginsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not stat pluginsDir %s", pluginsDir)
	}

	if !pluginsDirInfo.IsDir() {
		return nil, errors.Errorf("pluginsDir %s is not a directory", pluginsDir)
	}

	pluginsDirFile, err := os.Open(pluginsDir)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not open pluginsDir %s", pluginsDir)
	}

	fileNames, err := pluginsDirFile.Readdirnames(0)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not list plugindDir %s", pluginsDir)
	}

	pluginPETMs := map[string]*pluginPETM{}

	for _, curFile := range fileNames {
		pluginPath := filepath.Join(pluginsDir, curFile)
		pluginLogger := logger.WithField("plugin", pluginPath)
		pluginPETM, err := startPluginPETM(pluginPath)
		if err != nil {
			pluginLogger.WithError(err).Error("Could not start plugin")
			continue
		}
		peName := filepath.Base(curFile)
		params := configInfo.PEConfigs[peName]
		if params != nil {
			err = pluginPETM.petm.Init(params, configInfo.S3Config)
			if err != nil {
				pluginLogger.WithError(err).Error("Could not initialize plugin")
				continue
			}
		}
		pluginPETMs[peName] = pluginPETM
	}
	ppem := pluginProtectedEntityManager{
		logger:     logger,
		pluginsDir: pluginsDir,
		pluginPETMs: pluginPETMs,
	}
	return &ppem, nil
}

const (
	PetmPluginName = "petm"
)
// PluginMap is the map of plugins we can dispense.
var PluginMap = map[string]plugin.Plugin{
	PetmPluginName: &framework.ProtectedEntityTypeManagerPlugin{},
}

func startPluginPETM(pluginExecutable string) (*pluginPETM, error) {
	pluginExecutableInfo, err := os.Stat(pluginExecutable)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not stat plugin %s", pluginExecutable)
	}
	if pluginExecutableInfo.IsDir() {
		return nil, errors.Errorf("pluginsDir %s is not a file", pluginExecutable)
	}

	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: framework.Handshake,
		Plugins:         PluginMap,
		Cmd:             exec.Command("sh", "-c", pluginExecutable),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
	})

	// Connect via GRPC
	rpcClient, err := client.Client()
	if err != nil {
		return nil, errors.Wrapf(err, "Could not start client for plugin executable %s", pluginExecutable)
	}

	// Request the plugin
	raw, err := rpcClient.Dispense(PetmPluginName)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get PETM for plugin executable %s", pluginExecutable)
	}

	petm := framework.NewProtectedEntityTypeManagerClient(raw.(generated.ProtectedEntityTypeManagerClient))

	return &pluginPETM{
		client: client,
		petm:   petm,
	}, nil
}

func (recv * pluginProtectedEntityManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	petm := recv.GetProtectedEntityTypeManager(id.GetPeType())
	if petm == nil {
			errMsg := fmt.Sprintf("PeType, %v, is not available", id.GetPeType())
			recv.logger.Error(errMsg)
			return nil, errors.New(errMsg)
	}
	return petm.GetProtectedEntity(ctx, id)
}

func (recv * pluginProtectedEntityManager) GetProtectedEntityTypeManager(peType string) astrolabe.ProtectedEntityTypeManager {
	returnPluginPETM := recv.pluginPETMs[peType]
	if returnPluginPETM != nil {
		return returnPluginPETM.petm
	} else {
		return nil
	}
}

func (recv * pluginProtectedEntityManager) ListEntityTypeManagers() []astrolabe.ProtectedEntityTypeManager {
	returnPETMs := []astrolabe.ProtectedEntityTypeManager{}
	for _, petm := range recv.pluginPETMs {
		returnPETMs = append(returnPETMs, petm.petm)
	}
	return returnPETMs
}
