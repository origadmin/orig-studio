package validate

import (
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	userv1 "origadmin/application/origstudio/api/gen/v1/user"
)

// --- Test fixtures ----------------------------------------------------------

// nonProtoStruct is a plain Go struct that does NOT implement Validatable
// or ValidatableFirst. Validate must skip it and return nil.
type nonProtoStruct struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

// fakeMulti is a manually constructed composite error that mimics the
// generated *MultiError shape. It's used to assert flatten() can handle
// multiError implementations independent of real generated code.
type fakeMulti []error

func (f fakeMulti) Error() string  { return "fake multi" }
func (f fakeMulti) AllErrors() []error { return f }

// fakeSingle mimics a single generated ValidationError.
type fakeSingle struct {
	field  string
	reason string
	cause  error
}

func (f fakeSingle) Error() string  { return f.field + ": " + f.reason }
func (f fakeSingle) Field() string  { return f.field }
func (f fakeSingle) Reason() string { return f.reason }
func (f fakeSingle) Cause() error   { return f.cause }

// --- Tests -----------------------------------------------------------------

func TestValidate_LoginRequestEmpty_ReturnsValidationErrors(t *testing.T) {
	req := &userv1.LoginRequest{Username: "", Password: ""}

	err := Validate(req)

	if assert.Error(t, err) {
		assert.True(t, IsErrValidation(err), "expected *ErrValidation, got %T", err)
		msg := err.Error()
		// Two fields -> two violations concatenated with "; "
		assert.True(t, strings.Contains(msg, "username"), "message should mention username, got: %s", msg)
		assert.True(t, strings.Contains(msg, "password"), "message should mention password, got: %s", msg)

		ev := &ErrValidation{}
		errors.As(err, &ev)
		assert.Len(t, ev.Violations, 2, "expected 2 violations, got %+v", ev.Violations)
	}
}

func TestValidate_LoginRequestUsernameTooShort_ReturnsError(t *testing.T) {
	// min_len is 1 so "" already fails; additionally confirm a valid
	// password does not mask the username violation (ValidateAll lists
	// every problem at once).
	req := &userv1.LoginRequest{Username: "", Password: "secret123"}

	err := Validate(req)

	if assert.Error(t, err) {
		msg := err.Error()
		assert.True(t, strings.Contains(msg, "username"), "expected username in: %s", msg)
		assert.False(t, strings.Contains(msg, "password"), "unexpected password in: %s", msg)
	}
}

func TestValidate_LoginRequestPasswordTooShort_ReturnsError(t *testing.T) {
	req := &userv1.LoginRequest{Username: "alice", Password: "12345"} // min_len=6

	err := Validate(req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password")
}

func TestValidate_LoginRequestValid_ReturnsNil(t *testing.T) {
	req := &userv1.LoginRequest{Username: "alice", Password: "secret123"}
	assert.NoError(t, Validate(req))
}

func TestValidate_NonProtoStruct_ReturnsNil(t *testing.T) {
	// Progressive migration safety: structs that don't implement the
	// Validatable interfaces must be silently skipped so handlers that
	// haven't been converted to proto messages yet keep working.
	req := &nonProtoStruct{Name: "x", Age: 1}
	assert.NoError(t, Validate(req))
}

func TestValidate_NilInput_ReturnsNil(t *testing.T) {
	assert.NoError(t, Validate(nil))
}

func TestValidate_NilTypedPointer_ReturnsNil(t *testing.T) {
	var req *userv1.LoginRequest
	assert.NoError(t, Validate(req))
}

// --- IsErrValidation / flatten helpers -------------------------------------

func TestIsErrValidation_TrueForValidateOutput(t *testing.T) {
	err := Validate(&userv1.LoginRequest{})
	assert.True(t, IsErrValidation(err))
}

func TestIsErrValidation_FalseForPlainError(t *testing.T) {
	assert.False(t, IsErrValidation(nil))
	assert.False(t, IsErrValidation(errors.New("something else")))
}

func TestFlatten_MultiErrorContainingMultipleSingles(t *testing.T) {
	inner := fakeMulti{
		fakeSingle{field: "Username", reason: "bad"},
		fakeSingle{field: "Password", reason: "worse"},
	}

	out := flatten(inner)
	assert.Len(t, out, 2)
	// First-rune lower-case to match JSON request body key naming.
	assert.Equal(t, "username", out[0].Field)
	assert.Equal(t, "bad", out[0].Reason)
	assert.Equal(t, "password", out[1].Field)
	assert.Equal(t, "worse", out[1].Reason)
}

func TestFlatten_SingleValidationError(t *testing.T) {
	out := flatten(fakeSingle{field: "Title", reason: "empty"})
	assert.Len(t, out, 1)
	assert.Equal(t, "title", out[0].Field)
	assert.Equal(t, "empty", out[0].Reason)
}

func TestFlatten_PlainFallbackError_StillCapturesReason(t *testing.T) {
	out := flatten(errors.New("raw issue"))
	assert.Len(t, out, 1)
	assert.Equal(t, "", out[0].Field)          // unknown field -> blank
	assert.Equal(t, "raw issue", out[0].Reason) // but message preserved
}

// --- fieldJSONName ---------------------------------------------------------

func TestFieldJSONName_Simple(t *testing.T) {
	assert.Equal(t, "username", fieldJSONName("Username"))
	assert.Equal(t, "title", fieldJSONName("Title"))
}

func TestFieldJSONName_Empty(t *testing.T) {
	assert.Equal(t, "", fieldJSONName(""))
}

func TestFieldJSONName_SingleRune(t *testing.T) {
	assert.Equal(t, "x", fieldJSONName("X"))
}
