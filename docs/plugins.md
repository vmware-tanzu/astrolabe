# Hashicorp/GRPC Plugins

The plugin mechanism provides a way for Protected Entities to be loosely
coupled with clients and servers.  It is designed to replace static linking in
executables, not to provide a general remote access mechanism.  It uses the
[Hashicorp Go Plugin](https://github.com/hashicorp/go-plugin) plugin framework.

## Creating a plugin
To create a plugin, implement `ProtectedEntityTypeManager` and `ProtectedEntity`
for your Protected Entity type as Go classes.  The plugin will be packaged
as an executable with a single Protected Entity type per executable.  Create the
executable with a `main.go` similar to this:

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

The plugin executable should be named the Protected Entity Type, e.g. `fs` in the example above.

## Using plugins
The `PluginProtectedEntityManager` can be used to access plugins.  It will take
a set of Protected Entity configurations, either from a file or as in-memory structures
and a plugins directory and execute and configure the Protected Entities that have
configurations and plugin executables.

`NewPluginProtectedEntityManagerFromConfigDir` will read a [configuration directory](config_dir.md) and
a plugins directory and return a Protected Entity Manager that can be used to
access any of the types.

## Plugin lifecycle
Plugins are created and configured when a `PluginProtectedEntityManager` is configured.  
Plugin process are started and remain running for the life of the parent process.

## Cross-plugin communication
TBI - The parent process will provide a proxy PEM to all of the plugin processes that will
route requests back through the parent process.

