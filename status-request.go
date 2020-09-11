package main

import (
	"encoding/json"
	"github.com/olahol/melody"
)

type statusParams struct {
	UserAddress string `json:"user_addr"`
}

type statusResult struct {
	TotalInEnvelope   uint64 `json:"total_in_envelope"`
	ReceivedFromUser  uint64 `json:"received_from_user"`
	ReceivingFromAll  uint64 `json:"receiving_from_all"`
	ReceivingFromUser uint64 `json:"receiving_from_user"`
	Participants      uint32 `json:"participants"`
	OutgoingReward    uint64 `json:"outgoing_reward"`
	PaidReward        uint64 `json:"paid_reward"`
	AvailableReward   uint64 `json:"available_reward"`
	LastWinTime       int64  `json:"last_win_time"`
	OpenTime          int64  `json:"envelope_open_time"`
	CanWithdraw       bool   `json:"can_withdraw"` // TODO:refactor send automatic status updates from server and this can be removed
}

func onGetStatus(session* melody.Session, params *json.RawMessage) (res statusResult, err error) {
	var par statusParams
	if err = json.Unmarshal(*params, &par); err != nil {
		return
	}

	var user *User
	if user, err = users.Get(par.UserAddress); err != nil {
		return
	}

	var stats EnvelopeUserStats
	if stats, err = envelope.getUserStats(user); err != nil {
		return
	}

	res.TotalInEnvelope   = stats.TotalInEnvelope
	res.ReceivedFromUser  = stats.ReceivedFromUser
	res.ReceivingFromAll  = stats.ReceivingFromAll
	res.ReceivingFromUser = stats.ReceivingFromUser
	res.Participants      = stats.Participants
	res.OutgoingReward    = stats.OutgoingReward
	res.PaidReward        = stats.PaidReward
	res.AvailableReward   = stats.AvailableReward
	res.LastWinTime       = stats.LastWinTime
	res.OpenTime          = stats.OpenTime
	res.CanWithdraw       = user.CanWithdraw() && stats.AvailableReward > 0

	return
}
