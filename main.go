package main

import (
	"api/configs"
	"api/initializer"
	"log"
	"net/http"
)

func main() {
	cfg := configs.Load()
	if cfg.PprofEnabled {
		go func() {
			addr := ":" + cfg.PprofPort
			log.Printf("pprof server start at %s", addr)
			log.Fatal(http.ListenAndServe(addr, nil))
		}()
	}
	initializer.Run()
}
