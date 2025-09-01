// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package crypto

import (
	"crypto/ecdsa"
	"errors"
)

var (
	// ErrInvalidKeySize indicates the key size is incorrect
	ErrInvalidKeySize = errors.New("invalid key size")
	// ErrInvalidCiphertext indicates the ciphertext is malformed
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
	// ErrInvalidSignature indicates the signature verification failed
	ErrInvalidSignature = errors.New("invalid signature")
)

// KeyPair represents a public/private key pair
type KeyPair struct {
	PublicKey  []byte
	PrivateKey []byte
}

// Signer provides signing operations
type Signer interface {
	// Sign creates a signature for the given message
	Sign(privateKey *ecdsa.PrivateKey, message []byte) ([]byte, error)
	// Verify checks if a signature is valid
	Verify(publicKey *ecdsa.PublicKey, message, signature []byte) bool
}

// Encryptor provides encryption operations
type Encryptor interface {
	// Encrypt encrypts data with a public key
	Encrypt(publicKey []byte, plaintext []byte) ([]byte, error)
	// Decrypt decrypts data with a private key
	Decrypt(privateKey []byte, ciphertext []byte) ([]byte, error)
}

// Hasher provides cryptographic hash operations
type Hasher interface {
	// Hash computes a cryptographic hash
	Hash(data []byte) []byte
	// HashWithSalt computes a salted hash
	HashWithSalt(data, salt []byte) []byte
}

// HPKE provides Hybrid Public Key Encryption operations
type HPKE interface {
	// GenerateKeyPair generates an X25519 key pair
	GenerateKeyPair() (publicKey, privateKey []byte, err error)
	// Encapsulate generates ephemeral key and shared secret
	Encapsulate(recipientPublicKey []byte) (*Encapsulation, error)
	// Decapsulate recovers shared secret from encapsulated key
	Decapsulate(encapsulatedKey, privateKey []byte) ([]byte, error)
}

// Encapsulation contains the encapsulated key and shared secret
type Encapsulation struct {
	EncapsulatedKey []byte
	SharedSecret    []byte
}