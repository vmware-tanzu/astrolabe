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