package user

import "errors"

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}

const (
	CodeUsernameRequired      = "USR001"
	CodePasswordRequired      = "USR002"
	CodePasswordTooShort      = "USR003"
	CodeNameRequired          = "USR004"
	CodeUsernameAlreadyExists = "USR005"
	CodeUserNotFound          = "USR006"
	CodeInvalidCredentials    = "USR007"
	CodeInvalidInput          = "USR008"
)

var (
	ErrUsernameRequired      = &Error{Code: CodeUsernameRequired, Message: "Username is required"}
	ErrPasswordRequired      = &Error{Code: CodePasswordRequired, Message: "Password is required"}
	ErrPasswordTooShort      = &Error{Code: CodePasswordTooShort, Message: "Password must be at least 6 characters"}
	ErrNameRequired          = &Error{Code: CodeNameRequired, Message: "Name is required"}
	ErrUsernameAlreadyExists = &Error{Code: CodeUsernameAlreadyExists, Message: "Username already exists"}
	ErrUserNotFound          = &Error{Code: CodeUserNotFound, Message: "User not found"}
	ErrInvalidCredentials    = &Error{Code: CodeInvalidCredentials, Message: "Invalid email or password"}
	ErrInvalidInput          = &Error{Code: CodeInvalidInput, Message: "Invalid input"}
)

func AsDomainError(err error) (*Error, bool) {
	var de *Error
	if errors.As(err, &de) {
		return de, true
	}
	return nil, false
}
