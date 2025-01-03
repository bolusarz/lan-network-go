package main

import (
	"flag"
	"fmt"

	"reach.com/discovery/pkg/app"
)

func main() {
	deviceName := flag.String("name", "", "Device name")
	isMaster := flag.Bool("master", false, "Set this device as master")
	flag.Parse()

	if *deviceName == "" {
		panic("Device name is required")
	}

	app := app.NewApp(*deviceName)

	if *isMaster {
		app.SetAsMaster()
	}

	if err := app.Start(); err != nil {
		panic(fmt.Errorf("failed to start app: %v", err))
	}

	select {}
}