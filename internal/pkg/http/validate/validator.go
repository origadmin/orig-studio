// Package validate provides a framework-agnostic, protobuf-first input
// validation entry point used by all three HTTP adapters (gin, kratos,
// std). Handlers call ctx.BindJSON(req) (which decodes JSON), and each
// adapter invokes validate.Validate(req) after decoding.
//
// Design goals:
//   - ZERO framework dependencies (no gin/kratos). The package only
//     references interface{} and the two duck-typed interfaces:
//     Validatable (ValidateAll error) and ValidatableFirst (Validate error).
//   - No-op for non-proto structs. If a Go struct does not implement the
//     validate methods, Validate returns nil. This allows a progressive
//     migration: handlers still using inline Go structs (Phase 2-3 later)
//     keep working, and proto messages with buf.validate rules get
//     automatic validation.
//   - Error output is a single string compatible with the existing
//     server.ErrBadRequest style: "field1: reason1; field2: reason2".
//     The gin adapter before migration used go-playground/validator which
//     produced a similar concatenated string, so behaviour parity is
//     preserved for callers expecting a 400 with a compound message.
package validate

import (
	"errors"
	"fmt"
	"strings"
)

// Validatable is the duck-typed interface implemented by every protobuf
// message that was compiled with protoc-gen-validate (envoyproxy) AND has
// at least one (validate.rules) annotation. It exposes the exhaustive
// validator which returns ALL violations at once.
type Validatable interface {
	ValidateAll() error
}

// ValidatableFirst is the optional, faster variant that returns only the
// first violation. We use ValidateAll by default so handlers receive the
// full list of problems in one 400 response (matches gin binding output
// that enumerates every invalid field together).
type ValidatableFirst interface {
	Validate() error
}

// ValidationError describes a single field-level validation error. All
// protoc-gen-validate generated types expose a compatible struct that has
// Field()/Reason() methods. We declare the interface here so that the
// adapter can flatten violations into a human-readable string without
// importing generated packages.
type ValidationError interface {
	error
	Field() string
	Reason() string
	Cause() error
}

// multiError is the common interface implemented by all the generated
// *MultiError slice wrapper types (e.g. LoginRequestMultiError). It lets
// us unwrap a composite error into individual field errors.
type multiError interface {
	error
	AllErrors() []error
}

// Validate validates req using protovalidate/protoc-gen-validate semantics
// when applicable.
//
//   - If req implements Validatable, ValidateAll() is called and the
//     returned error is flattened into a compound ErrValidation.
//   - If req only implements ValidatableFirst (rare fallback), Validate()
//     is called and its error is wrapped into ErrValidation.
//   - Otherwise (non-proto struct / proto without validate rules), returns
//     nil immediately (NOT an error).
//
// The returned error is always nil or *ErrValidation so that adapters can
// detect it and convert into a 400 HTTP response with body
// {code,reason,message} using the exact same layout used by
// server.ErrBadRequest.
func Validate(req any) error {
	if req == nil {
		return nil
	}
	if v, ok := req.(Validatable); ok {
		if err := v.ValidateAll(); err != nil {
			return &ErrValidation{Violations: flatten(err)}
		}
		return nil
	}
	if v, ok := req.(ValidatableFirst); ok {
		if err := v.Validate(); err != nil {
			return &ErrValidation{Violations: flatten(err)}
		}
		return nil
	}
	return nil
}

// ErrValidation is sentinel type that wraps one or more protobuf-level
// validation violations. Adapters (gin/kratos/std) return ErrBadRequest
// with this error's Error() as the message text.
type ErrValidation struct {
	Violations []Violation
}

// Violation is a flattened, protocol-agnostic description of one field
// validation failure. It intentionally mirrors the generated
// *ValidationError struct fields but lives in this package so that
// business-layer code can construct violations (when needed) without
// pulling generated proto types.
type Violation struct {
	Field  string `json:"field"`
	Reason string `json:"reason"`
}

func (e *ErrValidation) Error() string {
	if len(e.Violations) == 0 {
		return "validation failed"
	}
	parts := make([]string, 0, len(e.Violations))
	for _, v := range e.Violations {
		if v.Field == "" {
			parts = append(parts, v.Reason)
		} else {
			parts = append(parts, fmt.Sprintf("%s: %s", v.Field, v.Reason))
		}
	}
	return strings.Join(parts, "; ")
}

// IsErrValidation reports whether err (or any error in its chain) is a
// protobuf-level validation failure produced by Validate(req). Adapters
// should convert it to ErrBadRequest (HTTP 400) rather than ErrInternal
// (HTTP 500). Binding JSON parse errors are separate (typically *json.SyntaxError
// or *json.UnmarshalTypeError) and should also become 400s via a distinct
// code path, so this helper only matches explicit validate output.
func IsErrValidation(err error) bool {
	if err == nil {
		return false
	}
	var ev *ErrValidation
	return errors.As(err, &ev)
}

// flatten unwraps err into zero or more Violation records. It handles the
// generated *MultiError, any individual *ValidationError, as well as a
// plain error string (rare fallback that captures the raw message).
func flatten(err error) []Violation {
	if err == nil {
		return nil
	}
	// Prefer the generated slice-wrapping multi-error first to get the
	// exhaustive list of field problems.
	var me multiError
	if errors.As(err, &me) {
		all := me.AllErrors()
		out := make([]Violation, 0, len(all))
		for _, e := range all {
			out = append(out, flattenOne(e)...)
		}
		return out
	}
	return flattenOne(err)
}

func flattenOne(err error) []Violation {
	if err == nil {
		return nil
	}
	// Single generated ValidationError: use Field() + Reason().
	var ve ValidationError
	if errors.As(err, &ve) {
		return []Violation{{
			Field:  fieldJSONName(ve.Field()),
			Reason: ve.Reason(),
		}}
	}
	// Fallback: unknown error shape (e.g. custom ValidateAll that returns
	// a plain fmt.Errorf). Preserve original text with an empty field so
	// the 400 message still contains actionable text.
	return []Violation{{Field: "", Reason: err.Error()}}
}

// fieldJSONName maps Go-generated field names to the lowerCamelCase JSON
// keys that REST clients send in their request bodies. protoc-gen-validate
// reports field names using the Go struct name (e.g. "Username") while
// frontend code uses the json_name annotation (e.g. "username"). Matching
// the json_name style keeps 400 messages aligned with the actual request
// body shape.
//
// This function intentionally only lowercases the first rune because the
// generated field names always follow the "UpperCamelCase" Go convention
// for proto3 fields. For rare edge cases (e.g. acronyms like "XMLHTTPRequest")
// the raw Go name is preserved so behaviour is still deterministic rather
// than guessing wrong.
func fieldJSONName(goField string) string {
	if goField == "" {
		return ""
	}
	r := []rune(goField)
	if len(r) == 1 {
		return strings.ToLower(string(r))
	}
	return strings.ToLower(string(r[0])) + string(r[1:])
}
