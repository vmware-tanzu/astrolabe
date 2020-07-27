package client

import (
	"bufio"
	"context"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/pkg/errors"
	"github.com/vmware-tanzu/astrolabe/gen/client/operations"
	"github.com/vmware-tanzu/astrolabe/pkg/astrolabe"
	"io"
	"io/ioutil"
	"net/http"
	"time"
)

type ClientProtectedEntity struct {
	id   astrolabe.ProtectedEntityID
	petm *ClientProtectedEntityTypeManager
}

func NewClientProtectedEntity(id astrolabe.ProtectedEntityID, petm *ClientProtectedEntityTypeManager) ClientProtectedEntity {
	return ClientProtectedEntity{
		id:   id,
		petm: petm,
	}
}

func (this ClientProtectedEntity) GetInfo(ctx context.Context) (astrolabe.ProtectedEntityInfo, error) {
	params := operations.GetProtectedEntityInfoParams{
		Service:           this.petm.typeName,
		ProtectedEntityID: this.id.String(),
	}
	params.SetTimeout(time.Minute)
	getInfoOK, err := this.petm.entityManager.restClient.Operations.GetProtectedEntityInfo(&params)
	if err != nil {
		return astrolabe.ProtectedEntityInfoImpl{}, errors.Wrap(err, "Failed in GetProtectedEntityInfo")
	}
	return astrolabe.NewProtectedEntityInfoFromModel(getInfoOK.GetPayload())
}

func (this ClientProtectedEntity) GetCombinedInfo(ctx context.Context) ([]astrolabe.ProtectedEntityInfo, error) {

	panic("implement me")
}

func (this ClientProtectedEntity) Snapshot(ctx context.Context, snapshotParams map[string]map[string]interface{}) (astrolabe.ProtectedEntitySnapshotID, error) {
	createSnapshotParams := operations.CreateSnapshotParams{
		Service:           this.petm.typeName,
		ProtectedEntityID: this.id.String(),
	}
	createSnapshotParams.SetTimeout(time.Minute * 10)
	snapshotOK, err := this.petm.entityManager.restClient.Operations.CreateSnapshot(&createSnapshotParams)
	if err != nil {
		return astrolabe.ProtectedEntitySnapshotID{}, errors.Wrap(err, "Failed in CreateSnapshot")
	}
	return astrolabe.NewProtectedEntitySnapshotIDFromModel(snapshotOK.GetPayload()), nil
}

func (this ClientProtectedEntity) ListSnapshots(ctx context.Context) ([]astrolabe.ProtectedEntitySnapshotID, error) {
	params := operations.ListSnapshotsParams{
		ProtectedEntityID: this.id.String(),
		Service:           this.petm.typeName,
	}
	params.SetTimeout(time.Minute)
	listSnapshotsOK, err := this.petm.entityManager.restClient.Operations.ListSnapshots(&params)
	if err != nil {
		return nil, errors.Wrap(err, "Failed in ListSnapshots")
	}
	returnList := make([]astrolabe.ProtectedEntitySnapshotID, len(listSnapshotsOK.GetPayload().List))
	for curModelSnapshotIDNum, curModelSnapshotID := range listSnapshotsOK.GetPayload().List {
		curPEID, err := astrolabe.NewProtectedEntityIDFromModel(curModelSnapshotID)
		if err != nil {
			return nil, errors.Wrapf(err, "Faield to parse %v", curModelSnapshotID)
		}
		returnList[curModelSnapshotIDNum] = curPEID.GetSnapshotID()
	}
	return returnList, nil
}

func (this ClientProtectedEntity) DeleteSnapshot(ctx context.Context, snapshotToDelete astrolabe.ProtectedEntitySnapshotID) (bool, error) {
	panic("implement me")
}

func (this ClientProtectedEntity) GetInfoForSnapshot(ctx context.Context, snapshotID astrolabe.ProtectedEntitySnapshotID) (*astrolabe.ProtectedEntityInfo, error) {
	panic("implement me")
}

func (this ClientProtectedEntity) GetComponents(ctx context.Context) ([]astrolabe.ProtectedEntity, error) {
	peInfo, err := this.GetInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed in GetProtectedEntityInfo")
	}
	componentIDs := peInfo.GetComponentIDs()
	returnEntities := make([]astrolabe.ProtectedEntity, len(componentIDs))
	for curComponentIDNum, curComponentID := range componentIDs {
		returnEntities[curComponentIDNum], err = this.petm.entityManager.GetProtectedEntity(ctx, curComponentID)
		if err != nil {
			return nil, errors.Wrapf(err, "Failed in GetProtectedEntity for %s", curComponentID.String())
		}
	}
	return returnEntities, nil
}

func (this ClientProtectedEntity) GetID() astrolabe.ProtectedEntityID {
	return this.id
}

func (this ClientProtectedEntity) GetDataReader(ctx context.Context) (io.ReadCloser, error) {
	peInfo, err := this.GetInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed in GetProtectedEntityInfo")
	}
	transports := peInfo.GetDataTransports()
	return getBestReaderForTransports(ctx, transports)
}

func (this ClientProtectedEntity) GetMetadataReader(ctx context.Context) (io.ReadCloser, error) {
	peInfo, err := this.GetInfo(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Failed in GetProtectedEntityInfo")
	}
	transports := peInfo.GetMetadataTransports()
	return getBestReaderForTransports(ctx, transports)
}

func getBestReaderForTransports(ctx context.Context, transports []astrolabe.DataTransport) (io.ReadCloser, error) {
	for _, checkTransport := range transports {
		switch checkTransport.GetTransportType() {
		case astrolabe.S3TransportType:
			return getReaderForS3Transport(ctx, checkTransport)
		}
	}
	return nil, nil
}

func getReaderForS3Transport(ctx context.Context, s3Transport astrolabe.DataTransport) (io.ReadCloser, error) {
	session, err := session.NewSession(&aws.Config{
		Region: aws.String("us-west-1")},
	)
	if err != nil {
		return nil, errors.Wrapf(err, "AWS NewSession failed")
	}
	s3Client := s3.New(session)

	bucket, ok := s3Transport.GetParam(astrolabe.S3BucketParam)

	var reader io.ReadCloser
	if ok {
		key, hasKey := s3Transport.GetParam(astrolabe.S3KeyParam)
		if !hasKey {
			return nil, errors.New("Missing key param")
		}
		input := s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		}
		getObjectOutput, err := s3Client.GetObject(&input)
		if err != nil {
			return nil, errors.Wrapf(err, "S3 GetObject failed, GetObjectInput = %v", input)
		}

		reader = getObjectOutput.Body
	} else {
		url, hasURL := s3Transport.GetParam(astrolabe.S3URLParam)
		if !hasURL {
			return nil, errors.New("Missing url param")
		}
		response, err := http.Get(url)
		if err != nil {
			return nil, errors.Wrapf(err, "S3 GetObject failed, url = %s", url)
		}
		if response.StatusCode != http.StatusOK {
			return nil, errors.New(fmt.Sprintf("Get failed for S3 url %s, status = %s", url, response.Status))
		}
		reader = response.Body
	}

	s3BufferedReader := bufio.NewReaderSize(reader, 1024*1024)

	return ioutil.NopCloser(s3BufferedReader), nil
}
