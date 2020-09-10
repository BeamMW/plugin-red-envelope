package main

import (
	"encoding/json"
	"fmt"
	"github.com/olahol/melody"
	"log"
)

type statusParams struct {
	UserAddress string `json:"user_addr"`
}

type statusResult struct {
	Total        uint64 `json:"total"`
	Receiving    uint64 `json:"receiving"`
	Participants uint32 `json:"participants"`
	OpenTime     uint32 `json:"open_time"`
	Stake        uint64 `json:"stake"`
}

func onGetStatus(session* melody.Session, params *json.RawMessage) (res statusResult, err error) {
	var par statusParams
	if err = json.Unmarshal(*params, &par); err != nil {
		return
	}

	if config.Debug {
		log.Printf("status request for id: %v", par.UserAddress)
	}

	user, ok := users.GetStrict(par.UserAddress)
	if !ok {
		err = fmt.Errorf("user not found, try to login first")
		return
	}

	var status WalletStatus
	if status, err = wallet.GetStatus(); err != nil {
		return
	}

	res.Total     = status.Available
	res.Receiving = status.Receiving

	var txs []Transaction
	if txs, err = wallet.GetTransactions(); err != nil {
		return
	}

	// TODO: optimize
	var participants = make(map[string]bool)
	for _, tx := range txs {
		if tx.Status == Completed {
			participants[tx.Receiver] = true
			if tx.Receiver == user.DepositAddress {
				res.Stake += tx.Value
			}
		}
	}

	res.Participants = uint32(len(participants))
	res.OpenTime     = 0

	return
}
