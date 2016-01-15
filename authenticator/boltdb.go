package authenticator

import (
	"errors"
	"github.com/boltdb/bolt"
)

type (
	boltdb struct{
		db *bolt.DB
	}
)

func init() {
	Register("boltdb", boltdb{})
}

func (b boltdb) Setup(config string) error {
	db, err := bolt.Open(config, 0600, nil)
	if err != nil {
		return err
	}

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tokens"))
		return err
	})
	if err != nil {
		return err
	}

	b.db = db

	return nil
}

func (b boltdb) Add(token string) error {
	if b.db == nil {
		return errors.New("I need to be setup first")
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		return bucket.Put([]byte(token), []byte(token))
	})
}

func (b boltdb) Remove(token string) error {
	if b.db == nil {
		return errors.New("I need to be setup first")
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		return bucket.Delete([]byte(token))
	})
}

func (b boltdb) Valid(token string) bool {
	if b.db == nil {
		return false
	}
	err := b.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		if string(bucket.Get([]byte(token))) != token {
			return errors.New("no match")
		}
		return nil
	})
	return err == nil
}
