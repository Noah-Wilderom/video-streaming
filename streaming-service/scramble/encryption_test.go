package scramble

import (
	"crypto/rand"
	"os"
	"testing"
)

func TestDefaultEncryption(t *testing.T) {
	expected := "password"

	encryption := DefaultEncryption{}

	encrypted, err := encryption.Encrypt(expected)
	if err != nil {
		t.Errorf("error on encryption: %v", err)
	}

	decrypted, err := encryption.Decrypt(encrypted)
	if err != nil {
		t.Errorf("error on decryption: %v", err)
	}

	if decrypted != expected {
		t.Errorf("expected decrypted to be %s, got %s", expected, decrypted)
	}
}

func TestAES256Encryption(t *testing.T) {
	expected := "password"

	secret := make([]byte, 32) // allocate a byte slice of 32 bytes
	if _, err := rand.Read(secret); err != nil {
		t.Error(err)
	}

	encryption := AES256Encryption{
		Key: secret,
	}

	encrypted, err := encryption.Encrypt(expected)
	if err != nil {
		t.Errorf("error on encryption: %v", err)
	}

	decrypted, err := encryption.Decrypt(encrypted)
	if err != nil {
		t.Errorf("error on decryption: %v", err)
	}

	if decrypted != expected {
		t.Errorf("expected decrypted to be %s, got %s", expected, decrypted)
	}
}

func TestPKCS1Encryption(t *testing.T) {
	expected := "password"

	encryption := PKCS1Encryption{
		PublicKeyPath:  "",
		PrivateKeyPath: "",
	}

	tmpPath := "./tmp"
	err := os.Mkdir(tmpPath, os.ModePerm)
	if err != nil {
		t.Error(err)
	}

	pairs, err := encryption.GeneratePairs(tmpPath, 4096)
	if err != nil {
		t.Errorf("error on generating key pairs: %v", err)
	}

	encryption.PublicKeyPath = pairs.Server.PublicPath
	encryption.PrivateKeyPath = pairs.Server.PrivatePath

	encrypted, err := encryption.Encrypt(expected)
	if err != nil {
		t.Errorf("error on encryption: %v", err)
	}

	decrypted, err := encryption.Decrypt(encrypted)
	if err != nil {
		t.Errorf("error on decryption: %v", err)
	}

	if decrypted != expected {
		t.Errorf("expected decrypted to be %s, got %s", expected, decrypted)
	}

	if err = os.RemoveAll(tmpPath); err != nil {
		t.Errorf("error on removing temporary directory: %v", err)
	}
}
