package crypt

import (
	"bytes"
	"reflect"
	"testing"
)

type testStruct struct {
	Value string
}

func TestEncryptString(t *testing.T) {
	t.Setenv("APP_SECRET", "foobarbazfoobarbazfoobarbazfoobarbaz")
	val := []byte("testing")

	encVal, err := Encrypt(val)
	if err != nil {
		t.Error("Error on encrypting string")
	}

	decVal, err := Decrypt(encVal)
	if err != nil {
		t.Error("Error on decrypting string")
	}

	if !bytes.Equal(val, decVal) {
		t.Errorf("got %s but want %s", decVal, val)
	}
}

func TestEncryptStruct(t *testing.T) {
	t.Setenv("APP_SECRET", "foobarbazfoobarbazfoobarbazfoobarbaz")
	anonStruct := &testStruct{
		Value: "testing",
	}

	encStruct, err := EncryptStruct(anonStruct)
	if err != nil {
		t.Error("Error on encrypting struct")
	}

	var decStruct testStruct
	err = DecryptStruct(encStruct, &decStruct)
	if err != nil {
		t.Error("Error on decrypting struct")
	}

	if !reflect.DeepEqual(anonStruct, &decStruct) {
		t.Errorf("got %+v but want %+v", decStruct, anonStruct)
	}
}

func TestEncryptStructBase64(t *testing.T) {
	t.Setenv("APP_SECRET", "foobarbazfoobarbazfoobarbazfoobarbaz")
	anonStruct := &testStruct{
		Value: "testing",
	}

	encStruct, err := EncryptStructBase64(anonStruct)
	if err != nil {
		t.Error("Error on encrypting struct")
	}

	var decStruct testStruct
	err = DecryptStructBase64(encStruct, &decStruct)
	if err != nil {
		t.Error("Error on decrypting struct")
	}

	if !reflect.DeepEqual(anonStruct, &decStruct) {
		t.Errorf("got %+v but want %+v", decStruct, anonStruct)
	}
}
