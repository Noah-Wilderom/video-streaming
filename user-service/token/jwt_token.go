package token

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTToken struct {
}

func NewJWTTokenHandler() *JWTToken {
	return &JWTToken{}
}

// New implements the Token interface.
func (t *JWTToken) New(data map[string]any) (string, error) {
	exp := time.Now().Add(time.Hour).Unix()
	// Create a new token object, specifying signing method and the claims
	// you would like it to contain.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data":       data,
		"created_at": time.Now(),
		"exp":        exp,
	})

	// Sign and get the complete encoded token as a string using the secret
	return token.SignedString([]byte(getSecret()))
}

// Validate implements the token interface.
func (t JWTToken) Validate(token string) bool {
	jwtToken, err := jwt.Parse(token, func(token *jwt.Token) (interface{}, error) {
		return []byte(getSecret()), nil
	})

	if err != nil {
		fmt.Println("err?")
		return false
	}

	if claims, ok := jwtToken.Claims.(jwt.MapClaims); ok {
		if exp, ok := claims["exp"]; ok {
			if expFloat, ok := exp.(float64); ok {
				if !time.Unix(int64(expFloat), 0).Before(time.Now()) {
					return true
				} else {
					fmt.Println("Token expired")
				}
			}
		}

	}

	return false
}
