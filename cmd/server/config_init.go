package main

import "api/internal/config"

func InitConfig() {
	_ = config.Load()
}
