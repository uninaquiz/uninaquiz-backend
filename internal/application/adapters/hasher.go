package adapters

import "golang.org/x/crypto/bcrypt"

type HashPasswordAdapter struct {
}

func NewHashPasswordAdapter() *HashPasswordAdapter {
	return &HashPasswordAdapter{}
}

func (adp *HashPasswordAdapter) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func (adp *HashPasswordAdapter) ComparePassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
