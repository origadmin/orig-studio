package server

// Common error codes
const (
	ErrOK           = 0
	ErrInternal     = 10000
	ErrNotFound     = 10001
	ErrUnauthorized = 10002
	ErrForbidden    = 10003
	ErrBadRequest   = 10004
	ErrConflict     = 10005
)

// Business error codes
const (
	ErrUserNotFound  = 20001
	ErrUserExists    = 20002
	ErrPasswordWrong = 20003
	ErrTokenExpired  = 20004
	ErrTokenInvalid  = 20005

	ErrMediaNotFound  = 30001
	ErrMediaTooLarge  = 30002
	ErrMediaForbidden = 30003
	ErrEncodingFailed = 30004

	ErrCommentNotFound  = 40001
	ErrCommentForbidden = 40002
)
