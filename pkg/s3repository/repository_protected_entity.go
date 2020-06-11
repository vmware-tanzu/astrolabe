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
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"io/ioutil"
	"sort"
	"strings"
)

type ProtectedEntity struct {
	rpetm  *ProtectedEntityTypeManager
	peinfo astrolabe.ProtectedEntityInfo
}

/*
 * S3Segment represents a single S3 object (segment) for a data stream.  Each data/metadata stream consists of
 * one or more S3Segments
 */
type s3Segment struct {
	segmentNumber       int
	startOffset, length int64
	key                 string
}

func NewProtectedEntityFromJSONBuf(rpetm *ProtectedEntityTypeManager, buf []byte) (pe ProtectedEntity, err error) {
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
	var err error
	bucket := this.rpetm.bucket

	peinfoName := this.rpetm.peinfoName(this.peinfo.GetID())
	_, err = this.deleteSnapshotComponents(ctx, bucket, peinfoName)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to delete peinfo from bucket %q", bucket)
	}

	mdName := this.rpetm.metadataName(this.peinfo.GetID())
	_, err = this.deleteSnapshotComponents(ctx, bucket, mdName)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to delete metadata from bucket %q", bucket)
	}

	dataName := this.rpetm.dataName(this.peinfo.GetID())
	_, err = this.deleteSnapshotComponents(ctx, bucket, dataName)
	if err != nil {
		return false, errors.Wrapf(err, "Failed to delete data from bucket %q", bucket)
	}

	return true, nil
}

func (this ProtectedEntity) getS3Segments(ctx context.Context, bucket string, componentName string) ([]s3Segment, error) {
	awsBucket := aws.String(bucket)
	awsKey := aws.String(componentName)
	var returnSegments []s3Segment
	err := this.rpetm.s3.ListObjectsPagesWithContext(ctx, &s3.ListObjectsInput{
		Bucket: awsBucket,
		Prefix: awsKey,
	}, func(output *s3.ListObjectsOutput, b bool) bool {
		for _, componentPart := range output.Contents {
			_, partNumber, startOffset := parseSegmentName(*componentPart.Key)
			returnSegments = append(returnSegments, s3Segment{
				segmentNumber: partNumber,
				startOffset:   startOffset,
				length:        *componentPart.Size,
				key:           *componentPart.Key,
			})
		}
		return true
	})
	sort.Slice(returnSegments, func(i, j int) bool {
		return returnSegments[i].segmentNumber < returnSegments[j].segmentNumber
	})
	return returnSegments, err
}
func (this ProtectedEntity) deleteSnapshotComponents(ctx context.Context, bucket string, componentName string) (bool, error) {
	awsBucket := aws.String(bucket)
	awsKey := aws.String(componentName)

	var combinedErrors []error
	err := this.rpetm.s3.ListObjectsPagesWithContext(ctx, &s3.ListObjectsInput{
		Bucket: awsBucket,
		Prefix: awsKey,
	}, func(output *s3.ListObjectsOutput, b bool) bool {
		for _, deleteObject := range output.Contents {
			_, err := this.rpetm.s3.DeleteObject(&s3.DeleteObjectInput{
				Bucket: awsBucket,
				Key:    deleteObject.Key,
			})
			if err == nil {
				err = this.rpetm.s3.WaitUntilObjectNotExists(&s3.HeadObjectInput{
					Bucket: awsBucket,
					Key:    deleteObject.Key,
				})
			}
			if err != nil {
				combinedErrors = append(combinedErrors, err)
			}
		}
		return true
	})
	if len(combinedErrors) > 0 {
		var combinedString string
		for _, curErr := range combinedErrors {
			combinedString += curErr.Error() + "\n"
		}
		err = errors.New("Multiple errors:\n" + combinedString)
	}
	if err != nil {
		return false, errors.Wrapf(err, "Unable to delete object %q from bucket %q", componentName, bucket)
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

func (this ProtectedEntity) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	if len(this.peinfo.GetDataTransports()) > 0 {
		dataName := this.rpetm.dataName(this.GetID())
		return this.getReader(ctx, dataName)
	}
	return nil, nil
}

func (this ProtectedEntity) GetMetadataReader(ctx context.Context) (io.Reader, error) {
	if len(this.peinfo.GetMetadataTransports()) > 0 {
		metadataName := this.rpetm.metadataName(this.GetID())
		return this.getReader(ctx, metadataName)
	}
	return nil, nil
}

func (this *ProtectedEntity) uploadStream(ctx context.Context, name string, maxSegmentSize int64, reader io.Reader) error {
	partNum := 0
	var startOffset int64
	// getS3Segments returns a list of existing segments, sorted by part number.  There may be gaps in the list if
	// segments were not uploaded
	this.rpetm.logger.Debugf("Searching for existing segments bucket: %v name: %v", this.rpetm.bucket, name)
	existingSegments, err := this.getS3Segments(ctx, this.rpetm.bucket, name)
	if err != nil {
		return err
	}

	this.rpetm.logger.Infof("Found %d existing segments for bucket: %v name: %v", len(existingSegments), this.rpetm.bucket, name)

	// Check here to make sure that the maxSegmentSize matches existing segments if any.  On mismatch, delete existing segments
	segmentNumber := 0
	for true {
		uploadSegment := true
		var bytesThisSegment int64
		if segmentNumber < len(existingSegments) {
			if existingSegments[segmentNumber].startOffset == startOffset {
				segmentNumber++
				uploadSegment = false
				bytesThisSegment, err = skipBytes(reader, existingSegments[segmentNumber].length, nil)
				if err != nil {
					return err
				}
			}
		}
		if uploadSegment {
			bytesThisSegment, err = this.uploadSegment(ctx, name, partNum, startOffset, maxSegmentSize, reader)
			if err != nil {
				return err
			}
			if bytesThisSegment < maxSegmentSize {
				return nil
			}
		}
		partNum++
		startOffset += bytesThisSegment
	}
	return nil
}

/*
 * skipBytes will either use the Seek API, if the Reader implements Seeker, or read and discard bytes from the
 * Reader to move the offset forward.  If discardBuf is set, it will be used to read bytes into for discarding (this is
 * simply to avoid  allocating a buffer when one is available)
 */
func skipBytes(reader io.Reader, bytesToSkip int64, discardBuf []byte) (int64, error) {
	if bytesToSkip < 0 {
		return 0, errors.New("Cannot skip negative bytes")
	}
	if bytesToSkip == 0 {
		return 0, nil // Bounce out here so we'll never get down to allocating a discard buf
	}
	var bytesSkipped int64
	if rs, ok := interface{}(reader).(io.Seeker); ok {
		startOffset, err := rs.Seek(0, io.SeekCurrent)
		if err != nil {
			return 0, nil
		}

		endOffset, err := rs.Seek(bytesToSkip, io.SeekCurrent)
		if err != nil {
			return 0, nil
		}
		bytesSkipped = endOffset - startOffset
	} else {
		if discardBuf == nil {
			bufSize := bytesToSkip
			if bytesToSkip > MaxBufferSize {
				bufSize = MaxBufferSize
			}
			discardBuf = make([]byte, bufSize)
		}
		var bytesSkipped int64
		for bytesSkipped < bytesToSkip {
			curBuf := discardBuf
			if bytesToSkip < int64(len(discardBuf)) {
				curBuf = discardBuf[0:bytesToSkip]
			}
			bytesRead, err := reader.Read(curBuf)
			if err != nil && err != io.EOF {
				return bytesSkipped, errors.Wrap(err, "Read failed while skipping")
			}
			bytesSkipped += int64(bytesRead)
		}
	}
	return bytesSkipped, nil
}
func segmentName(baseName string, partNumber int, startOffset int64) string {
	return baseName + "/" + fmt.Sprintf("%06d-%016d", partNumber, startOffset)
}

func parseSegmentName(key string) (baseName string, partNumber int, startOffset int64) {
	lastSep := strings.LastIndexByte(key, '/')
	if lastSep < 0 {
		return "", -1, -1
	}
	baseName = key[0:lastSep]
	segmentName := key[lastSep+1 : len(key)]
	fmt.Sscanf(segmentName, "%d-%d", &partNumber, &startOffset)
	return
}

func (this *ProtectedEntity) uploadSegment(ctx context.Context, baseName string, part int, startOffset int64, maxSegmentSize int64, reader io.Reader) (int64, error) {
	log := this.rpetm.logger

	if maxSegmentSize > this.rpetm.maxSegmentSize {
		return 0, errors.Errorf("maxSegmentSize %d too large (%d is kSegmentSizeLimit)", maxSegmentSize, SegmentSizeLimit)
	}
	name := segmentName(baseName, part, startOffset)
	awsBucketName := aws.String(this.rpetm.bucket)
	awsKey := aws.String(name)
	defer this.abortPendingMultipartUpload(&ctx, awsBucketName, awsKey)
	var uploadID string
	completedParts := make([]*s3.Part, this.rpetm.maxParts)
	// First, check to see if there's an on-going multipart upload.  We assume that there will not be multiple
	// repository managers trying to upload simultaneously, so if we find a multipart upload for this key, it needs
	// to be restarted
	listUploadsInput := &s3.ListMultipartUploadsInput{
		Bucket: awsBucketName,
		Prefix: awsKey,
	}

	multiPartUploadsOutput, err := this.rpetm.s3.ListMultipartUploadsWithContext(ctx, listUploadsInput)

	if err != nil {
		return 0, err
	}

	if len(multiPartUploadsOutput.Uploads) > 0 {
		if len(multiPartUploadsOutput.Uploads) == 1 {
			listPartsInput := &s3.ListPartsInput{
				Bucket:           awsBucketName,
				Key:              awsKey,
				MaxParts:         nil,
				PartNumberMarker: nil,
				RequestPayer:     nil,
				UploadId:         multiPartUploadsOutput.Uploads[0].UploadId,
			}
			err = this.rpetm.s3.ListPartsPagesWithContext(ctx, listPartsInput, func(listPartsOutput *s3.ListPartsOutput, b bool) bool {
				for _, curPart := range listPartsOutput.Parts {
					completedParts[int(*curPart.PartNumber)-1] = curPart
				}

				return true
			})
			if err != nil {
				return 0, errors.Wrap(err, "ListPartsPagesWithContext failed")
			}
			uploadID = *multiPartUploadsOutput.Uploads[0].UploadId
		} else {
			// TODO - remove all the multipart uploads for this key something is wonky, should only have one in-flight
		}
	}

	var moreBits bool = true
	var s3PartSize int64 = maxSegmentSize / this.rpetm.maxParts
	if s3PartSize < MinMultiPartSize {
		s3PartSize = MinMultiPartSize
	}

	var partNumber int64 = 0
	var bytesUploaded int64 = 0
	uploadBuffer := make([]byte, s3PartSize)
	for moreBits {
		log.Infof("Upload ongoing, Part: %d Bytes Uploaded: %d MB", partNumber, bytesUploaded/(1024*1024))
		// If the part has already been uploaded, we will skip
		uploadPart := completedParts[partNumber] == nil
		if uploadPart {
			bytesRead, err := reader.Read(uploadBuffer)
			if err != nil {
				if err == io.EOF {
					moreBits = false
				} else {
					return bytesUploaded, err
				}
			}
			thisPartNumber := partNumber + 1 // Copy because CompletedPart takes a pointer to the partNumber.
			// AWS part numbers start at 1, so we offset here
			uploadSlice := uploadBuffer[0:bytesRead]
			bufferReader := bytes.NewReader(uploadSlice)
			if partNumber == 0 && bytesRead < MinMultiPartSize {
				// We don't have enough data to do a multipart upload
				uploader := s3manager.NewUploader(&this.rpetm.session)

				result, err := uploader.UploadWithContext(ctx, &s3manager.UploadInput{
					Body:   bufferReader,
					Bucket: awsBucketName,
					Key:    awsKey,
				})
				if err == nil {
					log.Infof("Successfully uploaded to", result.Location)
				}

				if err != nil {
					return bytesUploaded, err
				}
				bytesUploaded += int64(bytesRead)
				moreBits = false
			} else {
				// Wait until here to start the multi-part upload in case we're too small
				if uploadID == "" {
					uploadInput := &s3.CreateMultipartUploadInput{
						Bucket: awsBucketName,
						Key:    awsKey,
					}
					resp, err := this.rpetm.s3.CreateMultipartUploadWithContext(ctx, uploadInput)
					if err != nil {
						return 0, err
					}
					uploadID = *resp.UploadId
				}

				partInput := &s3.UploadPartInput{
					Body:          bufferReader,
					Bucket:        awsBucketName,
					Key:           awsKey,
					PartNumber:    aws.Int64(thisPartNumber),
					UploadId:      &uploadID,
					ContentLength: aws.Int64(int64(bytesRead)),
				}
				partOutput, err := this.rpetm.s3.UploadPartWithContext(ctx, partInput)
				if err != nil {
					return bytesUploaded, err
				}

				log.Debug(partOutput)
				bytesRead64 := int64(bytesRead)
				completedPart := s3.Part{
					ETag:       partOutput.ETag,
					PartNumber: &thisPartNumber,
					Size:       &bytesRead64,
				}
				completedParts[partNumber] = &completedPart
			}
			bytesUploaded += int64(bytesRead)
		} else {
			log.Infof("Skipping part %d, found pre-existing part", partNumber)
			bytesSkipped, err := skipBytes(reader, *completedParts[partNumber].Size, uploadBuffer)
			if err == io.EOF {
				moreBits = false
			} else {
				return bytesUploaded, err
			}
			if bytesSkipped != bytesUploaded {
				return bytesUploaded, errors.Errorf("Did not skip correct number of bytes bytesSkipped: %d expected:%d", bytesSkipped, *completedParts[partNumber].Size)
			}
			bytesUploaded += *completedParts[partNumber].Size

		}
		partNumber++
		if bytesUploaded >= maxSegmentSize {
			moreBits = false
		}
	}

	// If we initiated or resumed a multipart upload, finish it here
	if uploadID != "" {
		parts := make([]*s3.CompletedPart, partNumber)
		for curPartNum, curPart := range completedParts[0:partNumber] {
			parts[curPartNum] = &s3.CompletedPart{
				ETag:       curPart.ETag,
				PartNumber: curPart.PartNumber, // This part number was already offset from 1
			}
		}
		completedInput := s3.CompleteMultipartUploadInput{
			Bucket: awsBucketName,
			Key:    awsKey,
			MultipartUpload: &s3.CompletedMultipartUpload{
				Parts: parts,
			},
			RequestPayer: nil,
			UploadId:     &uploadID,
		}
		completedOutput, err := this.rpetm.s3.CompleteMultipartUploadWithContext(ctx, &completedInput)
		if err != nil {
			return bytesUploaded, err
		}
		log.Print(completedOutput)
	}
	return bytesUploaded, err
}

func (this *ProtectedEntity) abortPendingMultipartUpload(ctx *context.Context, bucket *string, key *string) {
	log := this.rpetm.logger
	var combinedErrors []error
	if (*ctx).Err() != nil {
		log.Infof("The context was canceled for key: %v, proceeding with cleanup", *key)
		log.Infof("Processing pending multipart upload abort for key %v if present", *key)
		listUploadsInput := &s3.ListMultipartUploadsInput{
			Bucket: bucket,
			Prefix: key,
		}
		multiPartUploadsOutput, err := this.rpetm.s3.ListMultipartUploads(listUploadsInput)
		if err != nil {
			log.Errorf("Received error when retrieving pending multipart uploads for key %v during cleanup", *key)
			combinedErrors = append(combinedErrors, err)
			return
		}
		if len(multiPartUploadsOutput.Uploads) > 0 {
			log.Infof("Found %d pending multipart uploads for key: %v", len(multiPartUploadsOutput.Uploads), *key)
			for _, multiPartUpload := range multiPartUploadsOutput.Uploads {
				uploadId := *multiPartUpload.UploadId
				log.Infof("Found pending multipart upload for key: %v with upload-id: %v", *key, uploadId)
				abortMultipartUploadInput := s3.AbortMultipartUploadInput{
					Bucket:       bucket,
					Key:          key,
					RequestPayer: nil,
					UploadId:     &uploadId,
				}
				_, err := this.rpetm.s3.AbortMultipartUpload(&abortMultipartUploadInput)
				if err != nil {
					log.Errorf("Received error: %v when aborting pending multipart upload for key: %v with uploadId: %v during cleanup", err.Error(), *key, uploadId)
					combinedErrors = append(combinedErrors, err)
					continue
				}
				log.Infof("Successfully aborted the pending multipart upload for key: %v with uploadId: %v", *key, uploadId)
			}
		} else {
			log.Infof("No pending upload with key %v", *key)
		}
		if len(combinedErrors) > 0 {
			var combinedString string
			for _, curErr := range combinedErrors {
				combinedString += curErr.Error() + "\n"
			}
			errLog := errors.New("Multiple errors:\n" + combinedString)
			log.WithError(errLog).Errorf("Errors detected while aborting pending multi-part uploads.")
		}
	} else {
		log.Debugf("No abort detected for key: %v, no cleanup necessary.", *key)
	}
}

func (this *ProtectedEntity) copy(ctx context.Context, maxSegmentSize int64, dataReader io.Reader,
	metadataReader io.Reader) error {
	defer this.cleanupOnAbortedUpload(&ctx)
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
		err = this.uploadStream(ctx, dataName, maxSegmentSize, dataReader)
		if err != nil {
			return err
		}
	}

	if metadataReader != nil {
		mdName := this.rpetm.metadataName(peInfo.GetID())
		err = this.uploadStream(ctx, mdName, maxSegmentSize, metadataReader)
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
	_, err = this.rpetm.s3.PutObjectWithContext(ctx, jsonParams)
	if err != nil {
		return errors.Wrapf(err, "copy S3 PutObject for PE info failed for PE %s bucket %s key %s",
			peInfo.GetID(), this.rpetm.bucket, peinfoName)
	}
	return err
}

func (this *ProtectedEntity) cleanupOnAbortedUpload(ctx *context.Context) {
	log := this.rpetm.logger
	peInfo := this.peinfo
	if (*ctx).Err() != nil {
		log.Infof("The context was canceled during copy of pe %v, proceeding with cleanup", peInfo.GetName())
		log.Debugf("Attempting to delete any uploaded snapshots for %v", this.peinfo.GetID())
		// New context or else downstream "withContext" calls will error out.
		status, err := this.DeleteSnapshot(context.Background(), this.peinfo.GetID().GetSnapshotID())
		if err != nil {
			log.Errorf("Received error %v when deleting local snapshots of %v during cleanup", err.Error(), this.peinfo.GetID())
			return
		}
		if !status {
			log.Errorf("Failed in deleting local snapshots of %v during cleanup", this.peinfo.GetID())
			return
		}
		log.Infof("Successfully deleted any uploaded snapshots for %v present", this.peinfo.GetID())
	} else {
		log.Debugf("The PE: %v was uploaded successfully, no abort detected.", peInfo.GetName())
	}
}

func (this *ProtectedEntity) getReader(ctx context.Context, key string) (io.ReadCloser, error) {
	s3Segments, err := this.getS3Segments(ctx, this.rpetm.bucket, key)
	if err != nil {
		return nil, errors.Wrapf(err, "Could not get reader for bucket %s, key %s", this.rpetm.bucket, key)
	}
	segmentReader, err := newS3SegmentReader(this.rpetm.s3, s3Segments, this.rpetm.bucket)
	s3BufferedReader := bufio.NewReaderSize(&segmentReader, 1024*1024)
	return ioutil.NopCloser(s3BufferedReader), nil
}

type s3SegmentReader struct {
	s3                     s3.S3
	bucket                 string
	segments               []s3Segment
	offset                 int64
	curSegment             *s3Segment
	s3Reader               io.ReadCloser
	readerStart, readerEnd int64
}

func newS3SegmentReader(s3 s3.S3, s3Sgements []s3Segment, bucket string) (s3SegmentReader, error) {
	var nextStartOffset int64 = 0
	for segmentNum, checkSegment := range s3Sgements {
		if checkSegment.segmentNumber != segmentNum {
			return s3SegmentReader{}, errors.Errorf("Segments missing at segment %d/key %s", segmentNum, checkSegment.key)
		}
		if nextStartOffset != checkSegment.startOffset {
			return s3SegmentReader{}, errors.Errorf("Offsets do not match at segment %d/key %s, expecting %d, got %d", segmentNum, checkSegment.key, nextStartOffset, checkSegment.startOffset)
		}
		nextStartOffset += checkSegment.length
	}
	return s3SegmentReader{
		s3:       s3,
		bucket:   bucket,
		segments: s3Sgements,
		offset:   0,
	}, nil
}

func (this *s3SegmentReader) Read(p []byte) (n int, err error) {

	if this.s3Reader == nil {
		_, err := this.Seek(this.offset, io.SeekStart)
		if err != nil {
			return 0, err
		}
	}
	bytesRead, err := this.s3Reader.Read(p)
	if err == io.EOF {
		// Close out this reader, don't return EOF.  If this is the last segment, the Seek on the next read will
		// return EOF
		this.s3Reader.Close()
		this.s3Reader = nil
		this.curSegment = nil
		err = nil
	}
	if err != nil {
		return 0, err
	}
	this.offset += int64(bytesRead)
	return bytesRead, nil
}

func (this *s3SegmentReader) Seek(offset int64, whence int) (int64, error) {
	var absOffset int64
	switch whence {
	case io.SeekCurrent:
		absOffset = this.offset + offset
	case io.SeekEnd:
		return 0, errors.New("Not implemented")
	case io.SeekStart:
		absOffset = offset
	}
	if (absOffset < this.readerStart || absOffset > this.readerEnd) && this.s3Reader != nil {
		this.s3Reader.Close()
		this.s3Reader = nil
		this.curSegment = nil
	}

	if this.s3Reader == nil {
		for _, curSegment := range this.segments {
			if curSegment.startOffset <= absOffset && curSegment.startOffset+curSegment.length > absOffset {
				s3Object, err := this.s3.GetObject(&s3.GetObjectInput{
					Bucket: aws.String(this.bucket),
					Key:    aws.String(curSegment.key),
				})
				if err != nil {
					return 0, err
				}
				s3BufferedReader := bufio.NewReaderSize(s3Object.Body, 1024*1024)

				this.s3Reader = ioutil.NopCloser(s3BufferedReader)
				this.curSegment = &curSegment
				break
			}
		}
	}
	if this.curSegment == nil {
		return 0, io.EOF // Indexing past end of our segment list
	}
	segmentOffset := absOffset - this.curSegment.startOffset

	return skipBytes(this.s3Reader, segmentOffset, nil)
}
