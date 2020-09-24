package main

import (
	"github.com/BeamMW/red-envelope/game"
	"github.com/BeamMW/red-envelope/jsonrpc"
	"github.com/chapati/melody"
	"log"
)

type statusResult struct {
	EnvelopeAddress    string `json:"envelope_addr"`
	EnvelopeRemaining  uint64 `json:"envelope_remaining"`
	EnvelopeIncoming   uint64 `json:"envelope_incoming"`
	TakenAmount        uint64 `json:"taken_amount"`
	RewardAvailable    uint64 `json:"available_reward"`
	RewardPaid         uint64 `json:"paid_reward"`
	RwardOutgoing      uint64 `json:"outgoing_reward"`
}

const (
	GameStatusId = "game-status"
)

func formStatusMsg(status *game.Status, session* melody.Session) ([]byte, error) {
    var suid string
    var err error

	if suid, err = getUserID(session); err != nil {
		log.Println(err)
		return nil, err
	}

	var res = statusResult {
		EnvelopeAddress:   status.Address,
		EnvelopeRemaining: status.Remaining,
		EnvelopeIncoming:  status.Incoming,
	}

	var uid = game.UID(suid)
	var user = Game.GetUser(uid)

	res.RewardPaid       = user.Paid
	res.RwardOutgoing    = user.Out
	res.TakenAmount      = status.LastTakes[uid]

	if user.Paid + user.Out > user.Taken {
		log.Printf("Warning, user %s incorect amounts %d, %d, %d", suid, user.Taken, user.Paid, user.Out)
		res.RewardAvailable = 0
	} else {
		res.RewardAvailable = user.Taken - user.Paid - user.Out
	}

	var bytes []byte
	if bytes, err = jsonrpc.WrapMessage(GameStatusId, &res); err != nil {
		return nil, err
	}

	return bytes, nil
}

func sendStatus (session* melody.Session) error {
	var status = Game.GetStatus()

	var bytes []byte
	var err error

	if bytes, err = formStatusMsg(status, session); err != nil {
		return err
	}
	return session.Write(bytes)
}

func broadcastStatus(m *melody.Melody) {
	for {
		status := <- Game.NewStatus

		var bytes []byte
		var err error

		if err = m.BroadcastEx(func (session *melody.Session) []byte {
			if bytes, err = formStatusMsg(status, session); err != nil {
				log.Printf("Error: failed to form status message, %v", err)
				return nil
			}
			return bytes
		}); err != nil {
			log.Printf("Error: failed to broadcast status message, %v", err)
		}
	}
}
