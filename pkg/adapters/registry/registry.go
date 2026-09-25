package registry

import (
	"fmt"
	"sync"

	"sparkbridge/pkg/interfaces"
)

// InputFactory constructs an input adapter.
type InputFactory func() (interfaces.InputAdapter, error)

// SinkFactory constructs an output sink.
type SinkFactory func() (interfaces.OutputSink, error)

// StorageFactory constructs a state store.
type StorageFactory func() (interfaces.StateStore, error)

// CryptoFactory constructs an encryptor.
type CryptoFactory func() (interfaces.Encryptor, error)

var mu sync.RWMutex

var inputs = map[string]InputFactory{}
var sinks = map[string]SinkFactory{}
var storages = map[string]StorageFactory{}
var cryptos = map[string]CryptoFactory{}

// RegisterInput registers an input adapter factory.
func RegisterInput(name string, factory InputFactory) {
	mu.Lock()
	defer mu.Unlock()
	inputs[name] = factory
}

// RegisterSink registers an output sink factory.
func RegisterSink(name string, factory SinkFactory) {
	mu.Lock()
	defer mu.Unlock()
	sinks[name] = factory
}

// RegisterStorage registers a storage factory.
func RegisterStorage(name string, factory StorageFactory) {
	mu.Lock()
	defer mu.Unlock()
	storages[name] = factory
}

// RegisterCrypto registers a crypto factory.
func RegisterCrypto(name string, factory CryptoFactory) {
	mu.Lock()
	defer mu.Unlock()
	cryptos[name] = factory
}

// Input returns a registered input factory.
func Input(name string) (InputFactory, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := inputs[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not included in this build; recompile with -tags input_%s", name, name)
	}
	return factory, nil
}

// Sink returns a registered sink factory.
func Sink(name string) (SinkFactory, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := sinks[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not included in this build; recompile with -tags sink_%s", name, name)
	}
	return factory, nil
}

// Storage returns a registered storage factory.
func Storage(name string) (StorageFactory, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := storages[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not included in this build; recompile with -tags storage_%s", name, name)
	}
	return factory, nil
}

// Crypto returns a registered crypto factory.
func Crypto(name string) (CryptoFactory, error) {
	mu.RLock()
	defer mu.RUnlock()
	factory, ok := cryptos[name]
	if !ok {
		return nil, fmt.Errorf("adapter %q not included in this build; recompile with -tags crypto_%s", name, name)
	}
	return factory, nil
}

// InputNames returns a snapshot of registered inputs.
func InputNames() []string {
	mu.RLock()
	defer mu.RUnlock()
	out := make([]string, 0, len(inputs))
	for name := range inputs {
		out = append(out, name)
	}
	return out
}
