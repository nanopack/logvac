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
func (a *BoltArchive) Init() error {
	// add drain
	logvac.AddDrain("historical", a.Write)

	return nil
}

// Close closes the bolt db
func (a *BoltArchive) Close() {
	err := a.db.Close()
	if err != nil {
		config.Log.Error("Faile to close bolt - %s", err.Error())
	}
}

// Slice returns a slice of logs based on the name, offset, limit, and log-level
func (a *BoltArchive) Slice(name, host string, tag []string, offset, end, limit int64, level int) ([]logvac.Message, error) {
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
			oMsg := logvac.OldMessage{} // old message (type has changed for multi-tenancy)
			// if specified end is reached, be done
			if string(k) == final.String() {
				limit = 0
			}

			// unmarshal to check if match.. seems expensive
			if err := json.Unmarshal(v, &msg); err != nil {
				// for backwards compatibility (needed for approx 2 weeks only until old logs get cleaned up)
				if err2 := json.Unmarshal(v, &oMsg); err2 != nil {
					return fmt.Errorf("Couldn't unmarshal message - %s - %s", err, err2)
				}
				// convert old message to new message for saving
				msg.Time = oMsg.Time
				msg.UTime = oMsg.UTime
				msg.Id = oMsg.Id
				msg.Tag = []string{oMsg.Tag}
				msg.Type = oMsg.Type
				msg.Priority = oMsg.Priority
				msg.Content = oMsg.Content

				// return fmt.Errorf("Couldn't unmarshal message - %s", err)
			}

			if msg.Priority >= level {
				if host == "" || msg.Id == host {
					// todo: negate here if tag starts with "!"
					if len(tag) == 0 {
						limit--

						// prepend messages with new message (display newest last)
						messages = append([]logvac.Message{msg}, messages...)
					} else {
						for x := range msg.Tag {
							for y := range tag {
								if tag[y] == "" || msg.Tag[x] == tag[y] {
									limit--

									// prepend messages with new message (display newest last)
									messages = append([]logvac.Message{msg}, messages...)

									return nil
								}
							}
						}
					}
				}
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// config.Log.Trace("Messages: %+q", messages)
	return messages, nil
}

// Write writes the message to database
func (a *BoltArchive) Write(msg logvac.Message) {
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
		config.Log.Error("Historical write failed - %s", err)
	}
}

// Expire cleans up old logs by date or volume of logs
func (a *BoltArchive) Expire() {
	// if log-keep is "" expire is disabled
	if config.LogKeep == "" {
		config.Log.Debug("Log expiration disabled")
		return
	}

	var logKeep map[string]interface{}
	err := json.Unmarshal([]byte(config.LogKeep), &logKeep)
	if err != nil {
		config.Log.Fatal("Bad JSON syntax for log-keep - %s, saving logs indefinitely", err)
		return
	}

	if config.CleanFreq < 1 {
		config.CleanFreq = 60
	}

	config.Log.Trace("LogKeep - %v; CleanFreq - %d", logKeep, config.CleanFreq)

	// clean up every minute // todo: maybe 5mins?
	tick := time.Tick(time.Duration(config.CleanFreq) * time.Second)

	r, _ := regexp.Compile("([0-9]+)([a-za-z]+)")
	var (
		NANO_MIN  int64 = 60000000000
		NANO_HOUR int64 = NANO_MIN * 60
		NANO_DAY  int64 = NANO_HOUR * 24
		NANO_WEEK int64 = NANO_DAY * 7
		NANO_YEAR int64 = NANO_WEEK * 52
		duration  int64 = NANO_WEEK * 2
		NANO_SEC  int64 = NANO_MIN / 60
	)

	for {
		select {
		case <-tick:
			for bucketName, saveAmt := range logKeep { // todo: maybe rather/also loop through buckets
				config.Log.Trace("bucketName - %s; saveAmt - %v", bucketName, saveAmt)
				switch saveAmt.(type) {
				case string:
					var expireTime = time.Now().UnixNano()

					match := r.FindStringSubmatch(saveAmt.(string)) // "2w"
					config.Log.Trace("SaveAmt - %v, match - %v", saveAmt, match)
					if len(match) == 3 {
						number, err := strconv.ParseInt(match[1], 0, 64)
						if err != nil {
							config.Log.Fatal("Bad log-keep - %s", err)
							number = 2
						}
						switch match[2] {
						case "s": // second // for testing
							config.Log.Debug("Keeping logs for %d seconds", number)
							duration = NANO_SEC * number
						case "m": // minute
							config.Log.Debug("Keeping logs for %d minutes", number)
							duration = NANO_MIN * number
						case "h": // hour
							config.Log.Debug("Keeping logs for %d hours", number)
							duration = NANO_HOUR * number
						case "d": // day
							config.Log.Debug("Keeping logs for %d days", number)
							duration = NANO_DAY * number
						case "w": // week
							config.Log.Debug("Keeping logs for %d weeks", number)
							duration = NANO_WEEK * number
						case "y": // year
							config.Log.Debug("Keeping logs for %d years", number)
							duration = NANO_YEAR * number
						default: // 2 weeks
							config.Log.Debug("Keeping '%s' logs for 2 weeks", bucketName)
							duration = NANO_WEEK * 2
						}
					}

					expireTime = expireTime - duration
					eTime := &bytes.Buffer{}
					if err := binary.Write(eTime, binary.BigEndian, expireTime); err != nil {
						config.Log.Error("Failed to convert expire time to binary - %s", err.Error())
						continue
					}

					config.Log.Debug("Starting age cleanup batch...")
					a.db.Batch(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(bucketName))
						if bucket == nil {
							config.Log.Debug("No logs of type '%s' found", bucketName)
							return fmt.Errorf("No logs of type '%s' found", bucketName)
						}

						c := bucket.Cursor()

						var err error

						// loop through and remove outdated logs
						for k, _ := c.First(); k != nil; k, _ = c.Next() {
							// if logMessage.UTime < expireTime {
							if bytes.Compare(k, eTime.Bytes()) == -1 {
								config.Log.Trace("Deleting expired log of type '%s'...", bucketName)
								err = c.Delete()
								if err != nil {
									config.Log.Debug("Failed to delete expired log - %s", err)
								}
								config.Log.Trace("Deleted log")
							} else { // don't continue looping through newer logs (resource/file-lock hog)
								config.Log.Trace("Done with old logs")
								break
							}
						}

						config.Log.Debug("=======================================")
						config.Log.Debug("= DONE CHECKING/DELETING EXPIRED LOGS =")
						config.Log.Debug("=======================================")
						return nil
					})
					config.Log.Trace("Done defining batch")
				case float64, int:
					records := int(saveAmt.(float64)) // assertion is slow, do it once (casting is fast)

					config.Log.Debug("Starting record cleanup batch...")
					a.db.Batch(func(tx *bolt.Tx) error {
						bucket := tx.Bucket([]byte(bucketName))
						if bucket == nil {
							config.Log.Trace("No logs of type '%s' found", bucketName)
							return fmt.Errorf("No logs of type '%s' found", bucketName)
						}

						// trim the bucket to size
						c := bucket.Cursor()

						rSaved := 0
						// loop through and remove extra logs
						// if we ever stop ordering by time (oldest first) we'll need to change cursor placement
						for k, v := c.Last(); k != nil && v != nil; k, v = c.Prev() {
							rSaved += 1
							// if the number records we've traversed is larger than our limit, delet the current record
							if rSaved > records {
								config.Log.Trace("Deleting extra log of type '%s'...", bucketName)
								err = c.Delete()
								if err != nil {
									config.Log.Trace("Failed to delete extra log - %s", err)
								}
							}
						}

						config.Log.Debug("=======================================")
						config.Log.Debug("= DONE CHECKING/DELETING EXPIRED LOGS =")
						config.Log.Debug("=======================================")
						return nil
					})
				default:
					// todo: we should pre-parse these values and exit on startup, not x minutes into running
					config.Log.Fatal("Bad log-keep value")
					os.Exit(1)
				}
			} // range logKeep
		case <-a.Done:
			config.Log.Debug("Done recieved on channel. (Cleanup halting)")
			return
		}
	}
}

// Save writes a value to the database
func (a *BoltArchive) Save(db, key string, v interface{}) error {
	config.Log.Trace("Saving...")

	err := a.db.Batch(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(db))
		if err != nil {
			return err
		}

		value, err := json.Marshal(v)
		if err != nil {
			return fmt.Errorf("Failed to marshal - %s", err.Error())
		}

		if err = bucket.Put([]byte(key), value); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		err = fmt.Errorf("Save failed - %s", err)
	}

	return err
}

// Get gets values from the database
func (a *BoltArchive) Get(db, key string, v interface{}) error {
	// get all configs
	err := a.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(db))
		if bucket == nil {
			return fmt.Errorf("No bucket found")
		}

		value := bucket.Get([]byte(key))
		err := json.Unmarshal(value, &v)
		if err != nil {
			return fmt.Errorf("Bad JSON in stored drain config - %s", err.Error())
		}

		return nil
	})

	if err != nil {
		err = fmt.Errorf("Failed to get - %s", err)
	}

	return err
}
