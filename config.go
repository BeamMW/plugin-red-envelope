package main

import (
	"encoding/json"
	"errors"
	"github.com/chapati/melody"
	"log"
	"os"
	"path/filepath"
	"time"
)

const (
	ConfigFile = "config.json"
)

type StaticEndpoint struct {
	Folder string
	Endpoint string
	SameOrigin bool
}


type Config struct {
	LogName           string
	Debug             bool
	ListenAddress     string
	StaticFiles       string
	StaticEndpoints   [] StaticEndpoint
	SameOriginFolders [] string
	WalletAPIAddress  string
	DatabasePath      string
	EnvelopeDuration  time.Duration
}

var config = Config{
	Debug:            true,
	ListenAddress:    "127.0.0.1:13666",
	StaticEndpoints:  [] StaticEndpoint {
		StaticEndpoint {
			Folder:   "./html/",
			Endpoint: "/app/",
			SameOrigin: false,
		},
	},
	WalletAPIAddress: "http://127.0.0.1:14666/api/wallet",
	LogName:          "BEAM Red Envelope",
	DatabasePath:     "./db-files",
	EnvelopeDuration: time.Second * 10,
}

func loadConfig(m *melody.Melody) {
	config.Read(ConfigFile, m)
}

func (cfg *Config) Read(fname string, m *melody.Melody) {
	// Set melody params
	m.Config.MaxMessageSize = 1024

	// Read config file
	fpath, err := filepath.Abs(fname)
	if err != nil {
		panic(err)
	}

	log.Println("reading configuration from", fpath)
	file, err := os.Open(fname)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.DisallowUnknownFields()

	err = decoder.Decode(cfg)
	if err != nil {
		panic(err)
	}

	if len(cfg.ListenAddress) == 0 {
		panic(errors.New("config, missing ListenAddress"))
	}

	var mode = "RELEASE"
	if cfg.Debug {
		mode = "DEBUG"
	}

	log.Printf("starting in %v mode", mode)
}
