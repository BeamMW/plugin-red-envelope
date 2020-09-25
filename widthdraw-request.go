package main

import (
	"encoding/json"
	"github.com/BeamMW/red-envelope/game"
	"github.com/chapati/melody"
)

type widthdrawParams struct {
}

type withdrawResult struct {
}

func onClientWithdraw(session *melody.Session, params *json.RawMessage) (res withdrawResult, err error) {
	var uid string
	if uid, err = getUserID(session); err != nil {
		return
	}

	Game.Withdraw(game.UID(uid))
	return
}
