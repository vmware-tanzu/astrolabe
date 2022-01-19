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
package framework

import (
	"context"
	"github.com/hashicorp/go-plugin"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
	"google.golang.org/grpc"
)

type ProtectedEntityTypeManagerPlugin struct {
	plugin.NetRPCUnsupportedPlugin // Embed NetRPCUnsupportedPlugin to implement plugin.Plugin and turn off net/rpc plugins.
	// We implement GRPCPlugin interface instead
	InitFunc astrolabe.InitFunc // The init func to create a new Protected Entity Type Manager - only used for the server
}

func (recv ProtectedEntityTypeManagerPlugin) GRPCServer(broker *plugin.GRPCBroker, server *grpc.Server) error {
	impl, err := NewProtectedEntityTypeManagerServer(recv.InitFunc)
	if err != nil {
		return errors.Wrap(err, "Could not create ProtectedEntityTypeManagerServer for GRPCServer")
	}
	generated.RegisterProtectedEntityTypeManagerServer(server, impl)
	return nil
}

func (recv ProtectedEntityTypeManagerPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return generated.NewProtectedEntityTypeManagerClient(c), nil
}
