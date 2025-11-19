package initializer

import "api/analytics"

func Run() {
	InitConfig()
	InitDB()
	analytics.StartETLWorker()
	r := InitRouter()
	r.Run(":8080")
}
