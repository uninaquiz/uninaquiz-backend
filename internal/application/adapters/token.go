package adapters

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JwtTokenAdapter struct {
	SecretKey string
}

func NewJwtTokenAdapter(SecretKey string) *JwtTokenAdapter {
	return &JwtTokenAdapter{
		SecretKey: SecretKey,
	}
}

func (adp *JwtTokenAdapter) GenerateToken(userID string, username string, expirationTime time.Duration) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  userID,
		"username": username,
		"exp":      time.Now().Add(expirationTime).Unix(),
	})

	tokenString, err := token.SignedString([]byte(adp.SecretKey))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (adp *JwtTokenAdapter) Verify(tokenString string) error {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(adp.SecretKey), nil
	})

	if err != nil {
		return err
	}
	if !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}

func (adp *JwtTokenAdapter) ExtractClaims(tokenString string) (map[string]interface{}, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(adp.SecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	result := make(map[string]interface{})
	for k, v := range claims {
		result[k] = v
	}
	return result, nil
}
