package main

import (
	"encoding/json"
	"fmt"
	"github.com/olahol/melody"
	"log"
)

type loginParams struct {
	UserAddress string  `json:"user_addr"`
}

type loginResult struct {
	DepositAddress string `json:"envelope_addr"`
}

func onClientLogin(session* melody.Session, params *json.RawMessage) (res loginResult, err error) {
	var req loginParams
	if err = json.Unmarshal(*params, &req); err != nil {
		return
	}

	if len(req.UserAddress) == 0 {
		err = fmt.Errorf("provide valid withdrawal address")
		return
	}

	if config.Debug {
		log.Println("Login request for", req.UserAddress)
	}

	var user *User
	if user, err = users.GetOrAdd(req.UserAddress); err != nil {
		return
	}

	setUserID(session, user.UserAddress)
	res.DepositAddress = user.DepositAddress

	return
}
