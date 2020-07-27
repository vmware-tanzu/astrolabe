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

package main

import (
	"flag"
	"fmt"
	"github.com/go-openapi/loads"
	"github.com/vmware-tanzu/astrolabe/gen/restapi"
	"github.com/vmware-tanzu/astrolabe/gen/restapi/operations"
	"github.com/vmware-tanzu/astrolabe/pkg/server"
	"os"
	"strconv"

	//"github.com/labstack/gommon/log"
	"log"
)

func main() {
	confDirStr := flag.String("confDir", "", "Configuration directory")
	apiPortStr := flag.String("apiPort", "1323", "REST API port")
	insecure := flag.Bool("insecure", false, "Only use HTTP")
	flag.Parse()
	if *confDirStr == "" {
		log.Println("confDir is not defined")
		flag.Usage()
		return
	}
	apiPort, err := strconv.Atoi(*apiPortStr)
	if err != nil {
		fmt.Errorf("apiPort %s is not an integer", *apiPortStr)
		os.Exit(1)
	}
	pem := server.NewProtectedEntityManager(*confDirStr)
	tm := server.NewTaskManager()
	apiHandler := server.NewOpenAPIAstrolabeHandler(pem, tm)
	// load embedded swagger file
	swaggerSpec, err := loads.Analyzed(restapi.SwaggerJSON, "")
	if err != nil {
		log.Fatalln(err)
	}

	// create new service API
	api := operations.NewAstrolabeAPI(swaggerSpec)
	server := restapi.NewServer(api)
	if *insecure {
		server.EnabledListeners = []string{"http"}
	}
	defer server.Shutdown()

	// parse flags
	flag.Parse()
	// set the port this service will be run on
	server.Port = apiPort

	apiHandler.AttachHandlers(api)

	// serve API
	if err := server.Serve(); err != nil {
		log.Fatalln(err)
	}
}
