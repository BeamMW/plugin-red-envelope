package main

import (
	"github.com/BeamMW/plugin-red-envelope/game"
	"github.com/BeamMW/plugin-red-envelope/jsonrpc"
	"github.com/chapati/melody"
	"log"
)

type statusResult struct {
	EnvelopeAddress   string `json:"envelope_addr"`
	EnvelopeRemaining uint64 `json:"envelope_remaining"`
	EnvelopeIncoming  uint64 `json:"envelope_incoming"`
	TakenAmount       uint64 `json:"taken_amount"`
	RewardAvailable   uint64 `json:"available_reward"`
	RewardPaid        uint64 `json:"paid_reward"`
	RwardOutgoing     uint64 `json:"outgoing_reward"`
}

const (
	GameStatusMethod = "game-status"
)

var (
	statusReq = make(chan *melody.Session)
)

func formStatusMsg(status *game.Status, session *melody.Session) ([]byte, error) {
	var suid string
	var err error

	if suid, err = getUserID(session); err != nil {
		log.Println(err)
		return nil, err
	}

	var res = statusResult{
		EnvelopeAddress:   status.Address,
		EnvelopeRemaining: status.Remaining,
		EnvelopeIncoming:  status.Incoming,
	}

	var uid = game.UID(suid)
	var user = Game.GetUser(uid)

	res.RewardPaid = user.Paid
	res.RwardOutgoing = user.Out
	res.TakenAmount = status.LastTakes[uid]
	res.RewardAvailable = user.AvailableReward()

	var bytes []byte
	if bytes, err = jsonrpc.WrapNotification(GameStatusMethod, &res); err != nil {
		return nil, err
	}

	return bytes, nil
}

func sendStatus(session *melody.Session) {
	// this might block, so just launch in goroutine
	go func() {
		statusReq <- session
	}()
}

func broadcastStatus(m *melody.Melody) {
	go func() {
		var currStatus *game.Status

		for {
			select {
			case session := <- statusReq:
				var bytes []byte
				var err error

				if bytes, err = formStatusMsg(currStatus, session); err != nil {
					log.Printf("WARNING: failed to form status message, %v", err)
					break
				}

				if err = session.Write(bytes); err != nil {
					log.Printf("WARNING: failed to write status message, %v", err)
					break
				}

			case currStatus = <- Game.NewStatus:
				var bytes []byte
				var err error

				if err = m.BroadcastEx(func(session *melody.Session) []byte {
					if bytes, err = formStatusMsg(currStatus, session); err != nil {
						log.Printf("WARNING: failed to form status message, %v", err)
						return nil
					}
					return bytes
				}); err != nil {
					log.Printf("WARNING: failed to broadcast status message, %v", err)
					break
				}
			} // select
		} // for
	}()
}
