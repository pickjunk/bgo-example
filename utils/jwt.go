package utils

import (
	"errors"
	"fmt"

	"github.com/dgrijalva/jwt-go"
)

// JWTClaims struct
type JWTClaims struct {
	ID   *string `json:"id,omitempty"`
	Name *string `json:"name,omitempty"`
	jwt.StandardClaims
}

// GenerateJWT encode JWTClaims to jwt
func GenerateJWT(claims *JWTClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, err := token.SignedString([]byte(secret))
	if err != nil {
		return s, err
	}

	return s, nil
}

// ParseJWT decode jwt to JWTClaims
func ParseJWT(tokenString string, secret string, claims jwt.Claims) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	if sess, ok := token.Claims.(*Session); ok && token.Valid {
		return sess, nil
	}

	return nil, errors.New("ParseJWT fail")
}
