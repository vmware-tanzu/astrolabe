// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/runtime/middleware"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// NewGetAstrolabeTasksNexusTaskNexusIDParams creates a new GetAstrolabeTasksNexusTaskNexusIDParams object
//
// There are no default values defined in the spec.
func NewGetAstrolabeTasksNexusTaskNexusIDParams() GetAstrolabeTasksNexusTaskNexusIDParams {

	return GetAstrolabeTasksNexusTaskNexusIDParams{}
}

// GetAstrolabeTasksNexusTaskNexusIDParams contains all the bound params for the get astrolabe tasks nexus task nexus ID operation
// typically these are obtained from a http.Request
//
// swagger:parameters GetAstrolabeTasksNexusTaskNexusID
type GetAstrolabeTasksNexusTaskNexusIDParams struct {

	// HTTP Request Object
	HTTPRequest *http.Request `json:"-"`

	/*Last finished time seen by this client.  Tasks that have completed after this time tick will be returned, or if no tasks
	have finished, the call will hang until waitTime has passed or a task finishes.  Starting time tick should
	be the finished time of the last task that the caller saw completed on this nexus.  Use 0 to get all finished
	tasks (tasks that have finished and timed out of the server will not be shown)

	  Required: true
	  In: query
	*/
	LastFinishedNS int64
	/*The nexus to wait on
	  Required: true
	  In: path
	*/
	TaskNexusID string
	/*Time to wait (milliseconds) before returning if no tasks   complete
	  Required: true
	  In: query
	*/
	WaitTime int64
}

// BindRequest both binds and validates a request, it assumes that complex things implement a Validatable(strfmt.Registry) error interface
// for simple values it will use straight method calls.
//
// To ensure default values, the struct must have been initialized with NewGetAstrolabeTasksNexusTaskNexusIDParams() beforehand.
func (o *GetAstrolabeTasksNexusTaskNexusIDParams) BindRequest(r *http.Request, route *middleware.MatchedRoute) error {
	var res []error

	o.HTTPRequest = r

	qs := runtime.Values(r.URL.Query())

	qLastFinishedNS, qhkLastFinishedNS, _ := qs.GetOK("lastFinishedNS")
	if err := o.bindLastFinishedNS(qLastFinishedNS, qhkLastFinishedNS, route.Formats); err != nil {
		res = append(res, err)
	}

	rTaskNexusID, rhkTaskNexusID, _ := route.Params.GetOK("taskNexusID")
	if err := o.bindTaskNexusID(rTaskNexusID, rhkTaskNexusID, route.Formats); err != nil {
		res = append(res, err)
	}

	qWaitTime, qhkWaitTime, _ := qs.GetOK("waitTime")
	if err := o.bindWaitTime(qWaitTime, qhkWaitTime, route.Formats); err != nil {
		res = append(res, err)
	}
	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// bindLastFinishedNS binds and validates parameter LastFinishedNS from query.
func (o *GetAstrolabeTasksNexusTaskNexusIDParams) bindLastFinishedNS(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("lastFinishedNS", "query", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// AllowEmptyValue: false

	if err := validate.RequiredString("lastFinishedNS", "query", raw); err != nil {
		return err
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("lastFinishedNS", "query", "int64", raw)
	}
	o.LastFinishedNS = value

	return nil
}

// bindTaskNexusID binds and validates parameter TaskNexusID from path.
func (o *GetAstrolabeTasksNexusTaskNexusIDParams) bindTaskNexusID(rawData []string, hasKey bool, formats strfmt.Registry) error {
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// Parameter is provided by construction from the route
	o.TaskNexusID = raw

	return nil
}

// bindWaitTime binds and validates parameter WaitTime from query.
func (o *GetAstrolabeTasksNexusTaskNexusIDParams) bindWaitTime(rawData []string, hasKey bool, formats strfmt.Registry) error {
	if !hasKey {
		return errors.Required("waitTime", "query", rawData)
	}
	var raw string
	if len(rawData) > 0 {
		raw = rawData[len(rawData)-1]
	}

	// Required: true
	// AllowEmptyValue: false

	if err := validate.RequiredString("waitTime", "query", raw); err != nil {
		return err
	}

	value, err := swag.ConvertInt64(raw)
	if err != nil {
		return errors.InvalidType("waitTime", "query", "int64", raw)
	}
	o.WaitTime = value

	return nil
}
