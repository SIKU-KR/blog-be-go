package main

import (
	"bumsiku/controller"
)

func main() {
	r := controller.SetupRouter()
	r.Run()
}
