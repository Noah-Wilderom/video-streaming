package token

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"github.com/Noah-Wilderom/video-streaming/shared/crypt"
	"reflect"
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
	jwtToken, err := h.New(map[string]string{
		"hash": string(hash),
	})

	if err != nil {
		t.Error(err)
		return
	}

	ok, data := h.Validate(jwtToken)
	if !ok {
		t.Error("Token invalid")
	}

	if decryptedHash, ok := data["hash"].(string); ok {
		if string(decryptedHash) != decryptedHash {
			t.Errorf("exptected [%s], got [%s]", string(hash), decryptedHash)
		}
	} else {
		fmt.Printf("%+v\n", reflect.TypeOf(data["hash"]))
		t.Error("hash not found in data")
	}
}

func TestTokenWithStruct(t *testing.T) {
	h := NewJWTTokenHandler()

	t.Setenv("TOKEN_SECRET", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJmb28iOiJiYXIiLCJuYmYiOjE0NDQ0Nzg0MDB9.u1riaD1rW97opCoAuRCTy4w58Br-Zk-bh7vLiRIsrpU")
	t.Setenv("APP_SECRET", "x53PM149FiHuwa31aed5SGpOX16hybzaehzv2s8g3OZXwDEYLW8lIy4mGMSXg1xV")

	type testStruct struct {
		Hash []byte
	}

	dataBytes := make([]byte, 100)
	_, err := rand.Read(dataBytes)
	if err != nil {
		t.Error(err)
		return
	}

	hash := sha256.New().Sum(dataBytes)
	encryptedStruct, err := crypt.EncryptStructBase64(&testStruct{
		Hash: hash,
	})
	if err != nil {
		t.Error(err)
		return
	}

	jwtToken, err := h.New(map[string]string{
		"hash": encryptedStruct,
	})

	if err != nil {
		t.Error(err)
		return
	}

	ok, data := h.Validate(jwtToken)
	if !ok {
		t.Error("Token invalid")
	}

	var decryptedStruct testStruct
	if hashData, ok := data["hash"].(string); ok {
		err = crypt.DecryptStructBase64(hashData, &decryptedStruct)
		if err != nil {
			t.Error(err)
			return
		}
		if ok = bytes.Equal(decryptedStruct.Hash, hash); !ok {
			t.Errorf("expected [%s], got [%s]", string(hash), string(decryptedStruct.Hash))
		}
	}
}
