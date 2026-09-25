//go:build crypto_dummy || all_crypto

package dummy

import (
	"sparkbridge/pkg/adapters/registry"
	"sparkbridge/pkg/interfaces"
)

// DummyEncryptor is a no-op encryptor used for registry verification.
type DummyEncryptor struct{}

// Encrypt returns the payload unchanged.
func (d *DummyEncryptor) Encrypt(payload []byte) ([]byte, error) { return payload, nil }

// NewDummyEncryptor constructs the encryptor.
func NewDummyEncryptor() (interfaces.Encryptor, error) { return &DummyEncryptor{}, nil }

func init() { registry.RegisterCrypto("dummy", NewDummyEncryptor) }
