package main

import (
	"encoding/json"
	"log"
	"sync"
)

type User struct {
	UserAddress    string
	DepositAddress string
}

const (
	UserPrefix = "user-"
)

func (user *User) StoreChanges () {
	var key = UserPrefix + user.UserAddress
	if err := DBStore(key, user); err != nil {
		panic(err)
	}
}

type Users struct {
	all   map[string] *User
	mutex sync.Mutex
}

func (users *Users) GetStrict(address string) (*User, bool){
	users.mutex.Lock()
	defer users.mutex.Unlock()

	user, ok := users.all[address]
	return user, ok
}

func (users *Users) GetOrAdd(address string) (user *User, err error) {
	var ok bool
	if user, ok = users.GetStrict(address); ok {
		return
	}

	var depositAddress string
	if depositAddress, err = wallet.CreateAddress(); err != nil {
		return
	}

	user = &User {
		UserAddress: address,
		DepositAddress: depositAddress,
	}

	if config.Debug {
		log.Printf("user created [%s : %v]", user.UserAddress, user.DepositAddress)
	}

	user.StoreChanges()

	users.mutex.Lock()
	defer users.mutex.Unlock()
	users.all[address] = user

	return
}

func (users *Users) Load() {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	var counter int
	err := DBforEach(UserPrefix, func (raw []byte) (err error) {
		var user User
		if err = json.Unmarshal(raw, &user); err != nil {
			return
		}
		users.all[user.UserAddress] = &user
		if config.Debug {
			log.Printf("user loaded [%s : %v]", user.UserAddress, user.DepositAddress)
		}
		counter++
		return
	})

	if err != nil {
		panic(err)
	}

	log.Printf("Users loaded: %v", counter)
}

var (
	users = &Users{
		all:   make(map[string] *User),
		mutex: sync.Mutex{},
	}
)
