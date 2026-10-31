package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"time"

	"github.com/HanzJimmer/Gator/internal/database"
	. "github.com/HanzJimmer/Gator/internal/rss"
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

	user, err := s.db.GetUser(context.Background(), cmd.args[0])

	if err != nil {
		return fmt.Errorf("user does not exit")
	}

	if err := s.State.SetUser(user.Name); err != nil {
		return err
	}
	fmt.Println("Username has been set")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("not enough arguments to register user")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])

	if err == nil {
		return fmt.Errorf("user already exists")
	}

	userParams := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	newUser, err := s.db.CreateUser(context.Background(), userParams)

	if err != nil {
		return fmt.Errorf("error when creating new user: %w", err)
	}

	s.State.SetUser(newUser.Name)

	fmt.Printf("User created: %s", newUser.Name)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetTable(context.Background())

	if err != nil {
		return fmt.Errorf("error resetting the table: %v", err)
	}

	fmt.Println("Table reset successfully")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving user list: %v", err)
	}

	for _, user := range users {
		if s.State.CurrentUserName == user.Name {
			fmt.Printf("* %s (current)\n", user.Name)
			continue
		}
		fmt.Printf("* %s\n", user.Name)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	feedURL := "https://www.wagslane.dev/index.xml"
	rssFeed, err := fetchFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error retrieving RSS feed: %v", err)
	}

	fmt.Printf("Feed struct: %v", rssFeed)
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

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var newRSSFeed RSSFeed

	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gator")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if err := xml.Unmarshal(data, &newRSSFeed); err != nil {
		return nil, err
	}

	newRSSFeed.Channel.Title = html.UnescapeString(newRSSFeed.Channel.Title)
	newRSSFeed.Channel.Description = html.UnescapeString(newRSSFeed.Channel.Description)

	for i := range newRSSFeed.Channel.Item {
		newRSSFeed.Channel.Item[i].Title = html.UnescapeString(newRSSFeed.Channel.Item[i].Title)
		newRSSFeed.Channel.Item[i].Description = html.UnescapeString(newRSSFeed.Channel.Item[i].Description)
	}

	return &newRSSFeed, nil
}
