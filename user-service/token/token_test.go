package token

import (
	"crypto/rand"
	"crypto/sha256"
	"testing"
)

func TestToken(t *testing.T) {
	h := NewJWTTokenHandler()

	t.Setenv("TOKEN_SECRET", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU")

	dataBytes := make([]byte, 100)
	_, err := rand.Read(dataBytes)
	if err != nil {
		t.Error(err)
		return
	}

	hash := sha256.New().Sum(dataBytes)
	jwtToken, err := h.New(map[string]any{
		"hash": hash,
	})

	if err != nil {
		t.Error(err)
		return
	}

	if ok := h.Validate(jwtToken); !ok {
		t.Error("Token invalid")
	}
}
