package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"

	"github.com/HanzJimmer/Gator/internal/config"
	"github.com/HanzJimmer/Gator/internal/database"
)

func main() {
	cfg, err := config.Read()
	if err != nil {
		fmt.Printf("Error reading before config file: %s", err)
	}

	db, err := sql.Open("postgres", cfg.DbUrl)
	dbQueries := database.New(db)

	gatorState := state{&cfg, dbQueries}

	var handlerCommands = commands{make(map[string]func(*state, command) error)}

	handlerCommands.register("login", handlerLogin)
	handlerCommands.register("register", handlerRegister)
	handlerCommands.register("reset", handlerReset)
	handlerCommands.register("users", handlerUsers)

	args := os.Args
	if len(args) < 2 {
		log.Fatalf("error: no command found")
	}

	commandName := args[1]
	commandArgs := args[2:]

	handlerCmd := command{commandName, commandArgs}

	err = handlerCommands.run(&gatorState, handlerCmd)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
