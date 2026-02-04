// Copyright (C) 2025, ADXYZ Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package storage

import (
	"github.com/luxfi/database"
	"github.com/luxfi/database/badgerdb"
	"github.com/luxfi/database/memdb"
)

// Storage wraps luxfi's database interface
type Storage struct {
	db database.Database
}

// NewStorage creates a new storage instance using luxfi/database
func NewStorage(dbType string, path string) (*Storage, error) {
	var db database.Database
	var err error

	switch dbType {
	case "memory":
		db = memdb.New()
	case "badger":
		db, err = badgerdb.New(path, nil, "", nil)
		if err != nil {
			return nil, err
		}
	default:
		// Default to badger
		db, err = badgerdb.New(path, nil, "", nil)
		if err != nil {
			return nil, err
		}
	}

	return &Storage{db: db}, nil
}

// Put stores a key-value pair
func (s *Storage) Put(key, value []byte) error {
	return s.db.Put(key, value)
}

// Get retrieves a value by key
func (s *Storage) Get(key []byte) ([]byte, error) {
	return s.db.Get(key)
}

// Has checks if a key exists
func (s *Storage) Has(key []byte) (bool, error) {
	return s.db.Has(key)
}

// Delete removes a key-value pair
func (s *Storage) Delete(key []byte) error {
	return s.db.Delete(key)
}

// NewBatch creates a new batch for atomic operations
func (s *Storage) NewBatch() database.Batch {
	return s.db.NewBatch()
}

// NewIterator creates an iterator
func (s *Storage) NewIterator() database.Iterator {
	return s.db.NewIterator()
}

// NewIteratorWithPrefix creates an iterator with a key prefix
func (s *Storage) NewIteratorWithPrefix(prefix []byte) database.Iterator {
	return s.db.NewIteratorWithPrefix(prefix)
}

// Close closes the database
func (s *Storage) Close() error {
	return s.db.Close()
}

// Compact compacts the underlying database
func (s *Storage) Compact(start, limit []byte) error {
	return s.db.Compact(start, limit)
}

// GetDatabase returns the underlying database
func (s *Storage) GetDatabase() database.Database {
	return s.db
}
