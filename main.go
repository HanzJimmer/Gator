package main

import (
	"fmt"

	"github.com/HanzJimmer/Gator/internal/config"
)

func main() {
	beforeConfig, err := config.Read()
	if err != nil {
		fmt.Errorf("Error reading before config file: %s", err)
	}

	beforeConfig.SetUser()

	afterConfig, err := config.Read()
	if err != nil {
		fmt.Errorf("Error reading after config file: %s", err)
	}

	fmt.Println(afterConfig)
}
