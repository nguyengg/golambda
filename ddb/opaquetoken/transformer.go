package opaquetoken

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// Transformer can be used to encrypt/decrypt the tokens to conform to opaque token principle.
type Transformer interface {
	// Encode is called when converting from last evaluated key to string token.
	Encode(string) (string, error)
	// Decode is called when converting from string token to exclusive start key.
	Decode(string) (string, error)
}

// NewWithAES creates a new Tokenizer with AES encryption and decryption.
func NewWithAES(secretKey []byte) (*Tokenizer, error) {
	c, err := aes.NewCipher(secretKey)
	if err != nil {
		return nil, err
	}

	return &Tokenizer{Transformer: aesTransformer{c}}, nil
}

type aesTransformer struct {
	c cipher.Block
}

func (a aesTransformer) Encode(s string) (string, error) {
	gcm, err := cipher.NewGCM(a.c)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	data := gcm.Seal(nonce, nonce, []byte(s), nil)
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func (a aesTransformer) Decode(s string) (string, error) {
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(a.c)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	if n := len(s); n < nonceSize {
		return "", fmt.Errorf("invalid token; expected size of at least %d bytes, got %d bytes", nonceSize, n)
	}

	data, nonce := data[nonceSize:], data[:nonceSize]
	token, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}

	return string(token), err
}
