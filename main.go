package main

import (
	"context"
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
	handlerCommands.register("agg", handlerAgg)
	handlerCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	handlerCommands.register("feeds", handlerFeeds)
	handlerCommands.register("follow", middlewareLoggedIn(handlerFollow))
	handlerCommands.register("following", middlewareLoggedIn(handlerFollowing))
	handlerCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))

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

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.State.CurrentUserName)
		if err != nil {
			return fmt.Errorf("no current user: %v", err)
		}
		return handler(s, cmd, user)
	}
}
