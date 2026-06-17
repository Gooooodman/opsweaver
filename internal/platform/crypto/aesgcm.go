package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	cryptorand "crypto/rand"
	"errors"
	"fmt"
	"io"
)

const (
	aes256KeySize    = 32
	aesgcmVersion    = byte(1)
	aesgcmNonceSize  = 12
	aesgcmTagSize    = 16
	aesgcmHeaderSize = 1 + aesgcmNonceSize
)

// EncryptAESGCM encrypts plaintext with AES-256-GCM and returns a versioned envelope.
func EncryptAESGCM(masterKey []byte, plaintext []byte) ([]byte, error) {
	gcm, err := newAESGCM(masterKey)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, aesgcmNonceSize)
	if _, err := io.ReadFull(cryptorand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("generate nonce: %w", err)
	}

	envelope := make([]byte, 0, aesgcmHeaderSize+len(plaintext)+gcm.Overhead())
	envelope = append(envelope, aesgcmVersion)
	envelope = append(envelope, nonce...)
	envelope = gcm.Seal(envelope, nonce, plaintext, nil)

	return envelope, nil
}

// DecryptAESGCM decrypts a versioned AES-256-GCM envelope.
func DecryptAESGCM(masterKey []byte, envelope []byte) ([]byte, error) {
	gcm, err := newAESGCM(masterKey)
	if err != nil {
		return nil, err
	}

	if len(envelope) < aesgcmHeaderSize+aesgcmTagSize {
		return nil, errors.New("invalid ciphertext envelope")
	}
	if envelope[0] != aesgcmVersion {
		return nil, errors.New("unsupported ciphertext envelope version")
	}

	nonce := envelope[1:aesgcmHeaderSize]
	ciphertext := envelope[aesgcmHeaderSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, errors.New("decrypt ciphertext envelope")
	}

	return plaintext, nil
}

func newAESGCM(masterKey []byte) (cipher.AEAD, error) {
	if len(masterKey) != aes256KeySize {
		return nil, errors.New("master key must be 32 bytes")
	}

	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, errors.New("initialize AES cipher")
	}

	gcm, err := cipher.NewGCMWithNonceSize(block, aesgcmNonceSize)
	if err != nil {
		return nil, errors.New("initialize AES-GCM")
	}

	return gcm, nil
}
