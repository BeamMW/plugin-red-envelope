package game

import (
	"github.com/BeamMW/plugin-red-envelope/wallet"
	"log"
)

func (game *Game) Withdraw (uid UID) {
	go func () {
		game.withdraw <- uid
	} ()
}

func (game *Game) handleWithdraw () {
	for {
		uid := <-game.withdraw
		log.Printf("--- Withdraw request for %v ---", uid)

		// TODO: revise locks
		// TODO: return withdraw status
		game.lock()
			game.updateUser(uid, func(user *User) bool {
				var amount = user.AvailableReward()

				if amount == 0 {
					log.Printf("\tNothing to withdraw")
					return false
				}

				if amount < wallet.DefaultFee + 1 {
					log.Printf("\tAvailable amount (%v) is not enough to pay the fee", amount)
					return false
				}

				var err error
				if _, err = game.api.SendBEAM(string(uid), game.envelope.Address, amount - wallet.DefaultFee, wallet.DefaultFee); err != nil {
					log.Printf("\tfailed to send tx %v", err)
					return false
				}

				user.Out += amount // N.B. this will be overwritten during the next status update
				log.Printf("\tSent %v BEAM", wallet.GROTH2Beam(amount))

				return true
			})
		game.unlock()
	}
}
