package server

import (
	"github.com/semaphoreui/semaphore/db"
)

type RDMStorageTokenDeserializer interface {
	DeserializeSecret(key *db.AccessKey) error
}

type RDMAccessKeyDeserializer struct {
}

func NewRDMAccessKeyDeserializer(
	_ db.AccessKeyManager,
	_ db.SecretStorageRepository,
	_ VaultStorageTokenDeserializer,
) *RDMAccessKeyDeserializer {
	return &RDMAccessKeyDeserializer{}
}

func (d *VaultAccessKeyDeserializer) DeleteSecret(key *db.AccessKey) error {
	return nil
}

func (d *VaultAccessKeyDeserializer) SerializeSecret(key *db.AccessKey) (err error) {
	return
}

func (d *VaultAccessKeyDeserializer) DeserializeSecret(key *db.AccessKey) (res string, err error) {
	return
}
