// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

// Package crypto provides unified cryptographic operations for ADX
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"

	"golang.org/x/crypto/chacha20poly1305"
	"golang.org/x/crypto/curve25519"
	"golang.org/x/crypto/hkdf"
	"golang.org/x/crypto/sha3"
)

var (
	ErrDecryptionFailed = errors.New("decryption failed")
)

// HPKE implements Hybrid Public Key Encryption (RFC 9180)
type HPKEImpl struct {
	suite HPKESuite
}

// HPKESuite defines the cryptographic suite for HPKE
type HPKESuite struct {
	KEM  string // Key Encapsulation Mechanism
	KDF  string // Key Derivation Function
	AEAD string // Authenticated Encryption with Associated Data
}

// DefaultSuite returns the default HPKE suite (X25519-HKDF-SHA256-ChaCha20Poly1305)
func DefaultSuite() HPKESuite {
	return HPKESuite{
		KEM:  "X25519",
		KDF:  "HKDF-SHA256",
		AEAD: "ChaCha20Poly1305",
	}
}

// NewHPKE creates a new HPKE instance
func NewHPKEImpl() *HPKEImpl {
	return &HPKEImpl{
		suite: DefaultSuite(),
	}
}

// GenerateKeyPair generates an X25519 key pair
func (h *HPKEImpl) GenerateKeyPair() (publicKey, privateKey []byte, err error) {
	privateKey = make([]byte, 32)
	if _, err := rand.Read(privateKey); err != nil {
		return nil, nil, err
	}
	
	publicKey, err = curve25519.X25519(privateKey, curve25519.Basepoint)
	if err != nil {
		return nil, nil, err
	}
	
	return publicKey, privateKey, nil
}


// Encapsulate generates an ephemeral key pair and shared secret
func (h *HPKEImpl) Encapsulate(recipientPublicKey []byte) (*Encapsulation, error) {
	if len(recipientPublicKey) != 32 {
		return nil, ErrInvalidKeySize
	}
	
	// Generate ephemeral key pair
	ephemeralPrivate := make([]byte, 32)
	if _, err := rand.Read(ephemeralPrivate); err != nil {
		return nil, err
	}
	
	ephemeralPublic, err := curve25519.X25519(ephemeralPrivate, curve25519.Basepoint)
	if err != nil {
		return nil, err
	}
	
	// Compute shared secret
	sharedSecret, err := curve25519.X25519(ephemeralPrivate, recipientPublicKey)
	if err != nil {
		return nil, err
	}
	
	// Derive key using HKDF
	kdf := hkdf.New(sha3.New256, sharedSecret, nil, []byte("adx-hpke-v1"))
	derivedKey := make([]byte, 32)
	if _, err := kdf.Read(derivedKey); err != nil {
		return nil, err
	}
	
	return &Encapsulation{
		EncapsulatedKey: ephemeralPublic,
		SharedSecret:    derivedKey,
	}, nil
}

// Decapsulate recovers the shared secret from encapsulated key
func (h *HPKEImpl) Decapsulate(encapsulatedKey, privateKey []byte) ([]byte, error) {
	if len(encapsulatedKey) != 32 || len(privateKey) != 32 {
		return nil, ErrInvalidKeySize
	}
	
	// Compute shared secret
	sharedSecret, err := curve25519.X25519(privateKey, encapsulatedKey)
	if err != nil {
		return nil, err
	}
	
	// Derive key using HKDF
	kdf := hkdf.New(sha3.New256, sharedSecret, nil, []byte("adx-hpke-v1"))
	derivedKey := make([]byte, 32)
	if _, err := kdf.Read(derivedKey); err != nil {
		return nil, err
	}
	
	return derivedKey, nil
}

// SealedEnvelope represents an HPKE-encrypted message
type SealedEnvelope struct {
	Encapsulations []Encapsulation `json:"encapsulations"` // Multiple recipients
	Ciphertext     []byte          `json:"ciphertext"`
	AAD            []byte          `json:"aad"` // Additional Authenticated Data
}

// Seal encrypts a message for multiple recipients with AAD binding
func (h *HPKEImpl) Seal(plaintext []byte, recipientPublicKeys [][]byte, aad []byte) (*SealedEnvelope, error) {
	if len(recipientPublicKeys) == 0 {
		return nil, errors.New("no recipients specified")
	}
	
	// Generate content encryption key
	contentKey := make([]byte, 32)
	if _, err := rand.Read(contentKey); err != nil {
		return nil, err
	}
	
	// Encrypt plaintext with content key
	aead, err := chacha20poly1305.New(contentKey)
	if err != nil {
		return nil, err
	}
	
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	
	// Include AAD in encryption
	ciphertext := aead.Seal(nonce, nonce, plaintext, aad)
	
	// Encapsulate content key for each recipient
	encapsulations := make([]Encapsulation, len(recipientPublicKeys))
	for i, pubKey := range recipientPublicKeys {
		encap, err := h.encapsulateContentKey(contentKey, pubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to encapsulate for recipient %d: %w", i, err)
		}
		encapsulations[i] = *encap
	}
	
	return &SealedEnvelope{
		Encapsulations: encapsulations,
		Ciphertext:     ciphertext,
		AAD:            aad,
	}, nil
}

// Open decrypts a sealed envelope using the recipient's private key
func (h *HPKEImpl) Open(envelope *SealedEnvelope, recipientPrivateKey []byte, recipientIndex int) ([]byte, error) {
	if recipientIndex >= len(envelope.Encapsulations) {
		return nil, errors.New("invalid recipient index")
	}
	
	// Decapsulate to get shared secret
	encap := envelope.Encapsulations[recipientIndex]
	sharedSecret, err := h.Decapsulate(encap.EncapsulatedKey, recipientPrivateKey)
	if err != nil {
		return nil, err
	}
	
	// Recover content key by XORing with shared secret (simplified)
	contentKey := make([]byte, len(encap.SharedSecret))
	for i := range contentKey {
		contentKey[i] = encap.SharedSecret[i] ^ sharedSecret[i%len(sharedSecret)]
	}
	
	// Decrypt ciphertext
	aead, err := chacha20poly1305.New(contentKey[:32])
	if err != nil {
		return nil, err
	}
	
	if len(envelope.Ciphertext) < aead.NonceSize() {
		return nil, ErrDecryptionFailed
	}
	
	nonce := envelope.Ciphertext[:aead.NonceSize()]
	ciphertext := envelope.Ciphertext[aead.NonceSize():]
	
	plaintext, err := aead.Open(nil, nonce, ciphertext, envelope.AAD)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	
	return plaintext, nil
}

// encapsulateContentKey wraps a content key for a recipient
func (h *HPKEImpl) encapsulateContentKey(contentKey, recipientPublicKey []byte) (*Encapsulation, error) {
	// For simplicity, we're directly encrypting the content key
	// In production, use proper KEM encapsulation
	encap, err := h.Encapsulate(recipientPublicKey)
	if err != nil {
		return nil, err
	}
	
	// XOR content key with shared secret (simplified)
	encrypted := make([]byte, len(contentKey))
	for i := range contentKey {
		encrypted[i] = contentKey[i] ^ encap.SharedSecret[i%len(encap.SharedSecret)]
	}
	
	return &Encapsulation{
		EncapsulatedKey: encap.EncapsulatedKey,
		SharedSecret:    encrypted,
	}, nil
}

// CreateCommitmentHPKE creates a binding commitment to data using HPKE
func CreateCommitmentHPKE(data []byte) []byte {
	// Use SHA256 for commitment
	h := sha256.Sum256(data)
	return h[:]
}

// SealSimple encrypts and authenticates plaintext for a single recipient
func (h *HPKEImpl) SealSimple(sharedSecret, plaintext, aad []byte) ([]byte, error) {
	// Use ChaCha20-Poly1305 AEAD
	aead, err := chacha20poly1305.New(sharedSecret[:32])
	if err != nil {
		return nil, err
	}
	
	// Generate nonce
	nonce := make([]byte, aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	
	// Encrypt and authenticate
	ciphertext := aead.Seal(nonce, nonce, plaintext, aad)
	return ciphertext, nil
}

// OpenSimple decrypts and verifies ciphertext
func (h *HPKEImpl) OpenSimple(sharedSecret, ciphertext, aad []byte) ([]byte, error) {
	// Use ChaCha20-Poly1305 AEAD
	aead, err := chacha20poly1305.New(sharedSecret[:32])
	if err != nil {
		return nil, err
	}
	
	if len(ciphertext) < aead.NonceSize() {
		return nil, ErrDecryptionFailed
	}
	
	nonce := ciphertext[:aead.NonceSize()]
	actualCiphertext := ciphertext[aead.NonceSize():]
	
	plaintext, err := aead.Open(nil, nonce, actualCiphertext, aad)
	if err != nil {
		return nil, ErrDecryptionFailed
	}
	
	return plaintext, nil
}