package game

import (
	"encoding/json"
	"fmt"
	"github.com/BeamMW/red-envelope/database"
	"github.com/BeamMW/red-envelope/wallet"
	"github.com/dgraph-io/badger/v2"
	"log"
	"math/rand"
)

type TakesList map[UID]uint64

type Envelope struct {
	Address   string    `json:"address"`
	Remaining uint64    `json:"remaining"`
	Incoming  uint64    `json:"incoming"`
	Taken     uint64    `json:"taken"`      // amount taken for all rounds (lifetime)
	LastTakes TakesList `json:"last_takes"` // who has taken since the last deposit

	db  *database.Database
	api *wallet.API
}

const (
	EnvelopeKey = "red-envelope"
)

func (env *Envelope) loadOrCreate() error {
	// check if we have stored envelope
	err := env.db.Get(EnvelopeKey, func(raw []byte) error {
		return json.Unmarshal(raw, env)
	})

	if err == nil {
		log.Printf("Envelope loaded: %+v", *env)
		return nil
	}

	if err != badger.ErrKeyNotFound {
		// some unexpected error
		return err
	}

	//
	// There is no stored envelope
	// Must create a new one
	//
	addrs, err := env.api.OwnAddrList()
	if err != nil {
		return err
	}

	for _, addr := range addrs {
		if addr.Comment == EnvelopeKey {
			env.Address = addr.Address
		}
	}

	if len(env.Address) == 0 {
		env.Address, err = env.api.CreateAddress(EnvelopeKey, wallet.ExpNever)
		if err != nil {
			return err
		}
	}

	env.save()
	return nil
}

func (env *Envelope) save() {
	if err := env.db.Set(EnvelopeKey, env); err != nil {
		panic(err)
	}
}

func (env *Envelope) take(uid UID) (uint64, error) {
	if _, already := env.LastTakes[uid]; already {
		return 0, fmt.Errorf("user %s already took in this round", uid)
	}

	if env.Remaining == 0 {
		return 0, fmt.Errorf("there is nothing to take, balance is 0")
	}

	var minU64 = func (a, b uint64) uint64 {
		if a < b {
			return a
		}
		return b
	}

	const MinTake uint64 = 100000
	var takeAmount uint64

	if env.Remaining < MinTake {
		takeAmount = env.Remaining
	} else {
		var MaxTake = minU64(env.Remaining, 500000000)
		takeAmount = uint64(rand.Intn(int(MaxTake-MinTake))) + MinTake
	}

	env.Remaining -= takeAmount
	env.Taken += takeAmount

	// We pass takes list outside, so it is not safe to modify it. Do copy on change
	var newTakes = make(TakesList)
	newTakes[uid] = takeAmount

	for k, v := range env.LastTakes {
		newTakes[k] = v
	}

	env.LastTakes = newTakes
	env.save()

	return takeAmount, nil
}
