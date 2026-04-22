package initializer

import "api/configs"

func InitConfig() {
	_ = configs.Load()
}
