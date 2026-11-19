// Package external contains the implementation of the secret storage.
package external

import (
	sci "github.com/kubex-ecosystem/ethyr/tools/security/interfaces"
	"github.com/kubex-ecosystem/ethyr/tools/security/storage"
)

// SecretStorageAdapter is an adapter for the secret storage.
type SecretStorageAdapter struct {
	backend storage.ISecretStorage
}

// NewSecretStorageAdapter creates a new instance of SecretStorageAdapter.
func NewSecretStorageAdapter(backend storage.ISecretStorage) sci.IKeyringService {
	return &SecretStorageAdapter{backend: backend}
}

// StorePassword stores a password in the secret storage.
func (a *SecretStorageAdapter) StorePassword(password string) error {
	return a.backend.StorePassword(password)
}

// RetrievePassword retrieves a password from the secret storage.
func (a *SecretStorageAdapter) RetrievePassword() (string, error) {
	return a.backend.RetrievePassword()
}
