package auth

type PasswordHasher interface {
	Hash(password string) (string, error)
	Verify(hashedPassword, password string) error
}