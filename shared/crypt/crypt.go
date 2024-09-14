package crypt

import (
	"bytes"
	"encoding/gob"
	"log"
	"os"
)

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
	if key == "" || len(key) < 32 {
		log.Panicf("Environment variable [APP_SECRET] must be set with at least 32 charaters")
	}

	return []byte(key)
}
