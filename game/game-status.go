package game

import (
	"github.com/BeamMW/red-envelope/wallet"
	"log"
	"time"
)

type Status struct {
	Address   string
	Remaining uint64
	Incoming  uint64
	LastTakes TakesList // RETURNED BY POINTER FOR OPTIMIZATION, DO NOT MODIFY MAP ITSELF, ONLY READ
}

func (game *Game) getStatusNoLock() *Status {
	var status = &Status{
		Address:   game.envelope.Address,
		Remaining: game.envelope.Remaining,
		Incoming:  game.envelope.Incoming,
		LastTakes: game.envelope.LastTakes,
	}
	return status
}

func (game *Game) collectStatus() {
	for {
		func() {
			log.Println("--- collecting game status ---")
			defer log.Println("--- game status collected ---")

			// just print general wallet status
			var wstatus wallet.Status
			var err error

			if wstatus, err = game.api.Status(); err != nil {
				log.Printf("\tWARING: failed to get wallet status, %v", err)
			} else {
				log.Printf("\tTotal Available: %v", wallet.GROTH2Beam(wstatus.Available))
				log.Printf("\tTotal Receiving: %v", wallet.GROTH2Beam(wstatus.Receiving))
			}

			//
			// Iterate over all transactions and collect info
			//
			var txs []wallet.Transaction
			if txs, err = game.api.GetTransactions(); err != nil {
				log.Printf("\tERROR: failed to get txlist, %v", err)
				return
			}
			log.Printf("\tTxs: %v", len(txs))

			var envelopeIncome uint64
			var envelopeIncoming uint64
			var newAmounts = make(map[string]User)

			var addPaid = func(uid string, amount uint64) {
				amounts := newAmounts[uid]
				amounts.Paid = amount
				newAmounts[uid] = amounts
			}

			var addPaying = func(uid string, amount uint64) {
				amounts := newAmounts[uid]
				amounts.Out = amount
				newAmounts[uid] = amounts
			}

			var isInProgress = func(tx *wallet.Transaction) bool {
				return tx.Status == wallet.TxInProgress || tx.Status == wallet.TxPending || tx.Status == wallet.TxRegistering
			}

			var delTx = func (txid string) bool {
				if err = game.api.DeleteTransaction(txid); err != nil {
					log.Printf("WARNING: failed to delete transaction %s", txid)
					return false
				}
				return true
			}


			for _, tx := range txs {
				var counted = false

				//
				// -- Completed incoming transaction to the envelope address --
				// Count its to the total available balance in envelope
				//
				if tx.Status == wallet.TxCompleted && tx.Receiver == game.envelope.Address {
					counted = true
					if delTx(tx.TxId) {
						envelopeIncome += tx.Value
					}
				}

				//
				// -- In progress incoming transaction to the envelope address --
				// Count it to total incoming balance
				//
				if isInProgress(&tx) && tx.Receiver == game.envelope.Address {
					counted = true
					envelopeIncoming += tx.Value
				}

				//
				// -- Completed outgoing transaction ---
				// Count as reward paid to the user
				//
				if tx.Status == wallet.TxCompleted && !tx.Income {
					counted = true
					if delTx(tx.TxId) {
						addPaid(tx.Receiver, tx.Value+tx.Fee)
					}
				}

				//
				// -- In progress outgoing transaction ---
				// Count as reward being paid to the user
				//
				if isInProgress(&tx) && !tx.Income {
					addPaying(tx.Receiver, tx.Value+tx.Fee)
					counted = true
				}

				//
				// Safely ignore failed & cancelled transactions
				//
				if tx.Status == wallet.TxFailed || tx.Status == wallet.TxCancelled {
					delTx(tx.TxId)
					counted = true
				}

				if !counted {
					log.Printf("WARNING: transaction not counted, %+v", tx)
					// TODO: uncomment after testing
					// delTx(tx.TxId)
				}
			}

			var newStatus *Status
			game.lock()
				// if somebody deposited, allow to take for everyone even if participated before
				if envelopeIncome > 0 {
					game.envelope.LastTakes = make(TakesList)
				}

				game.envelope.Remaining += envelopeIncome
				game.envelope.Incoming = envelopeIncoming
				game.envelope.save()

				for uid, amounts := range newAmounts {
					game.updateUser(UID(uid), func(user *User) bool {
						user.Paid += amounts.Paid
						user.Out   = amounts.Out
						return true
					})
				}

				newStatus = game.getStatusNoLock()
			game.unlock()

			log.Printf("\tRemaining: %v", wallet.GROTH2Beam(newStatus.Remaining))
			log.Printf("\tIncoming: %v", wallet.GROTH2Beam(newStatus.Incoming))
			log.Printf("\tTakes: %v", len(newStatus.LastTakes))

			game.NewStatus <- newStatus
		}()

		time.Sleep(time.Second * 5) // TODO: update for release
	}
}
