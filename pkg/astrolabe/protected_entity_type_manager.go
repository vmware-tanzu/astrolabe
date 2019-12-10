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

package astrolabe

import (
	//	"archive/zip"
	"context"
)

type CopyCreateOptions int

const (
	AllocateNewObject    CopyCreateOptions = 1
	UpdateExistingObject CopyCreateOptions = 2
	AllocateObjectWithID CopyCreateOptions = 3
)

type ProtectedEntityTypeManager interface {
	GetTypeName() string
	GetProtectedEntity(ctx context.Context, id ProtectedEntityID) (ProtectedEntity, error)
	GetProtectedEntities(ctx context.Context) ([]ProtectedEntityID, error)
	Copy(ctx context.Context, pe ProtectedEntity, options CopyCreateOptions) (ProtectedEntity, error)
	CopyFromInfo(ctx context.Context, info ProtectedEntityInfo, options CopyCreateOptions) (ProtectedEntity, error)
}
