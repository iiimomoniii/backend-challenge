package bcrypt

import "golang.org/x/crypto/bcrypt"

type Hasher struct {
	bcryptCost int
}

func New() *Hasher {
	return &Hasher{bcryptCost: bcrypt.DefaultCost}
}

func (h *Hasher) Hash(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), h.bcryptCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (h *Hasher) Verify(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}