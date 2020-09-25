package main

import (
	"fmt"
	"net/http"
)

func helloRequest(w http.ResponseWriter, r *http.Request) {
	_, _ = fmt.Fprintf(w, "Hello! This is the %s", config.LogName)
}
