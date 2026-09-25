package aesgcm

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"io"

	"sparkbridge/pkg/interfaces"
)

// AESGCM wraps a sink with payload encryption.
type AESGCM struct {
	key  []byte
	peer interfaces.OutputSink
}

// NewAESGCM creates the decorator.
func NewAESGCM(key []byte, peer interfaces.OutputSink) *AESGCM { return &AESGCM{key: key, peer: peer} }

// Publish encrypts payload bytes before publishing.
func (a *AESGCM) Publish(ctx context.Context, topic string, payload []byte) error {
	block, err := aes.NewCipher(a.key)
	if err != nil {
		return err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := make([]byte, aead.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return err
	}
	encrypted := aead.Seal(nonce, nonce, payload, nil)
	if a.peer == nil {
		return nil
	}
	return a.peer.Publish(ctx, topic, encrypted)
}

// Close closes the wrapped sink.
func (a *AESGCM) Close() error {
	if a.peer != nil {
		return a.peer.Close()
	}
	return nil
}
