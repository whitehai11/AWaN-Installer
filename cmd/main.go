package main

import (
	"log"

	"github.com/whitehai11/AWaN-Installer/installer"
	"github.com/whitehai11/AWaN-Installer/ui/gui"
	"github.com/whitehai11/AWaN-Installer/ui/tui"
	"github.com/whitehai11/AWaN-Installer/utils"
)

func main() {
	logger := utils.NewLogger()
	flowInstaller, err := installer.New(logger)
	if err != nil {
		log.Fatal(err)
	}

	environment := installer.DetectEnvironment()
	logger.Log("INSTALL", "Detected environment "+environment)

	switch environment {
	case installer.EnvironmentDesktop:
		if err := gui.Run(flowInstaller); err != nil {
			logger.Log("INSTALL", "Falling back to terminal installer: "+err.Error())
			if err := tui.RunInstaller(flowInstaller); err != nil {
				log.Fatal(err)
			}
		}
	case installer.EnvironmentServer, installer.EnvironmentUnknown:
		if err := tui.RunInstaller(flowInstaller); err != nil {
			log.Fatal(err)
		}
	default:
		if err := tui.RunInstaller(flowInstaller); err != nil {
			log.Fatal(err)
		}
	}
}
