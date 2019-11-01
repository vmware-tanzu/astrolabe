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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"io/ioutil"
	"log"
)

type ProtectedEntity struct {
	rpetm  *ProtectedEntityTypeManager
	peinfo astrolabe.ProtectedEntityInfo
}

func NewProtectedEntityFromJSONBuf(rpetm *ProtectedEntityTypeManager, buf [] byte) (pe ProtectedEntity, err error) {
	peii := astrolabe.ProtectedEntityInfoImpl{}
	err = json.Unmarshal(buf, &peii)
	if err != nil {
		return
	}
	pe.peinfo = peii
	pe.rpetm = rpetm
	return
}

func NewProtectedEntityFromJSONReader(rpetm *ProtectedEntityTypeManager, reader io.Reader) (pe ProtectedEntity, err error) {
	decoder := json.NewDecoder(reader)
	peInfo := astrolabe.ProtectedEntityInfoImpl{}
	err = decoder.Decode(&peInfo)
	if err == nil {
		pe.peinfo = peInfo
		pe.rpetm = rpetm
	}
	return
}
func (this ProtectedEntity) GetInfo(ctx context.Context) (astrolabe.ProtectedEntityInfo, error) {
	return this.peinfo, nil
}

func (ProtectedEntity) GetCombinedInfo(ctx context.Context) ([]astrolabe.ProtectedEntityInfo, error) {
	panic("implement me")
}

func (ProtectedEntity) Snapshot(ctx context.Context) (astrolabe.ProtectedEntitySnapshotID, error) {
	return astrolabe.ProtectedEntitySnapshotID{}, errors.New("Snapshot not supported")
}

func (this ProtectedEntity) ListSnapshots(ctx context.Context) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	peID := this.peinfo.GetID()
	idPrefix := peID.GetPeType() + ":" + peID.GetID()
	retPEIDs, err := this.rpetm.GetProtectedEntitiesByIDPrefix(ctx, idPrefix)
	if err != nil {
		return nil, errors.Wrapf(err, "Failed to get PEs by the id prefix, %s", idPrefix)
	}
	retPESnapshotIDs := make([]astrolabe.ProtectedEntitySnapshotID, len(retPEIDs))
	for index, retPEID := range retPEIDs {
		retPESnapshotIDs[index] = retPEID.GetSnapshotID()
	}
	return retPESnapshotIDs, nil
}

func (this ProtectedEntity) DeleteSnapshot(ctx context.Context, snapshotToDelete astrolabe.ProtectedEntitySnapshotID) (bool, error) {
	bucket := this.rpetm.bucket
	peID := this.rpetm.peinfoName(this.peinfo.GetID())
	_, err := this.rpetm.s3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key: aws.String(peID),
	})
	if err != nil {
		return false, errors.Wrapf(err, "Unable to delete object %q from bucket %q", peID, bucket)
	}

	err = this.rpetm.s3.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(peID),
	})
	if err != nil {
		return false, errors.Wrapf(err, "Error occurred while waiting for object %q to be deleted", peID)
	}
	return true, nil
}

func (ProtectedEntity) GetInfoForSnapshot(ctx context.Context, snapshotID astrolabe.ProtectedEntitySnapshotID) (*astrolabe.ProtectedEntityInfo, error) {
	panic("implement me")
}

func (ProtectedEntity) GetComponents(ctx context.Context) ([]astrolabe.ProtectedEntity, error) {
	panic("implement me")
}

func (this ProtectedEntity) GetID() astrolabe.ProtectedEntityID {
	return this.peinfo.GetID()
}

func (this ProtectedEntity) GetDataReader(context.Context) (io.ReadCloser, error) {
	if len(this.peinfo.GetDataTransports()) > 0 {
		dataName := this.rpetm.dataName(this.GetID())
		return this.getReader(dataName)
	}
	return nil, nil
}

func (this ProtectedEntity) GetMetadataReader(context.Context) (io.Reader, error) {
	if len(this.peinfo.GetMetadataTransports()) > 0 {
		metadataName := this.rpetm.metadataName(this.GetID())
		return this.getReader(metadataName)
	}
	return nil, nil
}

func (this *ProtectedEntity) uploadStream(ctx context.Context, name string, reader io.Reader) error {
	uploader := s3manager.NewUploader(&this.rpetm.session)

	result, err := uploader.Upload(&s3manager.UploadInput{
		Body:   reader,
		Bucket: aws.String(this.rpetm.bucket),
		Key:    aws.String(name),
	})
	if err == nil {
		log.Println("Successfully uploaded to", result.Location)
	}
	return err
}

func (this *ProtectedEntity) copy(ctx context.Context, dataReader io.Reader,
	metadataReader io.Reader) error {
	peInfo := this.peinfo
	peinfoName := this.rpetm.peinfoName(peInfo.GetID())

	peInfoBuf, err := json.Marshal(peInfo)
	if err != nil {
		return err
	}
	if len(peInfoBuf) > maxPEInfoSize {
		return errors.New("JSON for pe info > 16K")
	}

	// TODO: defer the clean up of disk handle of source PE's data reader
	if dataReader != nil {
		dataName := this.rpetm.dataName(peInfo.GetID())
		err = this.uploadStream(ctx, dataName, dataReader)
		if err != nil {
			return err
		}
	}

	if metadataReader != nil {
		mdName := this.rpetm.metadataName(peInfo.GetID())
		err = this.uploadStream(ctx, mdName, metadataReader)
		if err != nil {
			return err
		}
	}
	jsonBytes := bytes.NewReader(peInfoBuf)

	jsonParams := &s3.PutObjectInput{
		Bucket:        aws.String(this.rpetm.bucket),
		Key:           aws.String(peinfoName),
		Body:          jsonBytes,
		ContentLength: aws.Int64(int64(len(peInfoBuf))),
		ContentType:   aws.String(peInfoFileType),
	}
	_, err = this.rpetm.s3.PutObject(jsonParams)
	if err != nil {
		return err
	}
	return err
}

func (this *ProtectedEntity) getReader(key string) (io.ReadCloser, error) {
	s3Object, err := this.rpetm.s3.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(this.rpetm.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}
	/*
		downloadMgr := s3manager.NewDownloaderWithClient(&this.rpetm.s3, func(d *s3manager.Downloader) {
			d.Concurrency = 1
			d.PartSize = 16 * 1024 * 1024
		})

		reader, writer := io.Pipe()
		seqWriterAt := util.NewSeqWriterAt(writer)
		go func() {
			defer writer.Close()
			downloadMgr.Download(seqWriterAt, &s3.GetObjectInput{
				Bucket: aws.String(this.rpetm.bucket),
				Key:    aws.String(key),
			})
			fmt.Printf("Download finished")
		}()
	*/
	s3BufferedReader := bufio.NewReaderSize(s3Object.Body, 1024*1024)
	return ioutil.NopCloser(s3BufferedReader), nil
}
