// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package ids

import (
	"encoding/hex"
	"fmt"
)

// NodeIDLen is the length of a NodeID in bytes
const NodeIDLen = 20

// NodeID is a unique identifier for a node
type NodeID [NodeIDLen]byte

// EmptyNodeID is an empty NodeID
var EmptyNodeID = NodeID{}

// String returns the string representation of a NodeID
func (id NodeID) String() string {
	return hex.EncodeToString(id[:])
}

// IsEmpty returns true if the NodeID is empty
func (id NodeID) IsEmpty() bool {
	return id == NodeID{}
}

// Bytes returns the byte representation of a NodeID
func (id NodeID) Bytes() []byte {
	return id[:]
}

// NodeIDFromString parses a NodeID from a hex string
func NodeIDFromString(s string) (NodeID, error) {
	var id NodeID
	b, err := hex.DecodeString(s)
	if err != nil {
		return id, err
	}
	if len(b) != NodeIDLen {
		return id, fmt.Errorf("invalid NodeID length: expected %d, got %d", NodeIDLen, len(b))
	}
	copy(id[:], b)
	return id, nil
}

// NodeIDFromBytes creates a NodeID from bytes
func NodeIDFromBytes(b []byte) (NodeID, error) {
	var id NodeID
	if len(b) != NodeIDLen {
		return id, fmt.Errorf("invalid NodeID length: expected %d, got %d", NodeIDLen, len(b))
	}
	copy(id[:], b)
	return id, nil
}

// GenerateNodeID generates a new random NodeID
func GenerateNodeID() NodeID {
	var id NodeID
	testID := GenerateTestID()
	copy(id[:], testID[:NodeIDLen])
	return id
}
