package authenticator

import (
	"errors"

	"github.com/boltdb/bolt"

	"github.com/nanopack/logvac/config"
)

type (
	boltdb struct {
		db *bolt.DB
	}
)

func NewBoltDb(config string) (*boltdb, error) {
	d, err := bolt.Open(config, 0600, nil)
	if err != nil {
		return nil, err
	}
	b := boltdb{
		db: d,
	}

	return &b, nil
}

func (b boltdb) initialize() error {
	err := b.db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tokens"))
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (b boltdb) add(token string) error {
	// todo: doesn't fail to write if db gets deleted
	if b.db == nil {
		return errors.New("I need to be setup first")
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		return bucket.Put([]byte(token), []byte(token))
	})
}

func (b boltdb) remove(token string) error {
	if b.db == nil {
		return errors.New("I need to be setup first")
	}
	return b.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		return bucket.Delete([]byte(token))
	})
}

func (b boltdb) valid(token string) bool {
	if b.db == nil {
		return false
	}

	if token == "" {
		config.Log.Trace("Blank token invalid when using authenticator!")
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
