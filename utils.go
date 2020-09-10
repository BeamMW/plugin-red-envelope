package main

import (
	"log"
	"os"
)

func getCWD() string {
	res, err := os.Getwd()
	if err != nil {
		log.Printf("Error while getting working directory, %v", err)
		return ""
	}
	return res
}
