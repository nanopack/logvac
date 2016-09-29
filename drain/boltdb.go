package drain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"

	"github.com/boltdb/bolt"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

var (
	// How often to clean, exported for testing
	CleanFreq = 60
)

type (
	// BoltArchive is a boltDB archiver
	BoltArchive struct {
		db   *bolt.DB
		Done chan bool
	}
)

// NewBoltArchive creates a new boltDB archiver
func NewBoltArchive(path string) (*BoltArchive, error) {
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return nil, err
	}
	d, err := bolt.Open(path, 0644, nil)
	if err != nil {
		return nil, err
	}

	archive := BoltArchive{
		db:   d,
		Done: make(chan bool),
	}

	return &archive, nil
}

// Init initializes the archiver drain
func (a BoltArchive) Init() error {
	// add drain
	logvac.AddDrain("historical", a.Write)

	return nil
}

// Close closes the bolt db
func (a BoltArchive) Close() {
	a.db.Close()
}

// Slice returns a slice of logs based on the name, offset, limit, and log-level
func (a BoltArchive) Slice(name, host, tag string, offset, end, limit int64, level int) ([]logvac.Message, error) {
	var messages []logvac.Message

	err := a.db.View(func(tx *bolt.Tx) error {
		messages = make([]logvac.Message, 0)
		bucket := tx.Bucket([]byte(name))

		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		last, _ := c.Last()
		if last == nil {
			return nil
		}

		// prepare to skip to the correct id
		initial := &bytes.Buffer{}
		if offset == 0 {
			// if no offset value is given, start with last log
			initial.Write(last)
		} else {
			// otherwise, start at their offset
			if err := binary.Write(initial, binary.BigEndian, offset); err != nil {
				return err
			}
		}

		// prepare to end at the specified time (pagination limits still apply)
		final := &bytes.Buffer{}
		if err := binary.Write(final, binary.BigEndian, end); err != nil {
			return err
		}

		// seek boltdb cursor to initial offset
		k, v := c.Seek(initial.Bytes())

		// if the record's utime (k) doesn't match the specified "initial" value, use previous record.
		// note: https://github.com/boltdb/bolt/blob/v1.2.0/cursor.go#L114 explains why.
		// (this step may not be needed if the order of logs returned is reversed)
		if string(k) != initial.String() {
			k, v = c.Prev()
		}

		// todo: make limit be len(bucket)? if limit < 0
		for ; k != nil && limit > 0; k, v = c.Prev() {
			msg := logvac.Message{}
			if err := json.Unmarshal(v, &msg); err != nil {
				return fmt.Errorf("Couldn't unmarshal message - %v", err)
			}
			// if specified end is reached, be done
			if string(k) == final.String() {
				limit = 0
			}
			if msg.Priority >= level {
				if msg.Id == host || host == "" {
					// todo: negate here if tag[0] == "!"
					if msg.Tag == tag || tag == "" {
						limit--
						// prepend messages with new message (display newest last)
						messages = append([]logvac.Message{msg}, messages...)
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// config.Log.Trace("Messages: %v", messages)
	return messages, nil
}

// Write writes the message to database
func (a BoltArchive) Write(msg logvac.Message) {
	config.Log.Trace("Bolt archive writing...")
	err := a.db.Batch(func(tx *bolt.Tx) error {
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
		if err = binary.Write(key, binary.BigEndian, msg.UTime); err != nil {
			return err
		}
		if err = bucket.Put(key.Bytes(), value); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		config.Log.Error("Historical write failed - %v", err.Error())
	}
}

// Expire cleans up old logs
func (a BoltArchive) Expire() {
	// if log-keep is "" expire is disabled
	if config.LogKeep == "" {
		return
	}

	var logKeep map[string]interface{}
	err := json.Unmarshal([]byte(config.LogKeep), &logKeep)
	if err != nil {
		config.Log.Fatal("Bad JSON syntax for log-keep - %v", err)
		os.Exit(1) // maybe not?
	}

	// clean up every minute // todo: maybe 5mins?
	tick := time.Tick(time.Duration(CleanFreq) * time.Second)
	for {
		select {
		case <-tick:
			for k, v := range logKeep { // todo: maybe rather/also loop through buckets
				switch v.(type) {
				case string:
					var expireTime = time.Now().UnixNano()

					r, _ := regexp.Compile("([0-9]+)([a-zA-Z]+)")
					var (
						NANO_MIN  int64 = 60000000000
						NANO_HOUR int64 = NANO_MIN * 60
						NANO_DAY  int64 = NANO_HOUR * 24
						NANO_WEEK int64 = NANO_DAY * 7
						NANO_YEAR int64 = NANO_WEEK * 52
						duration  int64 = NANO_WEEK * 2
						NANO_SEC  int64 = NANO_MIN / 60
					)

					match := r.FindStringSubmatch(v.(string)) // "2w"
					if len(match) == 3 {
						number, err := strconv.ParseInt(match[1], 0, 64)
						if err != nil {
							config.Log.Fatal("Bad log-keep - %v", err)
							number = 2
						}
						switch match[2] {
						case "s": // second // for testing
							duration = NANO_SEC * number
						case "m": // minute
							duration = NANO_MIN * number
						case "h": // hour
							duration = NANO_HOUR * number
						case "d": // day
							duration = NANO_DAY * number
						case "w": // week
							duration = NANO_WEEK * number
						case "y": // year
							duration = NANO_YEAR * number
						default: // 2 weeks
							config.Log.Debug("Keeping '%v' logs for 2 weeks", k)
							duration = NANO_WEEK * 2
						}
					}

					expireTime = expireTime - duration

					a.db.Update(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(k))
						if bucket == nil {
							config.Log.Trace("No logs of type '%s' found", k)
							return fmt.Errorf("No logs of type '%s' found", k)
						}

						c := bucket.Cursor()

						// loop through and remove outdated logs
						for kk, vv := c.First(); kk != nil; kk, vv = c.Next() {
							var logMessage logvac.Message
							err := json.Unmarshal([]byte(vv), &logMessage)
							if err != nil {
								config.Log.Fatal("Bad JSON syntax in log message - %v", err)
							}
							if logMessage.UTime < expireTime {
								config.Log.Trace("Deleting expired log of type '%v'...", k)
								err = c.Delete()
								if err != nil {
									config.Log.Trace("Failed to delete expired log - %v", err)
								}
							}
						}
						return nil
					}) // db.Update
				case float64:
					// todo: maybe View, then Update within and remove only those marked records?
					a.db.Update(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(k))
						if bucket == nil {
							config.Log.Trace("No logs of type '%s' found", k)
							return fmt.Errorf("No logs of type '%s' found", k)
						}

						// trim the bucket to size
						c := bucket.Cursor()
						c.First()

						// loop through and remove extra logs
						for key_count := float64(bucket.Stats().KeyN); key_count > v.(float64); key_count-- {
							config.Log.Trace("Deleting extra log of type '%v'...", k)
							err = c.Delete()
							if err != nil {
								config.Log.Trace("Failed to delete extra log - %v", err)
							}
							c.Next()
						}
						return nil
					}) // db.Update
				default:
					config.Log.Fatal("Bad log-keep value")
					os.Exit(1)
				}
			} // range logKeep
		case <-a.Done:
			return
		}
	}
}
