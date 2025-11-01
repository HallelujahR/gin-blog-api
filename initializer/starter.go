package initializer

func Run() {
	InitConfig()
	InitDB()
	r := InitRouter()
	r.Run(":8080")
}
