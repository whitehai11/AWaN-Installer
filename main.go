package main

import (
	"github.com/whitehai11/AWaN-Installer/installer"
	"github.com/whitehai11/AWaN-Installer/ui/gui"
	"github.com/whitehai11/AWaN-Installer/utils"
)

func main() {
	logger := utils.NewLogger()
	flowInstaller, err := installer.New(logger)
	if err != nil {
		panic(err)
	}

	if err := gui.Run(flowInstaller); err != nil {
		panic(err)
	}
}
