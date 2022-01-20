// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewGetTaskInfoParams creates a new GetTaskInfoParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewGetTaskInfoParams() *GetTaskInfoParams {
	return &GetTaskInfoParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewGetTaskInfoParamsWithTimeout creates a new GetTaskInfoParams object
// with the ability to set a timeout on a request.
func NewGetTaskInfoParamsWithTimeout(timeout time.Duration) *GetTaskInfoParams {
	return &GetTaskInfoParams{
		timeout: timeout,
	}
}

// NewGetTaskInfoParamsWithContext creates a new GetTaskInfoParams object
// with the ability to set a context for a request.
func NewGetTaskInfoParamsWithContext(ctx context.Context) *GetTaskInfoParams {
	return &GetTaskInfoParams{
		Context: ctx,
	}
}

// NewGetTaskInfoParamsWithHTTPClient creates a new GetTaskInfoParams object
// with the ability to set a custom HTTPClient for a request.
func NewGetTaskInfoParamsWithHTTPClient(client *http.Client) *GetTaskInfoParams {
	return &GetTaskInfoParams{
		HTTPClient: client,
	}
}

/* GetTaskInfoParams contains all the parameters to send to the API endpoint
   for the get task info operation.

   Typically these are written to a http.Request.
*/
type GetTaskInfoParams struct {

	/* TaskID.

	   The ID of the task to retrieve info for
	*/
	TaskID string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the get task info params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTaskInfoParams) WithDefaults() *GetTaskInfoParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the get task info params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *GetTaskInfoParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the get task info params
func (o *GetTaskInfoParams) WithTimeout(timeout time.Duration) *GetTaskInfoParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the get task info params
func (o *GetTaskInfoParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the get task info params
func (o *GetTaskInfoParams) WithContext(ctx context.Context) *GetTaskInfoParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the get task info params
func (o *GetTaskInfoParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the get task info params
func (o *GetTaskInfoParams) WithHTTPClient(client *http.Client) *GetTaskInfoParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the get task info params
func (o *GetTaskInfoParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithTaskID adds the taskID to the get task info params
func (o *GetTaskInfoParams) WithTaskID(taskID string) *GetTaskInfoParams {
	o.SetTaskID(taskID)
	return o
}

// SetTaskID adds the taskId to the get task info params
func (o *GetTaskInfoParams) SetTaskID(taskID string) {
	o.TaskID = taskID
}

// WriteToRequest writes these params to a swagger request
func (o *GetTaskInfoParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param taskID
	if err := r.SetPathParam("taskID", o.TaskID); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
