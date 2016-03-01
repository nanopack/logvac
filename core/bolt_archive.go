package logvac

import (
	"bytes"
	"encoding/binary"
	"encoding/json"

	"github.com/boltdb/bolt"

	"github.com/nanopack/logvac/config"
)

type (
	BoltArchive struct {
		DB            *bolt.DB
		MaxBucketSize uint64
	}
)

func NewBoltArchive(path string) (*BoltArchive, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	archive := BoltArchive{
		DB:            db,
		MaxBucketSize: 1000,
	}

	return &archive, nil
}

func (archive *BoltArchive) Slice(name string, offset, limit uint64, level int) ([]Message, error) {
	var messages []Message
	err := archive.DB.View(func(tx *bolt.Tx) error {
		messages = make([]Message, 0)
		bucket := tx.Bucket([]byte(name))

		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		k, _ := c.First()
		if k == nil {
			return nil
		}

		// I need to skip to the correct id
		initial := &bytes.Buffer{}
		if err := binary.Write(initial, binary.BigEndian, offset); err != nil {
			return err
		}

		c.Seek(initial.Bytes())
		for k, v := c.First(); k != nil && limit > 0; k, v = c.Next() {
			msg := Message{}
			if err := json.Unmarshal(v, &msg); err != nil {
				config.Log.Error("Couldn't unmarshal message") // todo: remove
				return err
			}
			if msg.Priority >= level {
				limit--
				messages = append(messages, msg)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	config.Log.Trace("Messages: %v", messages)
	return messages, nil
}

func (archive *BoltArchive) Write(log Logger, msg Message) {
	config.Log.Trace("Bolt archive writing...")
	err := archive.DB.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(msg.Type))
		if err != nil {
			return err
		}

		value, err := json.Marshal(msg)
		if err != nil {
			return err
		}

		// this needs to ensure lexographical order
		key := &bytes.Buffer{}
		nextLine, err := bucket.NextSequence()
		if err != nil {
			return err
		}

		if err = binary.Write(key, binary.BigEndian, nextLine); err != nil {
			return err
		}
		if err = bucket.Put(key.Bytes(), value); err != nil {
			return err
		}

		// trim the bucket to size
		c := bucket.Cursor()
		c.First()

		// KeyN does not take into account the new value added
		for key_count := uint64(bucket.Stats().KeyN) + 1; key_count > archive.MaxBucketSize; key_count-- {
			c.Delete()
			c.Next()
		}

		// I don't know how to do this other then scanning the collection periodically.
		// delete entries that are older then needed
		// c.First()
		// for {
		// 	c.Delete()
		// 	c.Next()
		// }

		return nil
	})

	if err != nil {
		config.Log.Error("Historical write failed") // todo: remove
		log.Error("[Historical][write]" + err.Error())
	}
}
