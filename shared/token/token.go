package token

import "os"

func getSecret() string {
	secret := os.Getenv("TOKEN_SECRET")

	return secret
}

type Token interface {
	New(map[string]interface{}) (string, error)
	Validate(token string) bool
}
