/*
 * Copyright the Astrolabe contributors
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

package localsnap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

type LocalSnapshotRepo struct {
	repoDir  string
	typename string
}

func NewLocalSnapshotRepo(typename string, repoDir string) (LocalSnapshotRepo, error) {
	err := ensureDirectoryExists(repoDir)
	if err != nil {
		return LocalSnapshotRepo{}, errors.Wrap(err, "repo dir cannot be used")
	}
	typeRepoDir := filepath.Join(repoDir, typename)
	err = ensureDirectoryExists(typeRepoDir)
	if err != nil {
		return LocalSnapshotRepo{}, errors.Wrap(err, "type repo dir cannot be used")
	}

	return LocalSnapshotRepo{
		repoDir:  typeRepoDir,
		typename: typename,
	}, nil
}

func ensureDirectoryExists(ensureDir string) error {
	ensureDirInfo, err := os.Stat(ensureDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(ensureDir, 0700)
			if err != nil {
				return errors.Wrapf(err, "could not create directory %s", ensureDir)
			}
			ensureDirInfo, err = os.Stat(ensureDir)
		}
		if err != nil {
			return errors.Wrapf(err, "stat on dir %s failed", ensureDir)
		}
	}
	if !ensureDirInfo.IsDir() {
		return errors.New(fmt.Sprintf("dir %s is not a directory", ensureDir))
	}
	return nil
}

func (this LocalSnapshotRepo) WriteProtectedEntity(ctx context.Context, pe astrolabe.ProtectedEntity, snapshotID astrolabe.ProtectedEntitySnapshotID) error {
	_, err := this.writeProtectedEntityInfo(ctx, pe, snapshotID)
	if err != nil {
		return errors.Wrapf(err, "failed to write peinfo for %s:%s", pe.GetID().String(), snapshotID.String())
	}
	_, err = this.writeMetadata(ctx, pe, snapshotID)
	if err != nil {
		return errors.Wrapf(err, "failed to write metadata for %s:%s", pe.GetID().String(), snapshotID.String())
	}
	_, err = this.writeData(ctx, pe, snapshotID)
	if err != nil {
		return errors.Wrapf(err, "failed to write data for %s:%s", pe.GetID().String(), snapshotID.String())
	}
	return nil
}

const snapshotDataExtension = ".snap-data"
const snapshotMDExtension = ".snap-md"
const snapshotPEInfoExtension = ".snap-peinfo"

func dataFilenameForSnapshot(peid astrolabe.ProtectedEntityID, snapshotID astrolabe.ProtectedEntitySnapshotID) string {
	return peid.IDWithSnapshot(snapshotID).String() + snapshotDataExtension
}

func mdFilenameForSnapshot(peid astrolabe.ProtectedEntityID, snapshotID astrolabe.ProtectedEntitySnapshotID) string {
	return peid.IDWithSnapshot(snapshotID).String() + snapshotMDExtension
}

func peinfoFilenameForSnapshot(peid astrolabe.ProtectedEntityID, snapshotID astrolabe.ProtectedEntitySnapshotID) string {
	return peid.IDWithSnapshot(snapshotID).String() + snapshotPEInfoExtension
}

func (this LocalSnapshotRepo) getSnapshotDirForPE(peid astrolabe.ProtectedEntityID) (string, error) {
	peSnapshotDir := filepath.Join(this.repoDir, peid.GetBaseID().String())
	peSnapshotDirInfo, err := os.Stat(peSnapshotDir)
	if err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(peSnapshotDir, 0700)
			if err != nil {
				return "", errors.Wrapf(err, "could not create snapshot directory %s", peSnapshotDir)
			}
			peSnapshotDirInfo, err = os.Stat(peSnapshotDir)
		}
		if err != nil {
			return "", errors.Wrapf(err, "stat on snapshotsdir %s failed", peSnapshotDir)
		}
	}
	if !peSnapshotDirInfo.IsDir() {
		return "", errors.New(fmt.Sprintf("snapshot directory %s for pe %s is not a directory", peSnapshotDir, peid.GetBaseID().String()))
	}
	return peSnapshotDir, nil
}

func (this LocalSnapshotRepo) writeData(ctx context.Context, pe astrolabe.ProtectedEntity, snapshotID astrolabe.ProtectedEntitySnapshotID) (int64, error) {
	dataReader, err := pe.GetDataReader(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "could not get datareader for %s", pe.GetID().String())
	}
	if dataReader != nil {
		defer dataReader.Close()
		dataWriter, err := this.GetDataWriter(ctx, pe.GetID(), snapshotID)
		if err != nil {
			return 0, errors.Wrapf(err, "could not get md reader for %s", pe.GetID().String())
		}
		defer dataWriter.Close()
		return io.Copy(dataWriter, dataReader)
	} else {
		return 0, nil
	}
}

func (this LocalSnapshotRepo) writeMetadata(ctx context.Context, pe astrolabe.ProtectedEntity, snapshotID astrolabe.ProtectedEntitySnapshotID) (int64, error) {
	mdReader, err := pe.GetMetadataReader(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "could not get md reader for %s", pe.GetID().String())
	}
	if mdReader != nil {
		defer mdReader.Close()
		mdWriter, err := this.GetMetadataWriter(ctx, pe.GetID(), snapshotID)
		if err != nil {
			return 0, errors.Wrapf(err, "could not get md reader for %s", pe.GetID().String())
		}
		defer mdWriter.Close()
		return io.Copy(mdWriter, mdReader)
	} else {
		return 0, nil
	}
}

func (this LocalSnapshotRepo) writeProtectedEntityInfo(ctx context.Context, pe astrolabe.ProtectedEntity, snapshotID astrolabe.ProtectedEntitySnapshotID) (int64, error) {
	peInfo, err := pe.GetInfo(ctx)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not get peinfo for %s", pe.GetID().String())
	}
	peInfoBytes, err := json.Marshal(peInfo)
	peInfoReader := bytes.NewReader(peInfoBytes)
	peInfoWriter, err := this.GetPEInfoWriter(ctx, pe.GetID(), snapshotID)
	if err != nil {
		return 0, errors.Wrapf(err, "Could not create peinfo file")
	}
	defer peInfoWriter.Close()
	return io.Copy(peInfoWriter, peInfoReader)
}

func (this LocalSnapshotRepo) GetPEInfoForID(ctx context.Context, peid astrolabe.ProtectedEntityID) (astrolabe.ProtectedEntityInfo, error) {
	if !peid.HasSnapshot() {
		return nil, errors.New(fmt.Sprintf("peid %s is not a snapshot peid", peid.String()))
	}
	peinfoReader, err := this.GetPEInfoReader(ctx, peid)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot get peinforeader for %s", peid.String())
	}
	peinfoBytes, err := ioutil.ReadAll(peinfoReader)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot read peinfo for %s", peid.String())
	}
	peInfo := astrolabe.ProtectedEntityInfoImpl{}

	err = json.Unmarshal(peinfoBytes, &peInfo)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse peinfo for %s", peid.String())
	}
	return peInfo, nil
}

func (this LocalSnapshotRepo) GetDataWriter(ctx context.Context, peid astrolabe.ProtectedEntityID, snapshotPEID astrolabe.ProtectedEntitySnapshotID) (io.WriteCloser, error) {
	dataFilename := dataFilenameForSnapshot(peid, snapshotPEID)
	return this.getWriter(peid, dataFilename)
}

func (this LocalSnapshotRepo) GetMetadataWriter(ctx context.Context, peid astrolabe.ProtectedEntityID, snapshotID astrolabe.ProtectedEntitySnapshotID) (io.WriteCloser, error) {
	mdFilename := mdFilenameForSnapshot(peid, snapshotID)
	return this.getWriter(peid, mdFilename)
}

func (this LocalSnapshotRepo) GetPEInfoWriter(ctx context.Context, peid astrolabe.ProtectedEntityID, snapshotID astrolabe.ProtectedEntitySnapshotID) (io.WriteCloser, error) {
	peinfoFilename := peinfoFilenameForSnapshot(peid, snapshotID)
	return this.getWriter(peid, peinfoFilename)
}

func (this LocalSnapshotRepo) GetPEInfoReader(ctx context.Context, peid astrolabe.ProtectedEntityID) (io.ReadCloser, error) {
	peinfoFilename := peinfoFilenameForSnapshot(peid, peid.GetSnapshotID())
	peSnapshotDir, err := this.getSnapshotDirForPE(peid)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get snapshot dir for pe %s", peid.String())
	}
	peInfoPath := filepath.Join(peSnapshotDir, peinfoFilename)
	returnReader, err := os.Open(peInfoPath)
	if err != nil {
		err = errors.Wrap(err, "Could not open data stream")
		return nil, err
	}
	return returnReader, nil
}

func (this LocalSnapshotRepo) getWriter(peid astrolabe.ProtectedEntityID, dataFilename string) (io.WriteCloser, error) {
	peSnapshotDir, err := this.getSnapshotDirForPE(peid)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get snapshot dir for pe %s", peid.String())
	}
	dataPath := filepath.Join(peSnapshotDir, dataFilename)
	writer, err := os.Create(dataPath)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create file %s", dataPath)
	}
	return writer, nil
}

func (this LocalSnapshotRepo) ListSnapshotsForPEID(peid astrolabe.ProtectedEntityID) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	peSnapshotDir := filepath.Join(this.repoDir, peid.String())
	peSnapshotDirInfo, err := os.Stat(peSnapshotDir)
	var returnIDs = make([]astrolabe.ProtectedEntitySnapshotID, 0)
	if err != nil {
		if os.IsNotExist(err) {
			return []astrolabe.ProtectedEntitySnapshotID{}, nil
		}
		return returnIDs, errors.Wrapf(err, "could not snap snapshot dir %s", peSnapshotDir)
	}
	if !peSnapshotDirInfo.IsDir() {
		return []astrolabe.ProtectedEntitySnapshotID{}, errors.New(fmt.Sprintf("snapshot directory %s for pe %s is not a directory", peSnapshotDir, peid.String()))
	}
	fileInfo, err := ioutil.ReadDir(peSnapshotDir)
	if err != nil {
		return returnIDs, errors.Wrapf(err, "could not read snapshot dir %s", peSnapshotDir)
	}

	for _, curFileInfo := range fileInfo {
		if strings.HasSuffix(curFileInfo.Name(), snapshotDataExtension) {
			peid, err := snapshotPEIDForFilename(curFileInfo.Name())
			if err != nil {

			} else {
				returnIDs = append(returnIDs, peid.GetSnapshotID())
			}
		}
	}
	return returnIDs, nil
}

func snapshotPEIDForFilename(filename string) (astrolabe.ProtectedEntityID, error) {
	if !strings.HasSuffix(filename, snapshotDataExtension) {
		return astrolabe.ProtectedEntityID{}, errors.New(fmt.Sprintf("%s does not end with %s", filename, snapshotDataExtension))
	}
	peidStr := strings.TrimSuffix(filename, snapshotDataExtension)
	return astrolabe.NewProtectedEntityIDFromString(peidStr)
}

func (this LocalSnapshotRepo) GetDataReaderForSnapshot(peid astrolabe.ProtectedEntityID) (io.ReadCloser, error) {
	if !peid.HasSnapshot() {
		return nil, errors.New(fmt.Sprintf("peid %s does not have a snapshot id", peid.String()))
	}
	dataFilename := dataFilenameForSnapshot(peid, peid.GetSnapshotID())
	peSnapshotDir, err := this.getSnapshotDirForPE(peid)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get snapshot dir for pe %s", peid.String())
	}
	peDataPath := filepath.Join(peSnapshotDir, dataFilename)
	reader, err := os.Open(peDataPath)
	if err != nil {
		err = errors.Wrap(err, "Could not open data stream")
	}
	return reader, err
}

func (this LocalSnapshotRepo) GetMetadataReaderForSnapshot(peid astrolabe.ProtectedEntityID) (io.ReadCloser, error) {
	if !peid.HasSnapshot() {
		return nil, errors.New(fmt.Sprintf("peid %s does not have a snapshot id", peid.String()))
	}
	mdFilename := mdFilenameForSnapshot(peid, peid.GetSnapshotID())
	peSnapshotDir, err := this.getSnapshotDirForPE(peid)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get snapshot dir for pe %s", peid.String())
	}
	peMDPath := filepath.Join(peSnapshotDir, mdFilename)
	reader, err := os.Open(peMDPath)
	if err != nil {
		err = errors.Wrap(err, "Could not open metadata stream")
	}
	return reader, err
}
