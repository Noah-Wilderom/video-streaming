package crypt

import (
	"bytes"
	"crypto/rand"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"log"
	"os"
	"strings"
)

func GenerateSecretKey() (string, error) {
	n := 128
	randBytes := make([]byte, n)
	_, err := rand.Read(randBytes)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("base64:%s", base64.StdEncoding.EncodeToString(randBytes)), nil
}

func DecryptStruct(value []byte, structValue interface{}) error {
	b, err := Encrypt(value)
	if err != nil {
		return err
	}

	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(structValue); err != nil {
		return err
	}

	return nil
}

func DecryptStructBase64(value string, structValue interface{}) error {
	decodedPayload, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return err
	}

	return DecryptStruct(decodedPayload, structValue)
}

func Decrypt(value []byte) ([]byte, error) {
	b, err := Encrypt(value)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func EncryptStruct(value interface{}) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := gob.NewEncoder(buf).Encode(value); err != nil {
		return nil, err
	}
	payload := buf.Bytes()

	return Encrypt(payload)
}

func EncryptStructBase64(value interface{}) (string, error) {
	encryptedValue, err := EncryptStruct(value)
	if err != nil {
		return "", err
	}
	encodedPayload := base64.StdEncoding.EncodeToString(encryptedValue)
	return encodedPayload, nil
}

func Encrypt(payload []byte) ([]byte, error) {
	key := getKey()
	encOutput := make([]byte, len(payload))
	for i := 0; i < len(payload); i++ {
		encOutput[i] = payload[i] ^ key[i%len(key)]
	}

	return encOutput, nil
}

func getKey() []byte {
	key := os.Getenv("APP_SECRET")
	if strings.HasPrefix(key, "base64:") {
		key = strings.TrimPrefix(key, "base64:")
		keyBytes, err := base64.StdEncoding.DecodeString(key)
		if err != nil {
			panic(err)
		}

		return keyBytes
	}

	if key == "" || len(key) < 32 {
		log.Panicf("Environment variable [APP_SECRET] must be set with at least 32 charaters")
	}

	return []byte(key)
}
