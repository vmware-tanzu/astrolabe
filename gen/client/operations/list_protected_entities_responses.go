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

// ListProtectedEntitiesReader is a Reader for the ListProtectedEntities structure.
type ListProtectedEntitiesReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *ListProtectedEntitiesReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewListProtectedEntitiesOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewListProtectedEntitiesNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	case 500:
		result := NewListProtectedEntitiesInternalServerError()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewListProtectedEntitiesOK creates a ListProtectedEntitiesOK with default headers values
func NewListProtectedEntitiesOK() *ListProtectedEntitiesOK {
	return &ListProtectedEntitiesOK{}
}

/*ListProtectedEntitiesOK handles this case with default header values.

200 response
*/
type ListProtectedEntitiesOK struct {
	Payload *models.ProtectedEntityList
}

func (o *ListProtectedEntitiesOK) Error() string {
	return fmt.Sprintf("[GET /astrolabe/{service}][%d] listProtectedEntitiesOK  %+v", 200, o.Payload)
}

func (o *ListProtectedEntitiesOK) GetPayload() *models.ProtectedEntityList {
	return o.Payload
}

func (o *ListProtectedEntitiesOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ProtectedEntityList)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListProtectedEntitiesNotFound creates a ListProtectedEntitiesNotFound with default headers values
func NewListProtectedEntitiesNotFound() *ListProtectedEntitiesNotFound {
	return &ListProtectedEntitiesNotFound{}
}

/*ListProtectedEntitiesNotFound handles this case with default header values.

404 response
*/
type ListProtectedEntitiesNotFound struct {
	Payload *models.NotFoundError
}

func (o *ListProtectedEntitiesNotFound) Error() string {
	return fmt.Sprintf("[GET /astrolabe/{service}][%d] listProtectedEntitiesNotFound  %+v", 404, o.Payload)
}

func (o *ListProtectedEntitiesNotFound) GetPayload() *models.NotFoundError {
	return o.Payload
}

func (o *ListProtectedEntitiesNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.NotFoundError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewListProtectedEntitiesInternalServerError creates a ListProtectedEntitiesInternalServerError with default headers values
func NewListProtectedEntitiesInternalServerError() *ListProtectedEntitiesInternalServerError {
	return &ListProtectedEntitiesInternalServerError{}
}

/*ListProtectedEntitiesInternalServerError handles this case with default header values.

500 response
*/
type ListProtectedEntitiesInternalServerError struct {
	Payload *models.ServerError
}

func (o *ListProtectedEntitiesInternalServerError) Error() string {
	return fmt.Sprintf("[GET /astrolabe/{service}][%d] listProtectedEntitiesInternalServerError  %+v", 500, o.Payload)
}

func (o *ListProtectedEntitiesInternalServerError) GetPayload() *models.ServerError {
	return o.Payload
}

func (o *ListProtectedEntitiesInternalServerError) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ServerError)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
