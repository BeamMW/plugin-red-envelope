package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
)

type User struct {
	UserAddress         string
	DepositAddress      string
	ProcessingWidthdraw int32  `json:"-"`
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

func (user* User) StartWithdraw () error {
	var processing = atomic.LoadInt32(&user.ProcessingWidthdraw)
	if processing != 0 {
		return fmt.Errorf("already withdrawing")
	}

	atomic.StoreInt32(&user.ProcessingWidthdraw, 1)
	envelope.Withdraw(user)

	return nil
}

func (user* User) CanWithdraw () bool {
	return atomic.LoadInt32(&user.ProcessingWidthdraw) == 0
}

func (user* User) FinishWithdraw() {
	atomic.StoreInt32(&user.ProcessingWidthdraw, 0)
}

type Users struct {
	all   map[string] *User
	mutex sync.Mutex
}

func (users *Users) Get(address string) (*User, error) {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	if user, ok := users.all[address]; ok {
		return user, nil
	}

	return nil, fmt.Errorf("user %s not found", address)
}

func (users *Users) GetByDepoAddr(depoaddr string) (*User, bool) {
	users.mutex.Lock()
	defer users.mutex.Unlock()

	for _, user := range users.all {
		if user.DepositAddress == depoaddr {
			return user, true
		}
	}

	return nil, false
}

func (users *Users) GetOrAdd(address string) (user *User, err error) {
	if user, err = users.Get(address); err == nil {
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
