package scramble

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Encryption interface {
	Encrypt(string) (string, error)
	Decrypt(string) (string, error)
}

type DefaultEncryption struct{}

// Encrypt implements the Encryption interface
func (d *DefaultEncryption) Encrypt(s string) (string, error) {
	return s, nil
}

// Decrypt implements the Encryption interface
func (d *DefaultEncryption) Decrypt(s string) (string, error) {
	return s, nil
}

type AES256Encryption struct {
	Key []byte
}

// Encrypt implements the Encryption interface
func (e *AES256Encryption) Encrypt(s string) (string, error) {
	plaintext := []byte(s)
	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt implements the Encryption interface
func (e *AES256Encryption) Decrypt(s string) (string, error) {
	cipherText, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(e.Key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := cipherText[:nonceSize], cipherText[nonceSize:]

	decrypted, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}

type PKCS1Encryption struct {
	PublicKeyPath  string
	PrivateKeyPath string
}

func (e *PKCS1Encryption) parsePublicKey() (*rsa.PublicKey, error) {
	publicKeyPEM, err := os.ReadFile(e.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read public key file: %v", err)
	}

	block, _ := pem.Decode(publicKeyPEM)
	if block == nil || block.Type != "PUBLIC KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing public key")
	}

	publicKey, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse public key: %v", err)
	}

	rsaPublicKey, ok := publicKey.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not an RSA public key")
	}

	return rsaPublicKey, nil
}
func (e *PKCS1Encryption) parsePrivateKey() (*rsa.PrivateKey, error) {
	privateKeyPEM, err := os.ReadFile(e.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %v", err)
	}

	block, _ := pem.Decode(privateKeyPEM)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		return nil, fmt.Errorf("failed to decode PEM block containing private key")
	}

	privateKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %v", err)
	}

	return privateKey, nil
}

type GeneratePairsResponse struct {
	Client *KeyPair
	Server *KeyPair
}

func (e *PKCS1Encryption) GeneratePairs(path string, bits int) (*GeneratePairsResponse, error) {
	clientPath := filepath.Join(path, "client")
	serverPath := filepath.Join(path, "server")

	err := os.Mkdir(clientPath, os.ModePerm)
	if err != nil {
		return nil, err
	}
	err = os.Mkdir(serverPath, os.ModePerm)
	if err != nil {
		return nil, err
	}

	clientPair, err := e.generatePair(filepath.Join(clientPath, "private.pem"), filepath.Join(serverPath, "public.pem"), bits)
	if err != nil {
		return nil, err
	}

	serverPair, err := e.generatePair(filepath.Join(serverPath, "private.pem"), filepath.Join(clientPath, "public.pem"), bits)
	if err != nil {
		return nil, err
	}

	return &GeneratePairsResponse{
		Client: clientPair,
		Server: serverPair,
	}, nil
}

type KeyPair struct {
	PublicPath  string
	PrivatePath string
	PublicKey   *rsa.PublicKey
	PrivateKey  *rsa.PrivateKey
}

func (e *PKCS1Encryption) generatePair(privatePath string, publicPath string, bits int) (*KeyPair, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, bits)
	if err != nil {
		return nil, err
	}

	privateFile, err := os.Create(privatePath)
	if err != nil {
		return nil, err
	}
	defer privateFile.Close()

	publicFile, err := os.Create(publicPath)
	if err != nil {
		return nil, err
	}
	defer publicFile.Close()

	privateKeyBytes := x509.MarshalPKCS1PrivateKey(privateKey)
	publicKeyBytes, err := x509.MarshalPKIXPublicKey(&privateKey.PublicKey)
	if err != nil {
		return nil, err
	}

	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privateKeyBytes,
	}

	publicKeyPEM := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: publicKeyBytes,
	}

	err = pem.Encode(privateFile, privateKeyPEM)
	if err != nil {
		return nil, err
	}

	err = pem.Encode(publicFile, publicKeyPEM)
	if err != nil {
		return nil, err
	}

	return &KeyPair{
		PublicPath:  publicPath,
		PrivatePath: privatePath,
		PublicKey:   &privateKey.PublicKey,
		PrivateKey:  privateKey,
	}, nil
}

func (e *PKCS1Encryption) Encrypt(s string) (string, error) {
	publicKey, err := e.parsePublicKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, []byte(s))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (e *PKCS1Encryption) Decrypt(s string) (string, error) {
	privateKey, err := e.parsePrivateKey()
	if err != nil {
		return "", err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	data, err := rsa.DecryptPKCS1v15(rand.Reader, privateKey, ciphertext)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
