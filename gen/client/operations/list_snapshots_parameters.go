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

// NewListSnapshotsParams creates a new ListSnapshotsParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewListSnapshotsParams() *ListSnapshotsParams {
	return &ListSnapshotsParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewListSnapshotsParamsWithTimeout creates a new ListSnapshotsParams object
// with the ability to set a timeout on a request.
func NewListSnapshotsParamsWithTimeout(timeout time.Duration) *ListSnapshotsParams {
	return &ListSnapshotsParams{
		timeout: timeout,
	}
}

// NewListSnapshotsParamsWithContext creates a new ListSnapshotsParams object
// with the ability to set a context for a request.
func NewListSnapshotsParamsWithContext(ctx context.Context) *ListSnapshotsParams {
	return &ListSnapshotsParams{
		Context: ctx,
	}
}

// NewListSnapshotsParamsWithHTTPClient creates a new ListSnapshotsParams object
// with the ability to set a custom HTTPClient for a request.
func NewListSnapshotsParamsWithHTTPClient(client *http.Client) *ListSnapshotsParams {
	return &ListSnapshotsParams{
		HTTPClient: client,
	}
}

/* ListSnapshotsParams contains all the parameters to send to the API endpoint
   for the list snapshots operation.

   Typically these are written to a http.Request.
*/
type ListSnapshotsParams struct {

	/* ProtectedEntityID.

	   The protected entity ID to retrieve info for
	*/
	ProtectedEntityID string

	/* Service.

	   The service for the protected entity
	*/
	Service string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the list snapshots params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListSnapshotsParams) WithDefaults() *ListSnapshotsParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the list snapshots params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ListSnapshotsParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the list snapshots params
func (o *ListSnapshotsParams) WithTimeout(timeout time.Duration) *ListSnapshotsParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the list snapshots params
func (o *ListSnapshotsParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the list snapshots params
func (o *ListSnapshotsParams) WithContext(ctx context.Context) *ListSnapshotsParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the list snapshots params
func (o *ListSnapshotsParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the list snapshots params
func (o *ListSnapshotsParams) WithHTTPClient(client *http.Client) *ListSnapshotsParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the list snapshots params
func (o *ListSnapshotsParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithProtectedEntityID adds the protectedEntityID to the list snapshots params
func (o *ListSnapshotsParams) WithProtectedEntityID(protectedEntityID string) *ListSnapshotsParams {
	o.SetProtectedEntityID(protectedEntityID)
	return o
}

// SetProtectedEntityID adds the protectedEntityId to the list snapshots params
func (o *ListSnapshotsParams) SetProtectedEntityID(protectedEntityID string) {
	o.ProtectedEntityID = protectedEntityID
}

// WithService adds the service to the list snapshots params
func (o *ListSnapshotsParams) WithService(service string) *ListSnapshotsParams {
	o.SetService(service)
	return o
}

// SetService adds the service to the list snapshots params
func (o *ListSnapshotsParams) SetService(service string) {
	o.Service = service
}

// WriteToRequest writes these params to a swagger request
func (o *ListSnapshotsParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param protectedEntityID
	if err := r.SetPathParam("protectedEntityID", o.ProtectedEntityID); err != nil {
		return err
	}

	// path param service
	if err := r.SetPathParam("service", o.Service); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
