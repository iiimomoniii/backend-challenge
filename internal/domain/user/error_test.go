package user

import (
	"errors"
	"fmt"
	"testing"
)

func TestError_Error(t *testing.T) {
	err := &Error{Code: "USR006", Message: "User not found"}

	if err.Error() != "User not found" {
		t.Errorf("want Error() to return the Message field, got %q", err.Error())
	}
}

func TestAsDomainError(t *testing.T) {
	t.Run("a sentinel domain error is recognized directly", func(t *testing.T) {
		de, ok := AsDomainError(ErrUserNotFound)
		if !ok {
			t.Fatal("want ok=true for a domain error")
		}
		if de.Code != CodeUserNotFound {
			t.Errorf("want Code %q, got %q", CodeUserNotFound, de.Code)
		}
	})

	t.Run("a domain error wrapped with fmt.Errorf(%w) is still recognized", func(t *testing.T) {
		wrapped := fmt.Errorf("create user: %w", ErrUsernameAlreadyExists)
		de, ok := AsDomainError(wrapped)
		if !ok {
			t.Fatal("want ok=true for a wrapped domain error")
		}
		if de.Code != CodeUsernameAlreadyExists {
			t.Errorf("want Code %q, got %q", CodeUsernameAlreadyExists, de.Code)
		}
	})

	t.Run("a plain non-domain error is not recognized", func(t *testing.T) {
		_, ok := AsDomainError(errors.New("connection refused"))
		if ok {
			t.Fatal("want ok=false for a non-domain error")
		}
	})

	t.Run("nil error is not recognized", func(t *testing.T) {
		_, ok := AsDomainError(nil)
		if ok {
			t.Fatal("want ok=false for a nil error")
		}
	})

	t.Run("every sentinel error's Code matches its own constant", func(t *testing.T) {
		cases := map[string]*Error{
			CodeUsernameRequired:      ErrUsernameRequired,
			CodePasswordRequired:      ErrPasswordRequired,
			CodePasswordTooShort:      ErrPasswordTooShort,
			CodeNameRequired:          ErrNameRequired,
			CodeUsernameAlreadyExists: ErrUsernameAlreadyExists,
			CodeUserNotFound:          ErrUserNotFound,
			CodeInvalidCredentials:    ErrInvalidCredentials,
			CodeInvalidInput:          ErrInvalidInput,
		}

		for wantCode, sentinel := range cases {
			if sentinel.Code != wantCode {
				t.Errorf("sentinel for %s has Code %q instead", wantCode, sentinel.Code)
			}
		}
	})
}
