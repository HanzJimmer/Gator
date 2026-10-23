package main

import (
	"context"
	"fmt"
	"time"

	"github.com/HanzJimmer/Gator/internal/database"
	"github.com/google/uuid"
)

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("not enough arguments for login")
	}
	if err := s.State.SetUser(cmd.args[0]); err != nil {
		return err
	}
	fmt.Println("Username has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("not enough arguments to register user")
	}

	var newUser database.User

	if newUser, err := s.db.GetUser(context.Background(), cmd.args[0]); err == nil {
		return fmt.Errorf("user already exists")
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	if newUser, err := s.db.CreateUser(context.Background(), userParams); err != nil {

	}

	return nil
}

type commands struct {
	commands map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, exists := c.commands[cmd.name]
	if !exists {
		return fmt.Errorf("command does not exist")
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commands[name] = f
}
