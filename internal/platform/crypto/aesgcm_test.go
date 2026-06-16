package crypto

import (
	"bytes"
	"testing"
)

func TestEncryptAESGCMDecryptAESGCM_RoundTrip(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("secret credential")

	envelope, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptAESGCM() error = %v", err)
	}

	if bytes.Contains(envelope, plaintext) {
		t.Fatal("envelope contains plaintext")
	}
	if len(envelope) <= 1+12 {
		t.Fatalf("len(envelope) = %d, want version, nonce, and ciphertext", len(envelope))
	}
	if envelope[0] != 1 {
		t.Errorf("envelope version = %d, want 1", envelope[0])
	}

	got, err := DecryptAESGCM(key, envelope)
	if err != nil {
		t.Fatalf("DecryptAESGCM() error = %v", err)
	}
	if !bytes.Equal(got, plaintext) {
		t.Errorf("DecryptAESGCM() = %q, want %q", got, plaintext)
	}
}

func TestEncryptAESGCM_UsesRandomNoncePerEnvelope(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")
	plaintext := []byte("same secret")

	first, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptAESGCM() first error = %v", err)
	}
	second, err := EncryptAESGCM(key, plaintext)
	if err != nil {
		t.Fatalf("EncryptAESGCM() second error = %v", err)
	}

	if bytes.Equal(first, second) {
		t.Fatal("EncryptAESGCM() returned identical envelopes for repeated plaintext")
	}
	if bytes.Equal(first[1:13], second[1:13]) {
		t.Fatal("EncryptAESGCM() reused nonce")
	}
}

func TestEncryptAESGCMDecryptAESGCM_RejectInvalidMasterKey(t *testing.T) {
	invalidKeys := [][]byte{
		nil,
		[]byte("short"),
		[]byte("0123456789abcdef0123456789abcde"),
		[]byte("0123456789abcdef0123456789abcdef0"),
	}

	for _, key := range invalidKeys {
		t.Run("encrypt", func(t *testing.T) {
			if _, err := EncryptAESGCM(key, []byte("secret")); err == nil {
				t.Fatal("EncryptAESGCM() error = nil, want invalid key error")
			}
		})
		t.Run("decrypt", func(t *testing.T) {
			if _, err := DecryptAESGCM(key, append([]byte{1}, make([]byte, 12+16)...)); err == nil {
				t.Fatal("DecryptAESGCM() error = nil, want invalid key error")
			}
		})
	}
}

func TestDecryptAESGCM_RejectsInvalidEnvelope(t *testing.T) {
	key := []byte("0123456789abcdef0123456789abcdef")

	envelope, err := EncryptAESGCM(key, []byte("secret"))
	if err != nil {
		t.Fatalf("EncryptAESGCM() error = %v", err)
	}

	cases := []struct {
		name     string
		envelope []byte
	}{
		{name: "empty", envelope: nil},
		{name: "too short", envelope: append([]byte{1}, make([]byte, 12+15)...)},
		{name: "unknown version", envelope: append([]byte{2}, envelope[1:]...)},
		{name: "tampered ciphertext", envelope: tamperLastByte(envelope)},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if _, err := DecryptAESGCM(key, c.envelope); err == nil {
				t.Fatal("DecryptAESGCM() error = nil, want envelope error")
			}
		})
	}
}

func tamperLastByte(in []byte) []byte {
	out := append([]byte(nil), in...)
	out[len(out)-1] ^= 0xff
	return out
}
