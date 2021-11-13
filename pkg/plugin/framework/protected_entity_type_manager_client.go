package framework

import (
	json "github.com/json-iterator/go"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	generated "github.com/vmware-tanzu/astrolabe/pkg/plugin/generated/v1"
	"golang.org/x/net/context"
)

type ProtectedEntityTypeManagerClient interface {
	Init(params map[string]interface{},	s3Config astrolabe.S3Config) error
	astrolabe.ProtectedEntityTypeManager
}
type protectedEntityTypeManagerClient struct {
	client generated.ProtectedEntityTypeManagerClient
}

func NewProtectedEntityTypeManagerClient(client generated.ProtectedEntityTypeManagerClient) ProtectedEntityTypeManagerClient {
	return &protectedEntityTypeManagerClient{client: client}
}

func (recv *protectedEntityTypeManagerClient) Init(params map[string]interface{},
	s3Config astrolabe.S3Config) error {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return errors.Wrap(err, "could not marshal params")
	}
	s3ConfigJSON, err := json.Marshal(s3Config)
	if err != nil {
		return errors.Wrap(err, "could not marshal s3Config")
	}
	ir := generated.InitRequest{
		ConfigInfo: string(paramsJSON),
		S3Config:   string(s3ConfigJSON),
	}
	_, err = recv.client.Init(context.Background(), &ir)
	return err
}

func (recv *protectedEntityTypeManagerClient) GetTypeName() string {
	resp, err := recv.client.GetTypeName(context.Background(), &generated.Empty{})
	if err != nil {
		return ""
	}
	return resp.TypeName
}

func (recv *protectedEntityTypeManagerClient) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	// In general PETMs don't check for existence when GetProtectedEntity is called so we don't make a call either
	return &protectedEntityClient{
		id:     id,
		client: recv.client,
		petm: recv,
	}, nil
}

func (recv *protectedEntityTypeManagerClient) GetProtectedEntities(ctx context.Context) ([]astrolabe.ProtectedEntityID, error) {
	resp, err := recv.client.GetProtectedEntities(context.Background(), &generated.Empty{})
	if err != nil {
		return nil, err
	}

	peIDs := make([]astrolabe.ProtectedEntityID, len(resp.Ids))
	for index, protoID := range resp.Ids {
		peIDs[index] = ProtectedEntityIDFromProto(protoID)
	}

	return peIDs, nil
}

func (recv *protectedEntityTypeManagerClient) Copy(ctx context.Context, pe astrolabe.ProtectedEntity, params map[string]map[string]interface{}, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (recv *protectedEntityTypeManagerClient) CopyFromInfo(ctx context.Context, info astrolabe.ProtectedEntityInfo, params map[string]map[string]interface{}, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}
