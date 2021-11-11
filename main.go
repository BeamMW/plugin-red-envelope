package main

import (
	"flag"
	"github.com/BeamMW/plugin-red-envelope/database"
	"github.com/BeamMW/plugin-red-envelope/game"
	"github.com/chapati/melody"
	"log"
	"net/http"
)

var (
	DB   *database.Database
	Game *game.Game
)

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) > 1 {
		log.Fatal("too many arguments")
	}

	if len(args) == 1 {
		if args[0] == "version" {
			printVersion()
			return
		} else {
			log.Fatalf("unknown command: %v", args[0])
		}
	}

	//
	// Command line is OK
	//
	log.Printf("starting %s, version %s", config.LogName, programVersion)
	log.Printf("cwd is %s", getCWD())

	//
	// Init stuff
	//
	var err error

	m := melody.New()
	loadConfig(m)

	DB, err = database.New(config.DatabasePath)
	if err != nil {
		panic(err)
	}

	Game, err = game.New(DB, config.WalletAPIAddress)
	if err != nil {
		panic(err)
	}

	// Automatically update status of all active users
	go broadcastStatus(m)

	//
	// HTTP API
	//
	// Now just hello
	http.HandleFunc("/", helloRequest)

	// File server
	for _, epoint := range config.StaticEndpoints {
		log.Printf("Serving static files from %s at %s, same origin is %v", epoint.Folder, epoint.Endpoint, epoint.SameOrigin)
		fs := http.FileServer(http.Dir(epoint.Folder))
		strip := http.StripPrefix(epoint.Endpoint, fs)
		same := epoint.SameOrigin

		http.HandleFunc(epoint.Endpoint, func (w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Cache-Control", "must-revalidate")

			if same {
				w.Header().Add("Cross-Origin-Opener-Policy", "same-origin")
				w.Header().Add("Cross-Origin-Embedder-Policy", "require-corp")
			}

			strip.ServeHTTP(w, r)
		})
	}

	//
	// JsonRPCv2.0 over WebSockets
	//
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if err := m.HandleRequest(w, r); err != nil {
			log.Printf("WARNING: websocket handle request error, %v", err)
		}
	})

	m.HandleMessage(func(session *melody.Session, msg []byte) {
		go func() {
			if resp := onClientMessage(session, msg); resp != nil {
				if err := session.Write(resp); err != nil {
					log.Printf("WARNING: websocket jsonRpcProcessWallet error, %v", err)
				}
			}
		}()
	})

	m.HandleConnect(func(session *melody.Session) {
		if config.Debug {
			log.Printf("websocket server new session %v", session)
		}
		if err := onClientConnect(session); err != nil {
			log.Printf("WARNING: websocket onClientConnect error, %v", err)
		}
	})

	m.HandleDisconnect(func(session *melody.Session) {
		if err := onClientDisconnect(session); err != nil {
			log.Printf("WARNING: websocket onClientDisconnect error, %v", err)
		}
	})

	m.HandlePong(func(session *melody.Session) {
		if err := onClientPong(session); err != nil {
			log.Printf("WARNING: websocket onClientPong error, %v", err)
		}
	})

	m.HandleError(func(session *melody.Session, err error) {
		if config.Debug {
			log.Printf("WARNING: websocket error, %v", err)
		}
	})

	log.Println(config.ListenAddress, "Go!")
	if err := http.ListenAndServe(config.ListenAddress, nil); err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}
