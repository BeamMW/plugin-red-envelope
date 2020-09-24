package main

import (
	"encoding/json"
	"github.com/BeamMW/red-envelope/game"
	"github.com/chapati/melody"
	"log"
)

type takeParams struct {
}

type takeResult struct {
	Amount uint64 `json:"amount"`
}

func onClientTake(session* melody.Session, params *json.RawMessage) (res takeResult, err error) {
	var req widthdrawParams
	if err = json.Unmarshal(*params, &req); err != nil {
		return
	}

	var uid string
	if uid, err = getUserID(session); err != nil {
		log.Println(err)
	}

	if config.Debug {
		log.Println("take request for", uid)
	}

	res.Amount, err = Game.Take(game.UID(uid))
	return
}