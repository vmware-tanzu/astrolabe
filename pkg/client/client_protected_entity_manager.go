package client

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/gen/client"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"sync"
)

type ClientProtectedEntityManager struct {
	restClient       *client.Astrolabe
	typeManagers     map[string]ClientProtectedEntityTypeManager
	typeManagerMutex sync.Mutex
}

func NewClientProtectedEntityManager(restClient *client.Astrolabe) (ClientProtectedEntityManager, error) {
	returnClient := ClientProtectedEntityManager{
		restClient:       restClient,
		typeManagerMutex: sync.Mutex{},
	}
	err := returnClient.syncTypeManagers()
	if err != nil {
		return ClientProtectedEntityManager{}, err
	}
	return returnClient, nil
}

func (this ClientProtectedEntityManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	petm, ok := this.typeManagers[id.GetPeType()]
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not find manager for type %s", id.GetPeType()))
	}
	return petm.GetProtectedEntity(ctx, id)
}

func (this ClientProtectedEntityManager) GetProtectedEntityTypeManager(peType string) astrolabe.ProtectedEntityTypeManager {
	petm, ok := this.typeManagers[peType]
	if !ok {
		return nil
	}
	return petm
}

func (this ClientProtectedEntityManager) ListEntityTypeManagers() []astrolabe.ProtectedEntityTypeManager {
	this.typeManagerMutex.Lock()
	defer this.typeManagerMutex.Unlock()
	returnPETMs := []astrolabe.ProtectedEntityTypeManager{}
	for _, curPETM := range this.typeManagers {
		returnPETMs = append(returnPETMs, curPETM)
	}
	return returnPETMs
}

func (this *ClientProtectedEntityManager) syncTypeManagers() error {
	listResult, err := this.restClient.Operations.ListServices(nil)
	if err != nil {
		return errors.Wrap(err, "ListServices failed")
	}

	newPETMs := make(map[string]ClientProtectedEntityTypeManager, len(listResult.GetPayload().Services))
	for _, curService := range listResult.GetPayload().Services {
		newPETMs[curService] = NewClientProtectedEntityTypeManager(curService, this)
	}

	this.typeManagerMutex.Lock()
	defer this.typeManagerMutex.Unlock()
	this.typeManagers = newPETMs
	return nil
}
