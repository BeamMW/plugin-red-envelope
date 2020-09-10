package main

import (
	"encoding/json"
	"github.com/olahol/melody"
)

func onClientDisconnect(session *melody.Session) error {
	// Client might be already logged out since disconnect is always called
	return nil
}

type logoutParams struct {
}

type logoutResult struct {
}

func onClientLogout(session* melody.Session, params *json.RawMessage) (res logoutResult, err error) {
	var req logoutParams
	if err = json.Unmarshal(*params, &req); err != nil {
		return
	}
	// TODO: may be purge all inactive non-participating users
	return
}
