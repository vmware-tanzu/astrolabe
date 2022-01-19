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
