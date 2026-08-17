package request

import "testing"

func TestRegisterRequest_ToUseCase(t *testing.T) {
	req := RegisterRequest{
		Name:     "Ada",
		Email:    "ada@example.com",
		Password: "secret1",
	}

	got := req.ToUseCase()

	if got.Name != req.Name {
		t.Errorf("want Name %q, got %q", req.Name, got.Name)
	}
	if got.Email != req.Email {
		t.Errorf("want Email %q, got %q", req.Email, got.Email)
	}
	if got.Password != req.Password {
		t.Errorf("want Password %q, got %q", req.Password, got.Password)
	}
}

func TestUpdateUserRequest_ToUseCase(t *testing.T) {
	t.Run("both fields set are passed through as pointers", func(t *testing.T) {
		name := "Ada"
		email := "ada@example.com"
		req := UpdateRequest{Name: &name, Email: &email}

		got := req.ToUseCase()

		if got.Name == nil || *got.Name != name {
			t.Errorf("want Name %q, got %v", name, got.Name)
		}
		if got.Email == nil || *got.Email != email {
			t.Errorf("want Email %q, got %v", email, got.Email)
		}
	})

	t.Run("nil fields stay nil (not converted to empty string)", func(t *testing.T) {
		req := UpdateRequest{}

		got := req.ToUseCase()

		if got.Name != nil {
			t.Errorf("want Name to stay nil, got %q", *got.Name)
		}
		if got.Email != nil {
			t.Errorf("want Email to stay nil, got %q", *got.Email)
		}
	})

	t.Run("only one field set leaves the other nil", func(t *testing.T) {
		name := "Ada"
		req := UpdateRequest{Name: &name}

		got := req.ToUseCase()

		if got.Name == nil || *got.Name != name {
			t.Errorf("want Name %q, got %v", name, got.Name)
		}
		if got.Email != nil {
			t.Errorf("want Email to remain nil, got %q", *got.Email)
		}
	})
}
