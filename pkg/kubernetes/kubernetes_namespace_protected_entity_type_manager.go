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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesNamespaceProtectedEntityTypeManager struct {
	clientset  *kubernetes.Clientset
	namespaces map[string]*KubernetesNamespaceProtectedEntity
	logger     logrus.FieldLogger
}

func NewKubernetesNamespaceProtectedEntityTypeManagerFromConfig(params map[string]interface{}, s3URLBase string,
	logger logrus.FieldLogger) (*KubernetesNamespaceProtectedEntityTypeManager, error) {
	masterURL := params["masterURL"].(string)
	kubeconfigPath := params["kubeconfigPath"].(string)
	config, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfigPath)
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	returnTypeManager := KubernetesNamespaceProtectedEntityTypeManager{
		clientset: clientset,
		logger:    logger,
	}
	returnTypeManager.namespaces = make(map[string]*KubernetesNamespaceProtectedEntity)
	err = returnTypeManager.loadNamespaceEntities()
	if err != nil {
		return nil, err
	}
	return &returnTypeManager, nil
}

func (this *KubernetesNamespaceProtectedEntityTypeManager) GetTypeName() string {
	return "kubernetes-ns"
}

func (this *KubernetesNamespaceProtectedEntityTypeManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (
	astrolabe.ProtectedEntity, error) {
	return nil, nil
}

func (this *KubernetesNamespaceProtectedEntityTypeManager) GetProtectedEntities(ctx context.Context) ([]astrolabe.ProtectedEntityID, error) {
	//TODO - fix concurrency issues here
	protectedEntities := make([]astrolabe.ProtectedEntityID, 0, len(this.namespaces))

	for _, pe := range this.namespaces {
		protectedEntities = append(protectedEntities, pe.GetID())
	}
	return protectedEntities, nil
}

func (this *KubernetesNamespaceProtectedEntityTypeManager) loadNamespaceEntities() error {
	namespaceList, err := this.clientset.CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, namespace := range namespaceList.Items {
		k8snsPE, err := NewKubernetesNamespaceProtectedEntity(this, &namespace)
		if err == nil {
			key := k8snsPE.id.String()
			if _, exists := this.namespaces[key]; !exists {
				this.namespaces[key] = k8snsPE
			}
		} else {
			// log it
		}
	}
	return nil
}

func (this *KubernetesNamespaceProtectedEntityTypeManager) Copy(ctx context.Context, pe astrolabe.ProtectedEntity, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	return nil, nil
}

func (this *KubernetesNamespaceProtectedEntityTypeManager) CopyFromInfo(ctx context.Context, pe astrolabe.ProtectedEntityInfo, options astrolabe.CopyCreateOptions) (astrolabe.ProtectedEntity, error) {
	return nil, nil
}
