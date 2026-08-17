package request

import "testing"

func TestLoginRequest_ToUseCase(t *testing.T) {
	req := LoginRequest{
		Email:    "ada@example.com",
		Password: "secret1",
	}

	got := req.ToUseCase()

	if got.Email != req.Email {
		t.Errorf("want Email %q, got %q", req.Email, got.Email)
	}
	if got.Password != req.Password {
		t.Errorf("want Password %q, got %q", req.Password, got.Password)
	}
}
