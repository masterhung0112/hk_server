package store

import (
	"fmt"
)

// ErrInvalidInput indicates an error that has occured due to an invalid input.
type ErrInvalidInput struct {
	Entity string      // The entity which was sent as the input.
	Field  string      // The field of the entity which was invalid.
	Value  interface{} // The actual value of the field.
}

func NewErrInvalidInput(entity, field string, value interface{}) *ErrInvalidInput {
	return &ErrInvalidInput{
		Entity: entity,
		Field:  field,
		Value:  value,
	}
}

func (e *ErrInvalidInput) Error() string {
	return fmt.Sprintf("invalid input: entity: %s field: %s value: %s", e.Entity, e.Field, e.Value)
}

// ErrNotFound indicates that a resource was not found
type ErrNotFound struct {
	resource string
	Id       string
}

func NewErrNotFound(resource, id string) *ErrNotFound {
	return &ErrNotFound{
		resource: resource,
		Id:       id,
	}
}

func (e *ErrNotFound) Error() string {
	return "resource: " + e.resource + " id: " + e.Id
}