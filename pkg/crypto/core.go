// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"

	luxcrypto "github.com/luxfi/crypto"
	"golang.org/x/crypto/hkdf"
)

// Core provides unified cryptographic operations using LuxFi crypto
type Core struct {
	hpke *HPKEImpl
}

// NewCore creates a new Core crypto instance
func NewCore() *Core {
	return &Core{
		hpke: NewHPKE(),
	}
}

// GenerateKeyPair generates an ECDSA key pair using LuxFi crypto
func (c *Core) GenerateKeyPair() (privateKey, publicKey []byte, err error) {
	privKey, err := luxcrypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	
	pubKeyBytes := luxcrypto.FromECDSAPub(&privKey.PublicKey)
	privKeyBytes := luxcrypto.FromECDSA(privKey)
	
	return privKeyBytes, pubKeyBytes, nil
}

// GenerateHPKEKeyPair generates an X25519 key pair for HPKE
func (c *Core) GenerateHPKEKeyPair() (publicKey, privateKey []byte, err error) {
	return c.hpke.GenerateKeyPair()
}

// Hash computes SHA256 hash
func (c *Core) Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// HashHex computes SHA256 and returns hex string
func (c *Core) HashHex(data []byte) string {
	return hex.EncodeToString(c.Hash(data))
}

// CreateCommitment creates a cryptographic commitment
func (c *Core) CreateCommitment(data []byte) []byte {
	return c.Hash(data)
}

// DeriveKey derives a key using HKDF
func (c *Core) DeriveKey(secret, salt, info []byte, length int) ([]byte, error) {
	h := hkdf.New(sha256.New, secret, salt, info)
	key := make([]byte, length)
	if _, err := h.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

// RandomBytes generates secure random bytes
func (c *Core) RandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return nil, err
	}
	return b, nil
}

// EncryptWithHPKE encrypts data using HPKE with a simplified interface
func (c *Core) EncryptWithHPKE(recipientPublicKey, plaintext []byte) ([]byte, error) {
	encap, err := c.hpke.Encapsulate(recipientPublicKey)
	if err != nil {
		return nil, err
	}
	
	ciphertext, err := c.hpke.Seal(encap.SharedSecret, plaintext, nil)
	if err != nil {
		return nil, err
	}
	
	// Prepend encapsulated key to ciphertext
	result := append(encap.EncapsulatedKey, ciphertext...)
	return result, nil
}

// DecryptWithHPKE decrypts data using HPKE with a simplified interface
func (c *Core) DecryptWithHPKE(privateKey, ciphertext []byte) ([]byte, error) {
	if len(ciphertext) < 32 {
		return nil, ErrInvalidCiphertext
	}
	
	encapsulatedKey := ciphertext[:32]
	actualCiphertext := ciphertext[32:]
	
	sharedSecret, err := c.hpke.Decapsulate(encapsulatedKey, privateKey)
	if err != nil {
		return nil, err
	}
	
	plaintext, err := c.hpke.Open(sharedSecret, actualCiphertext, nil)
	if err != nil {
		return nil, err
	}
	
	return plaintext, nil
}

// ValidateKeySize checks if key has expected size
func (c *Core) ValidateKeySize(key []byte, expectedSize int) error {
	if len(key) != expectedSize {
		return ErrInvalidKeySize
	}
	return nil
}

// Global functions for backward compatibility

var defaultCore = NewCore()

// GenerateKeyPair generates an ECDSA key pair
func GenerateKeyPair() (privateKey, publicKey []byte, err error) {
	return defaultCore.GenerateKeyPair()
}

// Hash computes SHA256 hash
func Hash(data []byte) []byte {
	return defaultCore.Hash(data)
}

// CreateCommitment creates a cryptographic commitment
func CreateCommitment(data []byte) []byte {
	return defaultCore.CreateCommitment(data)
}

// EncryptWithHPKE encrypts data using HPKE
func EncryptWithHPKE(publicKey, plaintext []byte) ([]byte, error) {
	return defaultCore.EncryptWithHPKE(publicKey, plaintext)
}

// DecryptWithHPKE decrypts data using HPKE
func DecryptWithHPKE(privateKey, ciphertext []byte) ([]byte, error) {
	return defaultCore.DecryptWithHPKE(privateKey, ciphertext)
}

// NewHPKE returns the default HPKE implementation
func NewHPKE() *HPKEImpl {
	return NewHPKEImpl()
}