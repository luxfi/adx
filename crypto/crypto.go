// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package crypto

import (
	"crypto/ecdsa"
	"crypto/sha256"
	"errors"
	
	luxcrypto "github.com/luxfi/crypto"
	"github.com/luxfi/crypto/hashing"
)

var (
	ErrInvalidCiphertext = errors.New("invalid ciphertext")
)

// CreateCommitment creates a cryptographic commitment using luxfi's hashing
func CreateCommitment(data []byte) []byte {
	// Use luxfi's hashing
	hasher := hashing.ComputeHash256(data)
	return hasher
}

// HashData hashes data using SHA256
func HashData(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// EncryptWithHPKE encrypts data using HPKE
func EncryptWithHPKE(publicKey, plaintext []byte) ([]byte, error) {
	// For now, use a simple encryption
	// In production, would use full HPKE implementation
	if len(publicKey) != 32 {
		return nil, errors.New("invalid public key size")
	}
	// Placeholder: prepend public key to ciphertext
	ciphertext := append(publicKey, plaintext...)
	return ciphertext, nil
}

// DecryptWithHPKE decrypts data using HPKE
func DecryptWithHPKE(privateKey, ciphertext []byte) ([]byte, error) {
	// For simplicity, using a basic decryption
	// In production, would need full HPKE context with encapsulated key
	if len(ciphertext) < 32 {
		return nil, ErrInvalidCiphertext
	}
	// This is a simplified version - real HPKE would handle the encapsulated key properly
	return ciphertext[32:], nil // Placeholder - returns plaintext portion
}

// GenerateKeyPair generates a new key pair using luxfi's crypto
func GenerateKeyPair() (privateKey, publicKey []byte, err error) {
	// Use luxfi crypto for key generation
	privKey, err := luxcrypto.GenerateKey()
	if err != nil {
		return nil, nil, err
	}
	
	// Get public key bytes
	pubKeyBytes := luxcrypto.FromECDSAPub(&privKey.PublicKey)
	
	// Get private key bytes
	privKeyBytes := luxcrypto.FromECDSA(privKey)
	
	return privKeyBytes, pubKeyBytes, nil
}

// Sign signs a message with a private key
func Sign(privateKey, message []byte) ([]byte, error) {
	// Convert bytes to ECDSA private key
	privKey, err := luxcrypto.ToECDSA(privateKey)
	if err != nil {
		return nil, err
	}
	
	// Hash the message
	hash := luxcrypto.Keccak256(message)
	
	// Sign the hash
	return luxcrypto.Sign(hash, privKey)
}

// Verify verifies a signature
func Verify(publicKey, message, signature []byte) bool {
	// Hash the message
	hash := luxcrypto.Keccak256(message)
	
	// Verify signature (remove recovery ID if present)
	if len(signature) > 64 {
		signature = signature[:64]
	}
	
	return luxcrypto.VerifySignature(publicKey, hash, signature)
}

// RecoverPublicKey recovers the public key from a signature
func RecoverPublicKey(hash, signature []byte) (*ecdsa.PublicKey, error) {
	return luxcrypto.SigToPub(hash, signature)
}