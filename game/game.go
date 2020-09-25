package game

import (
	"github.com/BeamMW/red-envelope/database"
	"github.com/BeamMW/red-envelope/wallet"
	"sync"
)

type Game struct {
	NewStatus chan *Status
	envelope  *Envelope
	withdraw  chan UID
	db        *database.Database
	api       *wallet.API
	mutex     *sync.RWMutex
}

func New(db *database.Database, apiaddr string) (*Game, error) {
	var api = wallet.New(apiaddr)
	var envelope = &Envelope{
		db:  db,
		api: api,
	}

	if err := envelope.loadOrCreate(); err != nil {
		return nil, err
	}

	var game = &Game{
		db:        db,
		api:       api,
		envelope:  envelope,
		mutex:     &sync.RWMutex{},
		withdraw:  make(chan UID),
		NewStatus: make(chan *Status, 10),
	}

	go game.collectStatus()
	go game.handleWithdraw()

	return game, nil
}

func (game *Game) rlock() {
	game.mutex.RLock()
}

func (game *Game) runlock() {
	game.mutex.RUnlock()
}

func (game *Game) lock() {
	game.mutex.Lock()
}

func (game *Game) unlock() {
	game.mutex.Unlock()
}
