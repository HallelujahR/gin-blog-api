package initializer

import (
	"api/analytics"
	"api/configs"
	"fmt"
)

func Run() {
	InitConfig()
	InitDB()
	cfg := configs.Load()
	if cfg.AnalyticsEnabled {
		analytics.StartETLWorker(cfg.AnalyticsETL)
	}
	r := InitRouter()
	_ = r.Run(fmt.Sprintf(":%s", cfg.HTTPPort))
}
