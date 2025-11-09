package opt

import "encoding/json"

// Opt is a lightweight option type used throughout the generated APIs.
// The zero value represents an unset option.
type Opt[T any] struct {
	v   T
	set bool
}

// OVal wraps a value in an Opt marked as present.
func OVal[T any](v T) Opt[T] {
	return Opt[T]{v: v, set: true}
}

// ONil returns an Opt marked as not set.
func ONil[T any]() Opt[T] {
	var zero T
	return Opt[T]{v: zero, set: false}
}

// IsSome reports whether the option has been set explicitly.
func (o Opt[T]) IsSome() bool {
	return o.set
}

// Value returns the stored value regardless of whether it is set.
// Callers should check IsSome before using the value if presence matters.
func (o Opt[T]) Value() T {
	return o.v
}

// On returns an Opt[bool] set to true.
func On() Opt[bool] {
	return OVal(true)
}

// Off returns an Opt[bool] set to false.
func Off() Opt[bool] {
	return OVal(false)
}

// MarshalJSON encodes the option as either null (unset) or the underlying value.
func (o Opt[T]) MarshalJSON() ([]byte, error) {
	if !o.set {
		return []byte("null"), nil
	}
	return json.Marshal(o.v)
}

// UnmarshalJSON decodes the option from JSON, marking it as set unless the value is null.
func (o *Opt[T]) UnmarshalJSON(data []byte) error {
	if o == nil {
		return nil
	}
	if string(data) == "null" {
		var zero T
		o.v = zero
		o.set = false
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.v = v
	o.set = true
	return nil
}
