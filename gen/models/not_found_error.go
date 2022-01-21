// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// NotFoundError not found error
//
// swagger:model NotFoundError
type NotFoundError struct {
	Error

	// entity not found
	Field string `json:"field,omitempty"`
}

// UnmarshalJSON unmarshals this object from a JSON structure
func (m *NotFoundError) UnmarshalJSON(raw []byte) error {
	// AO0
	var aO0 Error
	if err := swag.ReadJSON(raw, &aO0); err != nil {
		return err
	}
	m.Error = aO0

	// AO1
	var dataAO1 struct {
		Field string `json:"field,omitempty"`
	}
	if err := swag.ReadJSON(raw, &dataAO1); err != nil {
		return err
	}

	m.Field = dataAO1.Field

	return nil
}

// MarshalJSON marshals this object to a JSON structure
func (m NotFoundError) MarshalJSON() ([]byte, error) {
	_parts := make([][]byte, 0, 2)

	aO0, err := swag.WriteJSON(m.Error)
	if err != nil {
		return nil, err
	}
	_parts = append(_parts, aO0)
	var dataAO1 struct {
		Field string `json:"field,omitempty"`
	}

	dataAO1.Field = m.Field

	jsonDataAO1, errAO1 := swag.WriteJSON(dataAO1)
	if errAO1 != nil {
		return nil, errAO1
	}
	_parts = append(_parts, jsonDataAO1)
	return swag.ConcatJSON(_parts...), nil
}

// Validate validates this not found error
func (m *NotFoundError) Validate(formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with Error
	if err := m.Error.Validate(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// ContextValidate validate this not found error based on the context it is used
func (m *NotFoundError) ContextValidate(ctx context.Context, formats strfmt.Registry) error {
	var res []error

	// validation for a type composition with Error
	if err := m.Error.ContextValidate(ctx, formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

// MarshalBinary interface implementation
func (m *NotFoundError) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *NotFoundError) UnmarshalBinary(b []byte) error {
	var res NotFoundError
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
