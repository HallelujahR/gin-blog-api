package main

import (
	"api/internal/config"
	"log"
	"net/http"
)

func main() {
	cfg := config.Load()
	if cfg.PprofEnabled {
		go func() {
			addr := ":" + cfg.PprofPort
			log.Printf("pprof server start at %s", addr)
			log.Fatal(http.ListenAndServe(addr, nil))
		}()
	}
	Run()
}
