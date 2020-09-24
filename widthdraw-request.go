package main

import (
	"encoding/json"
	"github.com/chapati/melody"
)

type widthdrawParams struct {
	UserAddress string  `json:"user_addr"`
}

type withdrawResult struct {
}

func onClientWithdraw(session* melody.Session, params *json.RawMessage) (res withdrawResult, err error) {
	/*var req widthdrawParams
	if err = json.Unmarshal(*params, &req); err != nil {
		return
	}

	if len(req.UserAddress) == 0 {
		err = fmt.Errorf("provide valid user address")
		return
	}

	if config.Debug {
		log.Println("withdraw request for", req.UserAddress)
	}

	var user *User
	if user, err = users.Get(req.UserAddress); err != nil {
		return
	}

	err = user.StartWithdraw()*/
	return
}
