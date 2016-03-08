package drain

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/boltdb/bolt"

	"github.com/nanopack/logvac/config"
	"github.com/nanopack/logvac/core"
)

type (
	BoltArchive struct {
		dbAddr        string
		MaxBucketSize uint64
	}
)

func NewBoltArchive(path string) (*BoltArchive, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	archive := BoltArchive{
		dbAddr:        path,
		MaxBucketSize: 100000, // this should be configurable
	}

	return &archive, nil
}

func (archive *BoltArchive) Init() error {
	// add drain
	logvac.AddDrain("historical", archive.Write)

	return nil
}

func (archive *BoltArchive) Slice(name, host, tag string, offset, limit uint64, level int) ([]logvac.Message, error) {
	db, err := bolt.Open(archive.dbAddr, 0600, nil)
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var messages []logvac.Message

	err = db.View(func(tx *bolt.Tx) error {
		messages = make([]logvac.Message, 0)
		bucket := tx.Bucket([]byte(name))

		if bucket == nil {
			return nil
		}
		c := bucket.Cursor()
		k, _ := c.First()
		if k == nil {
			return nil
		}

		// skip to the correct id
		initial := &bytes.Buffer{}
		if err := binary.Write(initial, binary.BigEndian, offset); err != nil {
			return err
		}

		c.Seek(initial.Bytes())
		for k, v := c.First(); k != nil && limit > 0; k, v = c.Next() {
			msg := logvac.Message{}
			if err := json.Unmarshal(v, &msg); err != nil {
				return fmt.Errorf("Couldn't unmarshal message - %v", err)
			}
			if msg.Priority >= level {
				if msg.Id == host || host == "" {
					if msg.Tag == tag || tag == "" {
						limit--
						messages = append(messages, msg)
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

func (archive *BoltArchive) Write(msg logvac.Message) {
	db, err := bolt.Open(archive.dbAddr, 0600, nil)
	if err != nil {
		config.Log.Error("Historical db open failed - %v", err.Error())
		return
	}
	defer db.Close()

	config.Log.Trace("Bolt archive writing...")
	err = db.Update(func(tx *bolt.Tx) error {
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

		return nil
	})

	if err != nil {
		config.Log.Error("Historical write failed - %v", err.Error())
	}
}

func (archive *BoltArchive) Expire() {
	// if log-keep is "" expire is disabled
	if config.LogKeep != "" {
		var logKeep map[string]interface{}
		err := json.Unmarshal([]byte(config.LogKeep), &logKeep)
		if err != nil {
			config.Log.Fatal("Bad JSON syntax for log-keep - %v", err)
			os.Exit(1) // maybe not?
		}

		// clean up every minute // todo: maybe 5mins?
		tick := time.Tick(time.Minute)
		for _ = range tick {
			for k, v := range logKeep {
				switch v.(type) {
				case string:
					var logTime = time.Now().UnixNano()

					r, _ := regexp.Compile("([0-9]+)([a-zA-Z]+)")
					var (
						NANO_MIN  int64 = 60000000000
						NANO_HOUR int64 = NANO_MIN * 60
						NANO_DAY  int64 = NANO_HOUR * 24
						NANO_WEEK int64 = NANO_DAY * 7
						NANO_YEAR int64 = NANO_WEEK * 52
						duration  int64 = NANO_WEEK * 2
					)

					match := r.FindStringSubmatch(v.(string)) // "2w"
					if len(match) == 3 {
						number, err := strconv.ParseInt(match[1], 0, 64)
						if err != nil {
							config.Log.Fatal("Bad log-keep - %v", err)
							number = 2
						}
						switch match[2] {
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

					logTime = logTime - duration

					db, err := bolt.Open(archive.dbAddr, 0600, nil)
					if err != nil {
						config.Log.Error("Historical db open failed - %v", err.Error())
						return
					}

					db.Update(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(k))
						if bucket == nil {
							config.Log.Debug("No logs of type '%s' found", k)
						}

						c := bucket.Cursor()

						// loop through and remove outdated logs, unfortunately flocks.. == if many logs, more time before new log can be written
						for kk, vv := c.First(); kk != nil; kk, vv = c.Next() {
							var logMessage logvac.Message
							err := json.Unmarshal([]byte(vv), &logMessage)
							if err != nil {
								config.Log.Fatal("Bad JSON syntax in log message - %v", err)
							}
							if logMessage.UTime < logTime {
								config.Log.Trace("Deleting expired log of type '%v'...", k)
								err = c.Delete()
								if err != nil {
									config.Log.Trace("Failed to delete expired log - %v", err)
								}
							}
						}
						return nil
					}) // db.Update

					db.Close()
				case float64:
					db, err := bolt.Open(archive.dbAddr, 0600, nil)
					if err != nil {
						config.Log.Error("Historical db open failed - %v", err.Error())
						return
					}

					// todo: maybe View, then Update within and remove only those marked records?
					db.Update(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(k))
						if bucket == nil {
							config.Log.Debug("No logs of type '%s' found", k)
						}

						// trim the bucket to size
						c := bucket.Cursor()
						c.First()

						// loop through and remove extra logs, unfortunately flocks.. == if many logs, more time before new log can be written
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
					db.Close()
				default:
					config.Log.Fatal("Bad log-keep value")
					os.Exit(1)
				}
			} // range logKeep
		} // range tick
	}
}
