package services

import "time"

type ITokenService interface {
	GenerateToken(userID string, username string, expirationTime time.Duration) (string, error)
	Verify(token string) error
	ExtractClaims(tokenString string) (map[string]interface{}, error)
}
