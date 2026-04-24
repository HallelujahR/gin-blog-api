package main

import (
	"api/internal/config"
	"api/internal/modules/analytics"
	"fmt"
)

func Run() {
	InitConfig()
	InitDB()
	cfg := config.Load()
	if cfg.AnalyticsEnabled {
		analytics.StartETLWorker(cfg.AnalyticsETL)
	}
	r := InitRouter()
	_ = r.Run(fmt.Sprintf(":%s", cfg.HTTPPort))
}
