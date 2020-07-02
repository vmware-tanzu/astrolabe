package client

import (
	"context"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/gen/client/operations"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"time"
)

type ClientProtectedEntityTypeManager struct {
	entityManager *ClientProtectedEntityManager
	typeName      string
}

func NewClientProtectedEntityTypeManager(typeName string, entityManager *ClientProtectedEntityManager) ClientProtectedEntityTypeManager {
	return ClientProtectedEntityTypeManager{
		typeName:      typeName,
		entityManager: entityManager,
	}
}
func (this ClientProtectedEntityTypeManager) GetTypeName() string {
	return this.typeName
}

func (this ClientProtectedEntityTypeManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	params := operations.GetProtectedEntityInfoParams{
		Service:           this.typeName,
		ProtectedEntityID: id.String(),
	}
	params.SetTimeout(time.Minute)
	getInfoOK, err := this.entityManager.restClient.Operations.GetProtectedEntityInfo(&params)
	if err != nil {
		return ClientProtectedEntity{}, errors.Wrap(err, "Failed in GetProtectedEntityInfo")
	}
	peID, err := astrolabe.NewProtectedEntityIDFromModel(getInfoOK.GetPayload().ID)
	if err != nil {
		return ClientProtectedEntity{}, errors.Wrap(err, "Failed in parsing ID")
	}
	return NewClientProtectedEntity(peID, &this), nil
}

func (this ClientProtectedEntityTypeManager) GetProtectedEntities(ctx context.Context) ([]astrolabe.ProtectedEntityID, error) {
	params := operations.ListProtectedEntitiesParams{
		Service: this.typeName,
	}
	params.SetTimeout(time.Minute)
	listPEsOK, err := this.entityManager.restClient.Operations.ListProtectedEntities(&params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed in ListProtectedEntities")
	}
	returnPEIDs := make([]astrolabe.ProtectedEntityID, len(listPEsOK.GetPayload().List))
	for curPEIDNum, curPEID := range listPEsOK.GetPayload().List {
		returnPEIDs[curPEIDNum], err = astrolabe.NewProtectedEntityIDFromModel(curPEID)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed in to convert Protected Entity ID %v", curPEID)
		}
	}
	return returnPEIDs, nil
}

func (this ClientProtectedEntityTypeManager) Copy(ctx context.Context, pe astrolabe.ProtectedEntity, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (this ClientProtectedEntityTypeManager) CopyFromInfo(ctx context.Context, info astrolabe.ProtectedEntityInfo, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (this ClientProtectedEntityTypeManager) Delete(ctx context.Context, id astrolabe.ProtectedEntityID) error {
	panic("implement me")
}
