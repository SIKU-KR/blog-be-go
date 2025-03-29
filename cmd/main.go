package main

import (
	"bumsiku/controller"
	"bumsiku/internal/config"
	"os"
)

func main() {
	if os.Getenv("GO_ENV") != "production" {
		config.LoadEnv()
	}

	r := controller.SetupRouter()
	r.Run()
}
