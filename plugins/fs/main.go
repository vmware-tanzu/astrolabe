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
package main

import (
	"github.com/hashicorp/go-plugin"
	"github.com/vmware-tanzu/astrolabe/pkg/fs"
	astrolabe_plugin "github.com/vmware-tanzu/astrolabe/pkg/plugin"
	"github.com/vmware-tanzu/astrolabe/pkg/plugin/framework"
)

func main() {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: framework.Handshake,
		Plugins: map[string]plugin.Plugin{
			astrolabe_plugin.PetmPluginName: &framework.ProtectedEntityTypeManagerPlugin{InitFunc: fs.NewFSProtectedEntityTypeManagerFromConfig},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}