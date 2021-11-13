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

package server

import (
	"flag"
	"github.com/go-openapi/loads"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/gen/restapi"
	"github.com/vmware-tanzu/astrolabe/gen/restapi/operations"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"github.com/vmware-tanzu/astrolabe/pkg/plugin"
	"strconv"
)

func ServerMain(addonInits map[string]astrolabe.InitFunc) {
	log := logrus.StandardLogger()
	server, _, err := ServerInit(addonInits)
	if err != nil {
		log.Errorf("Error initializing server = %v", err)
		return
	}
	defer server.Shutdown()

	// serve API
	if err := server.Serve(); err != nil {
		log.Errorf("Fatal error while serving %v", err)
	}
}

func ServerInit(addonInits map[string]astrolabe.InitFunc) (*restapi.Server, astrolabe.ProtectedEntityManager, error) {
	confDirStr := flag.String("confDir", "", "Configuration directory")
	apiPortStr := flag.String("apiPort", "1323", "REST API port")
	insecure := flag.Bool("insecure", false, "Only use HTTP")
	pluginsDirStr := flag.String("pluginsDir", "", "Plugin directory")

	flag.Parse()
	if *confDirStr == "" {
		flag.Usage()
		return nil, nil, errors.New("confDir is not defined")
	}
	apiPort, err := strconv.Atoi(*apiPortStr)
	if err != nil {
		return nil, nil, errors.Errorf("apiPort %s is not an integer\n", *apiPortStr)
	}
	var pem astrolabe.ProtectedEntityManager
	if *pluginsDirStr == "" {
		pem = NewProtectedEntityManager(*confDirStr, addonInits, logrus.New())
	} else {
		pem, err = plugin.NewPluginProtectedEntityManagerFromConfigDir(*confDirStr, *pluginsDirStr, logrus.New())
		if err != nil {
			return nil, nil, errors.WithMessage(err, "Could not create plugin protected entity manager")
		}
	}
	tm := NewTaskManager()
	apiHandler := NewOpenAPIAstrolabeHandler(pem, tm)
	// load embedded swagger file
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		return nil, nil, errors.WithMessage(err, "Could not load swagger spec")
	}

	// create new service API
	api := operations.NewAstrolabeAPI(swaggerSpec)
	server := restapi.NewServer(api)
	if *insecure {
		server.EnabledListeners = []string{"http"}
	}

	// parse flags
	flag.Parse()
	// set the port this service will be run on
	server.Port = apiPort

	apiHandler.AttachHandlers(api)
	return server, pem, nil
}
