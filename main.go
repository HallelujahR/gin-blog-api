package main

import (
	"api/initializer"
	"log"
	"net/http"
)

func main() {
	go func() {
		log.Printf("pprof server start at :6060")
		log.Fatal(http.ListenAndServe(":6060", nil))
	}()
	initializer.Run()
}
