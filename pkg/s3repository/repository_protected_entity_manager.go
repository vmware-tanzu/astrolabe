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

package s3repository

import (
	"context"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"strings"
	"sync"
)

type RepositoryProtectedEntityManager struct {
	petms   map[string]astrolabe.ProtectedEntityTypeManager
	session session.Session
	bucket  string
	prefix  string
	logger  logrus.FieldLogger
	mutex   sync.Mutex
}

func NewS3RepositoryProtectedEntityManager(session session.Session, bucket string,
	prefix string, logger logrus.FieldLogger) (*RepositoryProtectedEntityManager, error) {
	s3Client := s3.New(&session)
	delimiter := "/"
	if !strings.HasSuffix(prefix, "/") {
		prefix = prefix + "/"
	}
	loi := s3.ListObjectsV2Input{
		Bucket:    &bucket,
		Delimiter: &delimiter,
		Prefix:    &prefix,
	}
	loo, err := s3Client.ListObjectsV2(&loi)
	if err != nil {
		return nil, errors.WithMessagef(err, "Failed retrieving prefixes bucket=%s, delimeter=%s, prefix=%s", loi.Bucket, loi.Delimiter, loi.Prefix)
	}
	returnPEM := RepositoryProtectedEntityManager{
		petms:   make(map[string]astrolabe.ProtectedEntityTypeManager),
		session: session,
		bucket:  bucket,
		prefix:  prefix,
		logger:  logger,
	}
	for _, curPrefix := range loo.CommonPrefixes {
		curPrefixStr := curPrefix.String()
		components := strings.Split(curPrefixStr, "/")
		curTypeName := components[len(components)-2]
		curPETM, err := NewS3RepositoryProtectedEntityTypeManager(curTypeName, session, bucket, prefix, logger)
		if err != nil {

		} else {
			returnPEM.petms[curTypeName] = curPETM
		}
	}

	return &returnPEM, nil
}

func (recv RepositoryProtectedEntityManager) GetProtectedEntity(ctx context.Context, id astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntity, error) {
	petm := recv.GetProtectedEntityTypeManager(id.GetPeType())
	if petm == nil {
		return nil, errors.Errorf("No PETM for type %s, cannot retrieve ID %s", id.GetPeType(), id.String())
	}
	return petm.GetProtectedEntity(ctx, id)
}

func (recv RepositoryProtectedEntityManager) GetProtectedEntityTypeManager(peType string) astrolabe.ProtectedEntityTypeManager {
	recv.mutex.Lock()
	defer recv.mutex.Unlock()
	returnPETM := recv.petms[peType]
	if returnPETM == nil {
		var err error
		returnPETM, err = NewS3RepositoryProtectedEntityTypeManager(peType, recv.session, recv.bucket, recv.prefix, recv.logger)
		if err != nil {
			recv.logger.Errorf("Could not create S3RepositoryProtectedEntityTypeManager for type %s, err = %v", peType, err)
		}
	}
	return returnPETM
}

func (recv RepositoryProtectedEntityManager) ListEntityTypeManagers() []astrolabe.ProtectedEntityTypeManager {
	recv.mutex.Lock()
	defer recv.mutex.Unlock()
	returnPETMs := make([]astrolabe.ProtectedEntityTypeManager, len(recv.petms))
	curPETMNum := 0
	for _, curPETM := range recv.petms {
		returnPETMs[curPETMNum] = curPETM
		curPETMNum++
	}
	return returnPETMs
}
