package main

import (
	"encoding/json"
	"fmt"
	"github.com/chapati/melody"
)

func onClientConnect(session *melody.Session) error {
	return nil
}

type loginParams struct {
	UserAddress string `json:"user_addr"`
}

type loginResult struct {
}

func onClientLogin(session *melody.Session, params *json.RawMessage) (res loginResult, err error) {
	var req loginParams
	if err = json.Unmarshal(*params, &req); err != nil {
		return
	}

	if len(req.UserAddress) == 0 {
		err = fmt.Errorf("provide valid withdrawal address")
		return
	}

	setUserID(session, req.UserAddress)
	sendStatus(session)

	return
}
