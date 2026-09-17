package main

import (
	"fmt"
	"log"
	"os"

	"github.com/HanzJimmer/Gator/internal/config"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading before config file: %s", err)
	}

	gatorState := state{&cfg}

	var handlerCommands = commands{make(map[string]func(*state, command) error)}

	handlerCommands.register("login", handlerLogin)

	args := os.Args
	if len(args) < 2 {
		log.Fatalf("not enough arguments")
	}

	commandName := args[1]
	commandArgs := args[2:]

	handlerCmd := command{commandName, commandArgs}

	handlerCommands.run(&gatorState, handlerCmd)

	afterConfig, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading after config file: %s", err)
	}

	fmt.Println(afterConfig)
}
