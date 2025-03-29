package main

import (
	"encoding/gob"
	"os"
	"time"

	"bumsiku/controller"
	"bumsiku/internal/config"
)

func main() {
	if os.Getenv("GO_ENV") != "production" {
		config.LoadEnv()
	}

	gob.Register(time.Time{})

	r := controller.SetupRouter()
	r.Run()
}
