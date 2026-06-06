package services

type IHasher interface {
	HashPassword(password string) (string, error)
	ComparePassword(password, hash string) bool
}
