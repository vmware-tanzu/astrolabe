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
	"archive/zip"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
)

func ZipProtectedEntityToWriter(ctx context.Context, pe ProtectedEntity, writer io.Writer) error {
	zipWriter := zip.NewWriter(writer)
	peInfo, err := pe.GetInfo(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to get info for %s", pe.GetID().String())
	}
	jsonBuf, err := json.Marshal(peInfo)
	if err != nil {
		return errors.Wrapf(err, "Failed to get marshal info for %s", pe.GetID().String())
	}
	peInfoWriter, err := zipWriter.Create(pe.GetID().String() + ".peinfo")
	if err != nil {
		return errors.Wrapf(err, "Failed to create peinfo zip writer for %s", pe.GetID().String())
	}
	var bytesWritten int64
	jsonWritten, err := peInfoWriter.Write(jsonBuf)
	if err != nil {
		return errors.Wrapf(err, "Failed to write info json for %s", pe.GetID().String())
	}
	bytesWritten += int64(jsonWritten)

	mdReader, err := pe.GetMetadataReader(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to get metadata reader for %s", pe.GetID().String())
	}
	if mdReader != nil {
		peMDWriter, err := zipWriter.Create(pe.GetID().String() + ".md")
		if err != nil {
			return errors.Wrapf(err, "Failed to create metadata zip writer for %s", pe.GetID().String())
		}
		mdWritten, err := io.Copy(peMDWriter, mdReader)
		if err != nil {
			return errors.Wrapf(err, "Failed to write metadata for %s", pe.GetID().String())
		}
		bytesWritten += mdWritten
	}

	dataReader, err := pe.GetDataReader(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to get data reader for %s", pe.GetID().String())
	}
	if dataReader != nil {
		peDataWriter, err := zipWriter.Create(pe.GetID().String() + ".data")
		if err != nil {
			return errors.Wrapf(err, "Failed to create data zip writer for %s", pe.GetID().String())
		}
		dataWritten, err := io.CopyBuffer(peDataWriter, dataReader, make([]byte, 1024 * 1024))
		if err != nil {
			return errors.Wrapf(err, "Failed to write data for %s", pe.GetID().String())
		}
		bytesWritten += dataWritten
	}
	components, err := pe.GetComponents(ctx)
	if err != nil {
		return errors.Wrapf(err, "Failed to get components for %s", pe.GetID().String())
	}

	for _, curComponent := range components {
		header := zip.FileHeader{
			Name:   "components/" + curComponent.GetID().String() + ".zip",
			Method: zip.Store, // Do not compress components, these are zip files so won't get compressed more
			// and we need decompressed data in the stream in order to extract components
		}
		componentWriter, err := zipWriter.CreateHeader(&header)
		if err != nil {
			return errors.Wrapf(err, "Failed to get create component writer for component %s of %s",
				curComponent.GetID().String(), pe.GetID().String())
		}
		err = ZipProtectedEntityToWriter(ctx, curComponent, componentWriter)
		if err != nil {
			return errors.Wrapf(err, "Failed to write zip for component %s of %s",
				curComponent.GetID().String(), pe.GetID().String())
		}
	}
	err = zipWriter.Flush()
	if err != nil {
		return errors.Wrapf(err, "Zipwriter flush failed for for %s", pe.GetID().String())
	}
	err = zipWriter.Close()
	if err != nil {
		return errors.Wrapf(err, "Zipwriter close failed for for %s", pe.GetID().String())
	}
	return nil
}

func zipPE(ctx context.Context, pe ProtectedEntity, writer io.WriteCloser) error {
	defer writer.Close()
	err := ZipProtectedEntityToWriter(ctx, pe, writer)
	if err != nil {
		return errors.Errorf("Failed to zip protected entity %s, err = %v", pe.GetID().String(), err)
	}
	return nil
}

func ZipProtectedEntityToFile(ctx context.Context, srcPE ProtectedEntity, zipFileWriter io.WriteCloser) (bytesWritten int64, err error) {
	reader, writer := io.Pipe()
	go func() {
		err = zipPE(ctx, srcPE, writer)
	}()
	if err != nil {
		return
	}

	bytesWritten, err = io.CopyBuffer(zipFileWriter, reader, make([]byte, 512*1024))
	if err != nil {
		err = errors.Errorf("Error copying %v", err)
		return
	}

	return
}

type ZipProtectedEntity struct {
	info       ProtectedEntityInfo
	data       *zip.File
	metadata   *zip.File
	combined   io.ReaderAt
	components []ProtectedEntity
}

func (recv ZipProtectedEntity) GetInfo(ctx context.Context) (ProtectedEntityInfo, error) {
	return recv.info, nil
}

func (recv ZipProtectedEntity) GetCombinedInfo(ctx context.Context) ([]ProtectedEntityInfo, error) {
	panic("implement me")
}

func (recv ZipProtectedEntity) Snapshot(ctx context.Context, params map[string]map[string]interface{}) (ProtectedEntitySnapshotID, error) {
	panic("implement me")
}

func (recv ZipProtectedEntity) ListSnapshots(ctx context.Context) ([]ProtectedEntitySnapshotID, error) {
	panic("implement me")
}

func (recv ZipProtectedEntity) DeleteSnapshot(ctx context.Context, snapshotToDelete ProtectedEntitySnapshotID, params map[string]map[string]interface{}) (bool, error) {
	panic("implement me")
}

func (recv ZipProtectedEntity) GetInfoForSnapshot(ctx context.Context, snapshotID ProtectedEntitySnapshotID) (*ProtectedEntityInfo, error) {
	panic("implement me")
}

func (recv ZipProtectedEntity) GetComponents(ctx context.Context) ([]ProtectedEntity, error) {
	return recv.components, nil
}

func (recv ZipProtectedEntity) GetID() ProtectedEntityID {
	return recv.info.GetID()
}

func (recv ZipProtectedEntity) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	return recv.data.Open()
}

func (recv ZipProtectedEntity) GetMetadataReader(ctx context.Context) (io.ReadCloser, error) {
	return recv.metadata.Open()
}

func (recv ZipProtectedEntity) Overwrite(ctx context.Context, sourcePE ProtectedEntity, params map[string]map[string]interface{}, overwriteComponents bool) error {
	panic("implement me")
}

func GetPEFromZipStream(ctx context.Context, reader io.ReaderAt, size int64) (ProtectedEntity, error) {
	zipReader, err := zip.NewReader(reader, size)
	if err != nil {
		return nil, errors.Wrap(err, "could not init zip file")
	}

	retEntity := ZipProtectedEntity{
		combined:   reader,
		components: []ProtectedEntity{},
	}

	for _, checkFile := range zipReader.File {
		if strings.HasSuffix(checkFile.Name, ".peinfo") {
			peinfoReader, err := checkFile.Open()
			if err != nil {
				return nil, errors.Wrapf(err, "could not open %s", checkFile.Name)
			}
			peinfoBuf, err := ioutil.ReadAll(peinfoReader)
			if err != nil {
				return nil, errors.Wrapf(err, "could not read data for %s", checkFile.Name)
			}
			peinfo := ProtectedEntityInfoImpl{}
			err = json.Unmarshal(peinfoBuf, &peinfo)
			if err != nil {
				return nil, errors.Wrapf(err, "could not unmarshal JSON from %s", checkFile.Name)
			}
			retEntity.info = peinfo
		}
		if strings.HasSuffix(checkFile.Name, ".md") {
			retEntity.metadata = checkFile
			continue
		}
		if strings.HasSuffix(checkFile.Name, "data") {
			retEntity.data = checkFile
			continue
		}
		if strings.HasPrefix(checkFile.Name, "/components") {
			dataOffset, err := checkFile.DataOffset()
			if err != nil {
				return nil, errors.Wrapf(err, "Could not get data offset for %s", checkFile.Name)
			}
			componentReader := io.NewSectionReader(reader, dataOffset, int64(checkFile.CompressedSize64))
			componentPE, err := GetPEFromZipStream(ctx, componentReader, int64(checkFile.CompressedSize64))
			retEntity.components = append(retEntity.components, componentPE)
			continue
		}
		// TODO - Should log an error here for unknown type
	}
	return retEntity, nil
}

func UnzipFileToProtectedEntity(ctx context.Context, zipFileReader io.ReaderAt, zipFileSize int64, destPE ProtectedEntity) error {
	srcPE, err := GetPEFromZipStream(ctx, zipFileReader, zipFileSize)
	if err != nil {
		return errors.Errorf("Got err %v when unzipping PE from the source zip file", err)
	}

	var params map[string]map[string]interface{}
	if err := destPE.Overwrite(ctx, srcPE, params, true); err != nil {
		return errors.Errorf("Got err %v overwriting the dest PE %v with source PE %v", err, destPE.GetID(), srcPE.GetID())
	}

	return nil
}
