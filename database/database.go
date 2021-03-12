package database

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v2"
	"time"
)

type Database struct {
	db *badger.DB
}

func New(path string) (*Database, error) {
	var res = &Database{}

	var err error
	if res.db, err = badger.Open(badger.DefaultOptions(path)); err != nil {
		return nil, err
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
		again:
			err := res.db.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	}()

	return res, nil
}

func (db *Database) Get(key string, handler func(entry []byte) error) error {
	return db.db.View(func(tx *badger.Txn) error {
		item, err := tx.Get([]byte(key))
		if err != nil {
			return err
		}

		err = item.Value(handler)
		if err != nil {
			return err
		}
		return nil
	})
}

func (db *Database) Set(key string, value interface{}) error {
	var raw []byte
	var err error

	if raw, err = json.Marshal(value); err != nil {
		return err
	}

	return db.db.Update(func(tx *badger.Txn) error {
		key := []byte(key)
		entry := badger.NewEntry(key, raw)
		return tx.SetEntry(entry)
	})
}

func (db *Database) ForEach(prefix string, handler func(entry []byte) error) error {
	return db.db.View(func(tx *badger.Txn) (err error) {
		iter := tx.NewIterator(badger.DefaultIteratorOptions)
		defer iter.Close()

		bprefix := []byte(prefix)
		for iter.Seek(bprefix); iter.ValidForPrefix(bprefix); iter.Next() {
			item := iter.Item()
			if err = item.Value(func(val []byte) (err error) {
				return handler(val)
			}); err != nil {
				return
			}
		}

		return
	})
}

func (db *Database) Update(key string, handler func([]byte) []byte) error {
	for {

		err := db.db.Update(func(tx *badger.Txn) error {
			var err error
			var item *badger.Item

			if item, err = tx.Get([]byte(key)); err != nil {
				return err
			}

			var copied []byte
			if copied, err = item.ValueCopy(nil); err != nil {
				return err
			}

			if modified := handler(copied); modified != nil {
				return tx.Set([]byte(key), modified)
			}

			return nil
		})

		if err != badger.ErrConflict {
			return err
		}

		// other goroutine has changed the key while our Update operation was running
		// read value again and perform update on the new value
	}
}