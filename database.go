package main

import (
	"encoding/json"
	badger "github.com/dgraph-io/badger/v2"
	"time"
)

var (
	Database *badger.DB
)

func initDatabase () {
	var err error
	if Database, err = badger.Open(badger.DefaultOptions(config.DatabasePath)); err != nil {
		panic(err)
	}

	go func() {
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
		again:
			err := Database.RunValueLogGC(0.7)
			if err == nil {
				goto again
			}
		}
	} ()
}

func DBforEach (prefix string, handler func(entry []byte) error ) error {
	return Database.View(func(tx *badger.Txn) (err error) {
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

func DBStore(key string, value interface{}) (err error) {
	var raw []byte
	if raw, err = json.Marshal(value); err != nil {
		return err
	}

	return Database.Update(func(tx *badger.Txn) error {
		key     := []byte(key)
		entry   := badger.NewEntry(key, raw)
		return tx.SetEntry(entry)
	})
}
