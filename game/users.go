package game

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v2"
	"log"
)

type UID string

type User struct {
	Taken uint64 `json:"taken"`
	Paid  uint64 `json:"paid"`
	Out   uint64 `json:"outgoing"`
}

func (user *User) AvailableReward () uint64 {
	var out = user.Paid + user.Out

	if user.Taken >= out {
		return user.Taken - out
	}

	log.Printf("WARNING, incorrect user amounts %d, %d, %d", user.Taken, user.Paid, user.Out)
	return 0
}

const (
	UserPrefix string = "user-"
)

type fnUserUpdate func(*User) bool

func (game *Game) updateUser(uid UID, handler fnUserUpdate) {

	var err error
	if err = game.db.Update(UserPrefix+string(uid), func(raw []byte) []byte {
		var user User
		var err error

		if err = json.Unmarshal(raw, &user); err != nil {
			panic(err)
		}

		if !handler(&user) {
			// not modified
			return nil
		}

		var updated []byte
		if updated, err = json.Marshal(&user); err != nil {
			panic(err)
		}

		return updated
	}); err == nil {
		return
	}

	if err != badger.ErrKeyNotFound {
		panic(err)
	}

	var user User
	if !handler(&user) {
		return
	}

	if err = game.db.Set(UserPrefix+string(uid), &user); err != nil {
		panic(err)
	}
}

func (game *Game) GetUser(uid UID) User {
	var err error
	var raw []byte

	err = game.db.Get(UserPrefix+string(uid), func(entry []byte) error {
		raw = make([]byte, len(entry))
		copy(raw, entry)
		return nil
	})

	if err != nil {
		if err != badger.ErrKeyNotFound {
			panic(err)
		}
		return User{}
	}

	var user User
	if err = json.Unmarshal(raw, &user); err != nil {
		panic(err)
	}

	return user
}
