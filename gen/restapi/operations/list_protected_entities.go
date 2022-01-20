// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the generate command

import (
	"net/http"

	"github.com/go-openapi/runtime/middleware"
)

// ListProtectedEntitiesHandlerFunc turns a function with the right signature into a list protected entities handler
type ListProtectedEntitiesHandlerFunc func(ListProtectedEntitiesParams) middleware.Responder

// Handle executing the request and returning a response
func (fn ListProtectedEntitiesHandlerFunc) Handle(params ListProtectedEntitiesParams) middleware.Responder {
	return fn(params)
}

// ListProtectedEntitiesHandler interface for that can handle valid list protected entities params
type ListProtectedEntitiesHandler interface {
	Handle(ListProtectedEntitiesParams) middleware.Responder
}

// NewListProtectedEntities creates a new http.Handler for the list protected entities operation
func NewListProtectedEntities(ctx *middleware.Context, handler ListProtectedEntitiesHandler) *ListProtectedEntities {
	return &ListProtectedEntities{Context: ctx, Handler: handler}
}

/* ListProtectedEntities swagger:route GET /astrolabe/{service} listProtectedEntities

List protected entities for the service.  Results will be returned in
canonical ID order (string sorted).  Fewer results may be returned than
expected, the ProtectedEntityList has a field specifying if the list has
been truncated.


*/
type ListProtectedEntities struct {
	Context *middleware.Context
	Handler ListProtectedEntitiesHandler
}

func (o *ListProtectedEntities) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, rCtx, _ := o.Context.RouteInfo(r)
	if rCtx != nil {
		*r = *rCtx
	}
	var Params = NewListProtectedEntitiesParams()
	if err := o.Context.BindValidRequest(r, route, &Params); err != nil { // bind params
		o.Context.Respond(rw, r, route.Produces, route, err)
		return
	}

	res := o.Handler.Handle(Params) // actually handle the request
	o.Context.Respond(rw, r, route.Produces, route, res)

}
