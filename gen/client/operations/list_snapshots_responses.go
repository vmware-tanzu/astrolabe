// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/vmware-tanzu/astrolabe/gen/models"
)

// ListSnapshotsReader is a Reader for the ListSnapshots structure.
type ListSnapshotsReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListSnapshotsReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListSnapshotsOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewListSnapshotsNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewListSnapshotsOK creates a ListSnapshotsOK with default headers values
func NewListSnapshotsOK() *ListSnapshotsOK {
	return &ListSnapshotsOK{}
}

/* ListSnapshotsOK describes a response with status code 200, with default header values.

List succeeded
*/
type ListSnapshotsOK struct {
	Payload *models.ProtectedEntityList
}

func (o *ListSnapshotsOK) Error() string {
	return fmt.Sprintf("[GET /astrolabe/{service}/{protectedEntityID}/snapshots][%d] listSnapshotsOK  %+v", 200, o.Payload)
}
func (o *ListSnapshotsOK) GetPayload() *models.ProtectedEntityList {
	return o.Payload
}

func (o *ListSnapshotsOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ProtectedEntityList)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListSnapshotsNotFound creates a ListSnapshotsNotFound with default headers values
func NewListSnapshotsNotFound() *ListSnapshotsNotFound {
	return &ListSnapshotsNotFound{}
}

/* ListSnapshotsNotFound describes a response with status code 404, with default header values.

Service or Protected Entity not found
*/
type ListSnapshotsNotFound struct {
}

func (o *ListSnapshotsNotFound) Error() string {
	return fmt.Sprintf("[GET /astrolabe/{service}/{protectedEntityID}/snapshots][%d] listSnapshotsNotFound ", 404)
}

func (o *ListSnapshotsNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}
