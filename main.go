package main

import (
	"flag"
	"github.com/olahol/melody"
	"log"
	"net/http"
)

func main () {
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

	m := melody.New()
	loadConfig(m)

	//
	// Functions below can panic
	//
	initDatabase()
	users.Load()
	initEnvelope()

	//
	// HTTP API
	//

	// Now just hello
	http.HandleFunc("/", helloRequest)

	//
	// JsonRPCv2.0 over WebSockets
	//
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		if err := m.HandleRequest(w, r); err != nil {
			log.Printf("websocket handle request error, %v", err)
		}
	})

	m.HandleMessage(func(session *melody.Session, msg []byte) {
		go func() {
			if resp := onClientMessage(session, msg); resp != nil {
				if err := session.Write(resp); err != nil {
					log.Printf("websocket jsonRpcProcessWallet error, %v", err)
				}
			}
		} ()
	})

	m.HandleConnect(func(session *melody.Session) {
		if config.Debug {
			log.Printf("websocket server new session")
		}
	})

	m.HandleDisconnect(func(session *melody.Session) {
		if err := onClientDisconnect(session); err != nil {
			log.Printf("websocket onClientDisconnect error, %v", err)
		}
	})

	m.HandlePong(func(session *melody.Session) {
		if err := onClientPong(session); err != nil {
			log.Printf("websocket onClientPong error, %v", err)
		}
	})

	m.HandleError(func(session *melody.Session, err error) {
		if config.Debug {
			log.Printf("websocket error, %v", err)
		}
	})

	log.Println(config.ListenAddress, "Go!")
	if err := http.ListenAndServe(config.ListenAddress, nil); err != nil {
		log.Fatalf("Failed to start server %v", err)
	}
}
