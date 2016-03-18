package authenticator

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/boltdb/bolt"

	"github.com/nanopack/logvac/config"
)

type (
	boltdb struct {
		dbAddr string
	}
)

// NewBoltDb creates a new boltdb (currently, it is critical to set dbAddr)
func NewBoltDb(config string) (*boltdb, error) {
	err := os.MkdirAll(filepath.Dir(config), 0755)
	if err != nil {
		return nil, err
	}
	d, err := bolt.Open(config, 0600, nil)
	if err != nil {
		return nil, err
	}

	// reconnect to db so we can export under a running logvac (todo: maybe?)
	defer d.Close()
	b := boltdb{
		dbAddr: config,
	}

	return &b, nil
}

func (b boltdb) initialize() error {
	if b.dbAddr == "" {
		return errors.New("I need to be setup first")
	}

	db, err := bolt.Open(b.dbAddr, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("tokens"))
		return err
	})
	if err != nil {
		return err
	}

	return nil
}

func (b boltdb) add(token string) error {
	if b.dbAddr == "" {
		return errors.New("I need to be setup first")
	}

	db, err := bolt.Open(b.dbAddr, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		return bucket.Put([]byte(token), []byte(token))
	})
}

func (b boltdb) remove(token string) error {
	if b.dbAddr == "" {
		return errors.New("I need to be setup first")
	}

	db, err := bolt.Open(b.dbAddr, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	return db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		return bucket.Delete([]byte(token))
	})
}

func (b boltdb) valid(token string) bool {
	if b.dbAddr == "" {
		return false
	}

	db, err := bolt.Open(b.dbAddr, 0600, nil)
	if err != nil {
		return false
	}
	defer db.Close()

	if token == "" {
		config.Log.Trace("Blank token invalid when using authenticator!")
		return false
	}

	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		if string(bucket.Get([]byte(token))) != token {
			return errors.New("no match")
		}
		return nil
	})
	return err == nil
}

func (b boltdb) exportLogvac(exportWriter io.Writer) error {
	if b.dbAddr == "" {
		return errors.New("I need to be setup first")
	}

	db, err := bolt.Open(b.dbAddr, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	var tokens []string

	// get all tokens
	err = db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("tokens"))
		if bucket == nil {
			return errors.New("No tokens found")
		}

		c := bucket.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			tokens = append(tokens, string(v))
		}

		return nil
	})

	// byte-ify tokens
	tkns, err := json.Marshal(tokens)
	if err != nil {
		return nil
	}

	// write tokens
	_, err = exportWriter.Write(tkns)
	if err != nil {
		return fmt.Errorf("Failed to write - %v", err)
	}

	return nil
}

func (b boltdb) importLogvac(importReader io.Reader) error {
	if b.dbAddr == "" {
		return errors.New("I need to be setup first")
	}
	var tokens []string
	// limit to ~10 mb
	tkns := make([]byte, 10240)
	_, err := importReader.Read(tkns)
	if err != nil {
		return fmt.Errorf("Failed to read - %v", err)
	}
	// clean up 0 bytes
	tkns = bytes.Trim(tkns, "\x00")

	err = json.Unmarshal(tkns, &tokens)
	if err != nil {
		return fmt.Errorf("Failed to unmarshal - %v", err)
	}

	db, err := bolt.Open(b.dbAddr, 0600, nil)
	if err != nil {
		return err
	}
	defer db.Close()

	for _, token := range tokens {
		err = db.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("tokens"))
			err = bucket.Put([]byte(token), []byte(token))
			if err != nil {
				return fmt.Errorf("Failed to put token - %v", err)
			}
			return nil
		})
		if err != nil {
			return fmt.Errorf("Failed to update db - %v", err)
		}
	}

	return nil
}
