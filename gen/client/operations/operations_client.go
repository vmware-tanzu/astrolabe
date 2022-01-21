// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
)

// New creates a new operations API client.
func New(transport runtime.ClientTransport, formats strfmt.Registry) ClientService {
	return &Client{transport: transport, formats: formats}
}

/*
Client for operations API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

// ClientOption is the option for Client methods
type ClientOption func(*runtime.ClientOperation)

// ClientService is the interface for Client methods
type ClientService interface {
	GetAstrolabeTasksNexusTaskNexusID(params *GetAstrolabeTasksNexusTaskNexusIDParams, opts ...ClientOption) (*GetAstrolabeTasksNexusTaskNexusIDOK, error)

	PostAstrolabeTasksNexus(params *PostAstrolabeTasksNexusParams, opts ...ClientOption) (*PostAstrolabeTasksNexusOK, error)

	CopyProtectedEntity(params *CopyProtectedEntityParams, opts ...ClientOption) (*CopyProtectedEntityAccepted, error)

	CreateSnapshot(params *CreateSnapshotParams, opts ...ClientOption) (*CreateSnapshotOK, error)

	DeleteProtectedEntity(params *DeleteProtectedEntityParams, opts ...ClientOption) (*DeleteProtectedEntityOK, error)

	GetProtectedEntityInfo(params *GetProtectedEntityInfoParams, opts ...ClientOption) (*GetProtectedEntityInfoOK, error)

	GetTaskInfo(params *GetTaskInfoParams, opts ...ClientOption) (*GetTaskInfoOK, error)

	ListProtectedEntities(params *ListProtectedEntitiesParams, opts ...ClientOption) (*ListProtectedEntitiesOK, error)

	ListServices(params *ListServicesParams, opts ...ClientOption) (*ListServicesOK, error)

	ListSnapshots(params *ListSnapshotsParams, opts ...ClientOption) (*ListSnapshotsOK, error)

	ListTaskNexus(params *ListTaskNexusParams, opts ...ClientOption) (*ListTaskNexusOK, error)

	ListTasks(params *ListTasksParams, opts ...ClientOption) (*ListTasksOK, error)

	SetTransport(transport runtime.ClientTransport)
}

/*
  GetAstrolabeTasksNexusTaskNexusID get astrolabe tasks nexus task nexus ID API
*/
func (a *Client) GetAstrolabeTasksNexusTaskNexusID(params *GetAstrolabeTasksNexusTaskNexusIDParams, opts ...ClientOption) (*GetAstrolabeTasksNexusTaskNexusIDOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetAstrolabeTasksNexusTaskNexusIDParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "GetAstrolabeTasksNexusTaskNexusID",
		Method:             "GET",
		PathPattern:        "/astrolabe/tasks/nexus/{taskNexusID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &GetAstrolabeTasksNexusTaskNexusIDReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetAstrolabeTasksNexusTaskNexusIDOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for GetAstrolabeTasksNexusTaskNexusID: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  PostAstrolabeTasksNexus Creates a new nexus for monitoring task completion
*/
func (a *Client) PostAstrolabeTasksNexus(params *PostAstrolabeTasksNexusParams, opts ...ClientOption) (*PostAstrolabeTasksNexusOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPostAstrolabeTasksNexusParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "PostAstrolabeTasksNexus",
		Method:             "POST",
		PathPattern:        "/astrolabe/tasks/nexus",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &PostAstrolabeTasksNexusReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PostAstrolabeTasksNexusOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for PostAstrolabeTasksNexus: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  CopyProtectedEntity Copy a protected entity into the repository.  There is no option to
embed data on this path, for a self-contained or partially
self-contained object, use the restore from zip file option in the S3
API REST API

*/
func (a *Client) CopyProtectedEntity(params *CopyProtectedEntityParams, opts ...ClientOption) (*CopyProtectedEntityAccepted, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCopyProtectedEntityParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "copyProtectedEntity",
		Method:             "POST",
		PathPattern:        "/astrolabe/{service}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &CopyProtectedEntityReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CopyProtectedEntityAccepted)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for copyProtectedEntity: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  CreateSnapshot Creates a new snapshot for this protected entity

*/
func (a *Client) CreateSnapshot(params *CreateSnapshotParams, opts ...ClientOption) (*CreateSnapshotOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewCreateSnapshotParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "createSnapshot",
		Method:             "POST",
		PathPattern:        "/astrolabe/{service}/{protectedEntityID}/snapshots",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &CreateSnapshotReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*CreateSnapshotOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for createSnapshot: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  DeleteProtectedEntity Deletes a protected entity or snapshot of a protected entity (if the
snapshot ID is specified)

*/
func (a *Client) DeleteProtectedEntity(params *DeleteProtectedEntityParams, opts ...ClientOption) (*DeleteProtectedEntityOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewDeleteProtectedEntityParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "deleteProtectedEntity",
		Method:             "DELETE",
		PathPattern:        "/astrolabe/{service}/{protectedEntityID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &DeleteProtectedEntityReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*DeleteProtectedEntityOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for deleteProtectedEntity: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  GetProtectedEntityInfo Get the info for a Protected Entity including name, data access and
components

*/
func (a *Client) GetProtectedEntityInfo(params *GetProtectedEntityInfoParams, opts ...ClientOption) (*GetProtectedEntityInfoOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetProtectedEntityInfoParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getProtectedEntityInfo",
		Method:             "GET",
		PathPattern:        "/astrolabe/{service}/{protectedEntityID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &GetProtectedEntityInfoReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetProtectedEntityInfoOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getProtectedEntityInfo: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  GetTaskInfo gets info about a running or recently completed task
*/
func (a *Client) GetTaskInfo(params *GetTaskInfoParams, opts ...ClientOption) (*GetTaskInfoOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewGetTaskInfoParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "getTaskInfo",
		Method:             "GET",
		PathPattern:        "/astrolabe/tasks/{taskID}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &GetTaskInfoReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*GetTaskInfoOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for getTaskInfo: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ListProtectedEntities List protected entities for the service.  Results will be returned in
canonical ID order (string sorted).  Fewer results may be returned than
expected, the ProtectedEntityList has a field specifying if the list has
been truncated.

*/
func (a *Client) ListProtectedEntities(params *ListProtectedEntitiesParams, opts ...ClientOption) (*ListProtectedEntitiesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListProtectedEntitiesParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listProtectedEntities",
		Method:             "GET",
		PathPattern:        "/astrolabe/{service}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &ListProtectedEntitiesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListProtectedEntitiesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listProtectedEntities: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ListServices lists available services

  This returns the list of services that this Astrolabe server supports

*/
func (a *Client) ListServices(params *ListServicesParams, opts ...ClientOption) (*ListServicesOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListServicesParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listServices",
		Method:             "GET",
		PathPattern:        "/astrolabe",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &ListServicesReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListServicesOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listServices: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ListSnapshots Gets the list of snapshots for this protected entity

*/
func (a *Client) ListSnapshots(params *ListSnapshotsParams, opts ...ClientOption) (*ListSnapshotsOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListSnapshotsParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listSnapshots",
		Method:             "GET",
		PathPattern:        "/astrolabe/{service}/{protectedEntityID}/snapshots",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &ListSnapshotsReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListSnapshotsOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listSnapshots: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ListTaskNexus Provides a list of current task nexus
*/
func (a *Client) ListTaskNexus(params *ListTaskNexusParams, opts ...ClientOption) (*ListTaskNexusOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListTaskNexusParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listTaskNexus",
		Method:             "GET",
		PathPattern:        "/astrolabe/tasks/nexus",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &ListTaskNexusReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListTaskNexusOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listTaskNexus: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
  ListTasks Lists running and recent tasks
*/
func (a *Client) ListTasks(params *ListTasksParams, opts ...ClientOption) (*ListTasksOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewListTasksParams()
	}
	op := &runtime.ClientOperation{
		ID:                 "listTasks",
		Method:             "GET",
		PathPattern:        "/astrolabe/tasks",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"application/json"},
		Schemes:            []string{"https"},
		Params:             params,
		Reader:             &ListTasksReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}
	for _, opt := range opts {
		opt(op)
	}

	result, err := a.transport.Submit(op)
	if err != nil {
		return nil, err
	}
	success, ok := result.(*ListTasksOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for listTasks: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

// SetTransport changes the transport on the client
func (a *Client) SetTransport(transport runtime.ClientTransport) {
	a.transport = transport
}
