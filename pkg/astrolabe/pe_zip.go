/*
 * Copyright 2019 VMware, Inc..
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
	"archive/zip"
	"context"
	"encoding/json"
	"io"
)

func ZipProtectedEntity(ctx context.Context, entity ProtectedEntity, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	peInfo, err := entity.GetInfo(ctx)
	if (err != nil) {
		return err
	}
	jsonBuf, err := json.Marshal(peInfo)
	if (err != nil) {
		return err
	}
	peInfoWriter, err := zipWriter.Create(entity.GetID().String() + ".peinfo")
	if (err != nil) {
		return err
	}
	_, err = peInfoWriter.Write(jsonBuf)
	if (err != nil) {
		return err
	}
	return nil
}
