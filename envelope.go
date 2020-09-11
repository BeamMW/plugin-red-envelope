package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

type EnvelopeWin struct {
	DepositAddress string
	UserAddress    string
	Reward         uint64
	Height         uint64
	UnixTime       int64
}

const (
	WinPrefix = "win-"
)

type Envelope struct {
	users    chan uint32
	withdraw chan string
	lastWinHeight uint64 // this is accessed as atomic
	openTime      int64  // this is accessed as atomic, can be in past, this means already opened and not launched again

	//
	// Data below is accessed from multiple threads
	//
	wins []*EnvelopeWin
	mutex sync.Mutex
}

func (ev* Envelope) initEnvelope () {
	//
	// load Wins
	//
	ev.mutex.Lock()
		var lastWinH uint64 = 0
		err := DBforEach(WinPrefix, func (raw []byte) (err error) {
			var win EnvelopeWin
			if err = json.Unmarshal(raw, &win); err != nil {
				return
			}

			ev.wins = append(ev.wins, &win)
			if lastWinH < win.Height {
				lastWinH = win.Height
			}

			if config.Debug {
				log.Printf("Win loaded [%s : %v : %v]", win.UserAddress, win.Height, time.Unix(win.UnixTime, 0))
			}
			return
		})
	ev.mutex.Unlock()

	if err != nil {
		panic(err)
	}

	atomic.StoreUint64(&ev.lastWinHeight, lastWinH)
	log.Printf("Last win height: %v", lastWinH)

	//
	// Start staking goroutine
	//
	go func () {
		log.Println("--- Envelope goroutine started ---")
		defer log.Println("--- Envelope goroutine stopped ---")

		var users uint32 = 0

		// make unpack timer, since it is impossible to create
		// not running timer we stop and drain it immediately
		var envelopeUnpack = time.NewTimer(time.Duration(1))

		stopEnvelope := func () {
			if envelopeUnpack.Stop() {
				log.Println("--- Envelope Stopped ---")
			}
		}

		startEnvelope := func () {
			log.Println("--- Starting Envelope ---")
			var nextOpen = time.Now().Add(config.EnvelopeDuration).Unix()
			atomic.StoreInt64(&ev.openTime, nextOpen)
			envelopeUnpack.Reset(config.EnvelopeDuration)
		}

		stopEnvelope()

		for {
			select {
			//
			// Participants changes, start/stop envelope
			//
			case newUsers := <- ev.users:
				if newUsers > 0 && users == 0 {
					// somebody present, should start envelope if nobody was here before
					startEnvelope()
				}

				if newUsers == 0 {
					// no participants, stop
					stopEnvelope()
				}

				users = newUsers
			//
			// Envelope finished running
			//
			case <- envelopeUnpack.C:
				go func() {
					if err := ev.unpackEnvelope(); err != nil {
						log.Printf("Failed to unpack envelope: %v", err)
					}
				}()
			}
		}
	} ()

	//
	// Start withdraw goroutine
	//
	go func () {
		for {
			userid := <- ev.withdraw
			log.Printf("--- Withdraw request for %v ---", userid)

			user, err := users.Get(userid)
			if err != nil {
				panic(err)
			}

			stats, err := ev.getUserStats(user)
			if err != nil {
				panic(err)
			}

			if stats.AvailableReward == 0 {
				log.Printf("Cannot withdraw for %s, funds not available", userid)
				user.FinishWithdraw()
				continue
			}

			var fee uint64 = 100
			var amount uint64 = stats.AvailableReward - fee

			txid, err := wallet.SendBEAM(user.UserAddress, user.DepositAddress, amount, fee)
			if err != nil {
				log.Printf("Failed to withdraw for %s, %v", userid, err)
				user.FinishWithdraw()
				continue
			}

			log.Printf("Sent to %s %v BEAM, fee %v GOTH, txid %s", userid, amount / 100000000, fee, txid)
			user.FinishWithdraw()
		}
	} ()
}

func (ev *Envelope) updateEnvelope (users uint32) {
	ev.users <- users
}

func (ev *Envelope) getCurrentTxList() (txs []Transaction, err error) {
	var lwh = atomic.LoadUint64(&ev.lastWinHeight)

	var alltxs []Transaction
	if alltxs, err = wallet.GetTransactions(); err != nil {
		return
	}

	for _, tx := range alltxs {
		if tx.Height > lwh {
			txs = append(txs, tx)
		}
	}

	return
}

func (ev *Envelope) unpackEnvelope() (err error) {
	log.Println("--- Unpacking Envelope ---")

	var win = &EnvelopeWin{
	}
	win.UnixTime = time.Now().Unix()

	var txs []Transaction
	if txs, err = ev.getCurrentTxList(); err != nil {
		return
	}

	var participants = make(map[string]bool)
	for _, tx := range txs {
		if tx.Status == Completed && tx.Income {
			win.Reward += tx.Value

			if win.Height < tx.Height {
				win.Height = tx.Height
			}

			var depoAddr = tx.Receiver
			if _, ok := users.GetByDepoAddr(depoAddr); !ok {
				log.Printf("WARNING: participating user not found %s, orphaned transaction", tx.Receiver)
				continue
			} else {
				participants[tx.Receiver] = true
			}
		}
	}

	//
	// Find out winner
	//
	var pcnt = len(participants)
	if pcnt == 0 {
		// This can happen in case if database was deleted but wallet not reset
		log.Println("--- Envelope Unpack halted, to next round ---")
		return fmt.Errorf("no valid participants found")
	}

	rand.Seed(time.Now().UnixNano())
	var winnerIdx = rand.Intn(pcnt)

	var idx = 0
	for depoAddr, _ := range participants {
		if idx == winnerIdx {
			win.DepositAddress = depoAddr
			break
		}
		idx++
	}

	var ok bool
	var user *User
	if user, ok = users.GetByDepoAddr(win.DepositAddress); !ok {
		return fmt.Errorf("failed to get winner user address")
	}
	win.UserAddress = user.UserAddress

	log.Println("\tReward:", win.Reward)
	log.Println("\tParticipants:", pcnt)
	log.Println("\tWinner index is:", winnerIdx)
	log.Println("\tWinner deposit addr is:", win.DepositAddress)
	log.Println("\tWinner user addr is:", win.UserAddress)
	log.Println("\tWin height:", win.Height)
	log.Println("\tWin time:", time.Unix(win.UnixTime, 0))

	var prefx = WinPrefix + strconv.FormatUint(win.Height, 10)
	if err = DBStore(prefx, win); err != nil {
		return
	}

	ev.mutex.Lock()
	ev.wins = append(ev.wins, win)
	ev.mutex.Unlock()

	atomic.StoreUint64(&ev.lastWinHeight, win.Height)
	log.Println("--- Envelope Unpacked ---")

	return
}

type EnvelopeUserStats struct {
	TotalInEnvelope   uint64
	ReceivedFromUser  uint64
	ReceivingFromAll  uint64
	ReceivingFromUser uint64
	Participants      uint32
	OutgoingReward    uint64
	PaidReward        uint64
	AvailableReward   uint64
	LastWinTime       int64
	OpenTime          int64
}

func (ev *Envelope) getUserStats (user *User) (stats EnvelopeUserStats, err error) {
	if config.Debug {
		var status WalletStatus
		if status, err = wallet.GetStatus(); err != nil {
			return
		}

		log.Println("Wallet Status, available:", status.Available, " incoming:", status.Receiving)
	}

	var lwh = atomic.LoadUint64(&ev.lastWinHeight)
	var participants = make(map[string]bool)

	var txs []Transaction
	if txs, err = wallet.GetTransactions(); err != nil {
		return
	}

	var c = 0
	for _, tx := range txs {
		if tx.Height > lwh {
			c++
		}
		//
		// -- Completed incoming transaction after the last win --
		// Count its to total available balance in envelope
		// and may be as user's stake. Also count all participants
		//
		if tx.Height > lwh && tx.Income && tx.Status == Completed  {
			participants[tx.Receiver] = true
			stats.TotalInEnvelope += tx.Value
			if tx.Receiver == user.DepositAddress {
				stats.ReceivedFromUser += tx.Value
			}
		}

		//
		// -- Pending incoming transaction after the last win --
		// Count it to total incoming balance in envelope
		// and may be as incoming from user
		//
		if tx.Income  && (tx.Status == InProgress || tx.Status == Registering) {
			stats.ReceivingFromAll += tx.Value
			if tx.Receiver == user.DepositAddress {
				stats.ReceivingFromUser += tx.Value
			}
		}

		//
		// Transaction is going to user
		// Count as outgoing reward
		//
		if tx.Receiver == user.UserAddress && (tx.Status == InProgress || tx.Status == Registering) {
			stats.OutgoingReward += tx.Value + tx.Fee
		}

		//
		// Transaction is sent to user
		// Count as paid reward
		//
		if tx.Status == Completed && tx.Receiver == user.UserAddress {
			stats.PaidReward += tx.Value + tx.Fee
		}
	}

	stats.Participants = uint32(len(participants))
	ev.updateEnvelope(stats.Participants)

	//
	// now we need to check wins
	//
	ev.mutex.Lock()
		var totalReward uint64
	    var lastWinTime int64
		for _, win := range ev.wins {
			if win.UserAddress == user.UserAddress {
				totalReward += win.Reward
				if lastWinTime < win.UnixTime {
					lastWinTime = win.UnixTime
				}
			}
		}
	ev.mutex.Unlock()

	stats.AvailableReward = totalReward - stats.PaidReward - stats.OutgoingReward
	stats.LastWinTime = lastWinTime

	stats.OpenTime = atomic.LoadInt64(&ev.openTime)
	if stats.OpenTime < time.Now().Unix() {
		stats.OpenTime = 0
	}

	return
}

func (ev *Envelope) Withdraw(user *User) {
	ev.withdraw <- user.UserAddress
}

var (
	// TODO: move to init
	envelope = &Envelope{
		users: make(chan uint32),
		withdraw: make(chan string),
	}
)
