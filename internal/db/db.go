package db

import (
	"encoding/json"
	"time"

	"go.etcd.io/bbolt"
)

const (
	defaultPerm = 0600
	defaultName = "tod.db"
)

type Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	Expiry       time.Time `json:"expiry"`
	TokenType    string    `json:"token_type"`
}

// DB is a wrapper around a *bbolt.DB.
type DB struct {
	db *bbolt.DB
}

// Open opens a database file and returns a *DB.
func Open() (*DB, error) {
	db, err := bbolt.Open(defaultName, defaultPerm, nil)
	if err != nil {
		return nil, err
	}

	return &DB{db}, nil
}

// Close closes a DB.
func (db *DB) Close() {
	db.db.Close()
}

// SetToken sets a Spotify OAuth token.
func (db *DB) SetToken(id string, token *Token) error {
	tx, err := db.db.Begin(true)
	if err != nil {
		return err
	}

	defer tx.Rollback() //nolint:errcheck

	tokens, err := tx.CreateBucketIfNotExists([]byte("tokens"))
	if err != nil {
		return err
	}

	tok, err := json.Marshal(token)
	if err != nil {
		return err
	}

	if err := tokens.Put([]byte(id), tok); err != nil {
		return err
	}

	return tx.Commit()
}

// Token returns a Spotify OAuth token for a user.
func (db *DB) Token(id string) (*Token, error) {
	tx, err := db.db.Begin(false)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback() //nolint:errcheck

	tokens := tx.Bucket([]byte("tokens"))
	if tokens == nil {
		return nil, nil
	}

	buk := tokens.Get([]byte(id))
	if buk == nil {
		return nil, nil
	}

	var t *Token
	if err := json.Unmarshal(buk, &t); err != nil {
		return nil, err
	}

	return t, nil
}
