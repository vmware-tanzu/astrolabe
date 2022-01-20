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

// PostAstrolabeTasksNexusReader is a Reader for the PostAstrolabeTasksNexus structure.
type PostAstrolabeTasksNexusReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PostAstrolabeTasksNexusReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPostAstrolabeTasksNexusOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		return nil, runtime.NewAPIError("response status code does not match any response statuses defined for this endpoint in the swagger spec", response, response.Code())
	}
}

// NewPostAstrolabeTasksNexusOK creates a PostAstrolabeTasksNexusOK with default headers values
func NewPostAstrolabeTasksNexusOK() *PostAstrolabeTasksNexusOK {
	return &PostAstrolabeTasksNexusOK{}
}

/* PostAstrolabeTasksNexusOK describes a response with status code 200, with default header values.

New task nexus
*/
type PostAstrolabeTasksNexusOK struct {
	Payload models.TaskNexusID
}

func (o *PostAstrolabeTasksNexusOK) Error() string {
	return fmt.Sprintf("[POST /astrolabe/tasks/nexus][%d] postAstrolabeTasksNexusOK  %+v", 200, o.Payload)
}
func (o *PostAstrolabeTasksNexusOK) GetPayload() models.TaskNexusID {
	return o.Payload
}

func (o *PostAstrolabeTasksNexusOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
