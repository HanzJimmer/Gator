package main

import "fmt"

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
