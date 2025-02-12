package shpanstream

import (
	"encoding/json"
	"errors"
)

type Optional[T any] interface {
	AsStream() Stream[T]
	AsPtr() *T
	Get() (T, error)
	MustGet() T
	Present() bool
	OrElse(v T) T
	OrElseGet(alt func() T) T
	Or(alt Optional[T]) Optional[T]
	IfPresent(fn func(T))
	IsEmpty() bool
	Filter(predicate func(*T) bool) Optional[T]

	// Validate the validator function with the value if present, otherwise return nil
	Validate(validator func(T) error) error
}

// Optional - a generic optional type
type optional[T any] struct {
	value *T
}

// NewOptional creates a new optional with value
func NewOptional[T any](v T) Optional[T] {
	return &optional[T]{value: &v}
}

// EmptyOptional gets an empty optional
func EmptyOptional[T any]() Optional[T] {
	return NewOptionalFromPtr[T](nil)
}

// NewOptionalFromPtr creates an Optional from Ptr (or nil)
func NewOptionalFromPtr[T any](v *T) Optional[T] {
	return &optional[T]{value: v}
}

func (o *optional[T]) AsStream() Stream[T] {
	if o.Present() {
		return Just(*o.value)
	} else {
		return EmptyStream[T]()
	}
}

// AsPtr returns a *int of the value or nil if not present.
func (o *optional[T]) AsPtr() *T {
	return o.value
}

// Get returns the int value or an error if not present.
func (o *optional[T]) Get() (T, error) {
	if !o.Present() {
		var zero T
		return zero, errors.New("value not present")
	} else {
		return *o.value, nil
	}
}

// MustGet returns the int value or panics if not present.
func (o *optional[T]) MustGet() T {
	if !o.Present() {
		panic("value not present")
	}
	return *o.value
}

func (o *optional[T]) IsEmpty() bool {
	return o.value == nil
}

// Present returns whether we have a value or not.
func (o *optional[T]) Present() bool {
	return o.value != nil
}

// OrElse returns the int value or a default value if the value is not present.
func (o *optional[T]) OrElse(v T) T {
	if o.Present() {
		return *o.value
	}
	return v
}

func (o *optional[T]) Filter(predicate func(*T) bool) Optional[T] {
	if o.Present() && predicate(o.value) {
		return o
	}
	return EmptyOptional[T]()
}

func MapOptional[SRC any, TGT any](src Optional[SRC], mapper func(SRC) TGT) Optional[TGT] {
	srcPtr := src.AsPtr()
	if srcPtr != nil {
		return NewOptional[TGT](mapper(*srcPtr))
	} else {
		return EmptyOptional[TGT]()
	}
}
func MapOptionalWithError[SRC any, TGT any](src Optional[SRC], mapper func(SRC) (TGT, error)) (Optional[TGT], error) {
	srcPtr := src.AsPtr()
	if srcPtr != nil {
		tgt, err := mapper(*srcPtr)
		if err != nil {
			return nil, err
		}
		return NewOptional[TGT](tgt), nil
	} else {
		return EmptyOptional[TGT](), nil
	}
}

// OrElseGet returns the int value or a default value if the value is not present.
func (o *optional[T]) OrElseGet(alt func() T) T {
	if o.Present() {
		return *o.value
	}
	return alt()
}

func (o *optional[T]) Validate(validator func(T) error) error {
	if o.Present() {
		return validator(*o.value)
	}
	return nil
}

func (o *optional[T]) Or(alt Optional[T]) Optional[T] {
	if o.Present() {
		return o
	} else {
		return alt
	}
}

// IfPresent calls the function f with the value if the value is present.
func (o *optional[T]) IfPresent(fn func(T)) {
	if o.Present() {
		fn(*o.value)
	}
}

func (o *optional[T]) MarshalJSON() ([]byte, error) {
	if o.Present() {
		return json.Marshal(o.value)
	}
	return json.Marshal(nil)
}

func (o *optional[T]) UnmarshalJSON(data []byte) error {

	if string(data) == "null" {
		o.value = nil
		return nil
	}

	var value T

	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	o.value = &value
	return nil
}
