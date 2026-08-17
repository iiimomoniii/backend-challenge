package response

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	domainuser "github.com/iiimomoniii/backend-challenge/internal/domain/user"
	"github.com/iiimomoniii/backend-challenge/pkg/code"
)

func TestFromUser(t *testing.T) {
	createdAt := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	u := &domainuser.User{
		ID:        "u1",
		Name:      "Ada",
		Email:     "ada@example.com",
		Password:  "super-secret-hash",
		CreatedAt: createdAt,
	}

	got := FromUser(u)

	if got.ID != u.ID || got.Name != u.Name || got.Email != u.Email || !got.CreatedAt.Equal(u.CreatedAt) {
		t.Errorf("FromUser mapped fields incorrectly: got %+v, from %+v", got, u)
	}

	if _, ok := reflect.TypeOf(got).FieldByName("Password"); ok {
		t.Error("UserResponse must not have a Password field")
	}
}

func TestFromUsers(t *testing.T) {
	t.Run("nil input returns an empty, non-nil slice", func(t *testing.T) {
		got := FromUsers(nil)
		if got == nil {
			t.Fatal("want non-nil slice, got nil")
		}
		if len(got) != 0 {
			t.Errorf("want empty slice, got %d items", len(got))
		}
	})

	t.Run("maps every user in order", func(t *testing.T) {
		users := []*domainuser.User{
			{ID: "u1", Name: "Ada"},
			{ID: "u2", Name: "Grace"},
		}
		got := FromUsers(users)
		if len(got) != 2 {
			t.Fatalf("want 2 items, got %d", len(got))
		}
		if got[0].ID != "u1" || got[1].ID != "u2" {
			t.Errorf("want order preserved [u1, u2], got [%s, %s]", got[0].ID, got[1].ID)
		}
	})
}

func TestFromErrorCode(t *testing.T) {
	t.Run("resolves the message in the requested language", func(t *testing.T) {
		got := FromErrorCode(domainuser.CodeUserNotFound, code.LangTH)
		if got.ErrorCode != domainuser.CodeUserNotFound {
			t.Errorf("want ErrorCode %q, got %q", domainuser.CodeUserNotFound, got.ErrorCode)
		}
		if got.Message == "" {
			t.Error("want a non-empty message")
		}
	})

	t.Run("English and Thai produce different messages for the same code", func(t *testing.T) {
		en := FromErrorCode(domainuser.CodeUserNotFound, code.LangEN)
		th := FromErrorCode(domainuser.CodeUserNotFound, code.LangTH)
		if en.Message == th.Message {
			t.Error("want EN and TH messages to differ")
		}
	})
}

func TestStatusForCode(t *testing.T) {
	tests := []struct {
		code string
		want int
	}{
		{domainuser.CodeUsernameRequired, http.StatusBadRequest},
		{domainuser.CodePasswordRequired, http.StatusBadRequest},
		{domainuser.CodePasswordTooShort, http.StatusBadRequest},
		{domainuser.CodeNameRequired, http.StatusBadRequest},
		{domainuser.CodeInvalidInput, http.StatusBadRequest},
		{domainuser.CodeUsernameAlreadyExists, http.StatusConflict},
		{domainuser.CodeUserNotFound, http.StatusNotFound},
		{domainuser.CodeInvalidCredentials, http.StatusUnauthorized},
		{"USR999", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := StatusForCode(tt.code)
			if got != tt.want {
				t.Errorf("StatusForCode(%q) = %d, want %d", tt.code, got, tt.want)
			}
		})
	}
}
