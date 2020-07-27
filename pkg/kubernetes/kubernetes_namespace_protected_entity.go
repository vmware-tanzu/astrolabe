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

package kubernetes

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	v1 "k8s.io/api/core/v1"
)

type KubernetesNamespaceProtectedEntity struct {
	knpetm    *KubernetesNamespaceProtectedEntityTypeManager
	id        astrolabe.ProtectedEntityID
	namespace *v1.Namespace
	logger    logrus.FieldLogger
}

func (this *KubernetesNamespaceProtectedEntity) GetDataReader(context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func (this *KubernetesNamespaceProtectedEntity) GetMetadataReader(context.Context) (io.ReadCloser, error) {
	return nil, nil
}

func NewKubernetesNamespaceProtectedEntity(knpetm *KubernetesNamespaceProtectedEntityTypeManager,
	namespace *v1.Namespace) (*KubernetesNamespaceProtectedEntity, error) {
	nsPEID := astrolabe.NewProtectedEntityID("k8sns", namespace.Name)
	returnPE := KubernetesNamespaceProtectedEntity{
		knpetm:    knpetm,
		id:        nsPEID,
		namespace: namespace,
		logger:    knpetm.logger,
	}
	return &returnPE, nil
}

func (this *KubernetesNamespaceProtectedEntity) GetInfo(ctx context.Context) (astrolabe.ProtectedEntityInfo, error) {
	return nil, nil
}
func (this *KubernetesNamespaceProtectedEntity) GetCombinedInfo(ctx context.Context) ([]astrolabe.ProtectedEntityInfo, error) {
	return nil, nil

}

func (this *KubernetesNamespaceProtectedEntity) Snapshot(ctx context.Context, params map[string]map[string]interface{}) (astrolabe.ProtectedEntitySnapshotID, error) {
	return astrolabe.ProtectedEntitySnapshotID{}, nil

}
func (this *KubernetesNamespaceProtectedEntity) ListSnapshots(ctx context.Context) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	return nil, nil

}
func (this *KubernetesNamespaceProtectedEntity) DeleteSnapshot(ctx context.Context,
	snapshotToDelete astrolabe.ProtectedEntitySnapshotID) (bool, error) {
	return false, nil

}
func (this *KubernetesNamespaceProtectedEntity) GetInfoForSnapshot(ctx context.Context,
	snapshotID astrolabe.ProtectedEntitySnapshotID) (*astrolabe.ProtectedEntityInfo, error) {
	return nil, nil

}

func (this *KubernetesNamespaceProtectedEntity) GetComponents(ctx context.Context) ([]astrolabe.ProtectedEntity, error) {
	return nil, nil

}

func (this *KubernetesNamespaceProtectedEntity) GetID() astrolabe.ProtectedEntityID {
	return this.id
}

func (this *KubernetesNamespaceProtectedEntity) Overwrite(ctx context.Context, sourcePE astrolabe.ProtectedEntity, params map[string]map[string]interface{},
	overwriteComponents bool) error {
	return nil
}
