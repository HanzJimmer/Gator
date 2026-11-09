package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
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

	fmt.Printf("User created: %s\n", newUser.Name)

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
	if len(cmd.args) != 1 {
		return fmt.Errorf("please enter only a refresh time. it should be a number and a time digit (i.e. 's' = second, 'm' = minute, 'h' = hour)")
	}
	time_between_reqs := cmd.args[0]
	timeBetween, err := time.ParseDuration(time_between_reqs)
	if err != nil {
		return fmt.Errorf("refresh time should be a number and a time digit (i.e. 's' = second, 'm' = minute, 'h' = hour): %v", err)
	}

	fmt.Printf("Refreshing feeds every %v\n", timeBetween)
	ticker := time.NewTicker(timeBetween)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *state) error {
	RSSFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error retrieving next feed: %v", err)
	}

	fetchedParams := database.MarkFeedFetchedParams{LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true}, ID: RSSFeed.ID}

	err = s.db.MarkFeedFetched(context.Background(), fetchedParams)
	if err != nil {
		return fmt.Errorf("error marking feed as fetched: %v", err)
	}
	renewFeed, err := fetchFeed(context.Background(), RSSFeed.Url)
	if err != nil {
		return fmt.Errorf("error renewing feed: %v", err)
	}

	for _, item := range renewFeed.Channel.Item {
		var pubAt sql.NullTime
		if pubTime, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			pubAt = sql.NullTime{Time: pubTime, Valid: true}
		}

		postParams := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       sql.NullString{String: item.Title, Valid: true},
			Url:         item.Link,
			Description: sql.NullString{String: item.Description, Valid: true},
			PublishedAt: pubAt,
			FeedID:      RSSFeed.ID,
		}

		err = s.db.CreatePost(context.Background(), postParams)
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			log.Printf("Couldn't create post: %v", err)
			continue
		}
	}
	fmt.Printf("Feed '%v' renewed\n", renewFeed.Channel.Title)
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("not enough args")
	}

	feedName := cmd.args[0]
	feedURL := cmd.args[1]

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      feedName,
		Url:       feedURL,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("error creating feed-follow relationship: %v", err)
	}

	fmt.Printf("Feed added: %+v\n", feed.Name)
	fmt.Printf("Feed-Follow relation created: \nUser: %v\nFeed: %v", feedFollow.UserName, feedFollow.FeedName)

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("not enough args")
	}

	feedURL := cmd.args[0]
	feed, err := s.db.GetFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error retrieving feed: %v", err)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	fmt.Printf("%v now following %v", user.Name, feed.Name)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	followList, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("error retrieving follow list: %v", err)
	}

	fmt.Printf("Current User: %v\nFeeds Followed:\n", user.Name)
	for _, follow := range followList {
		fmt.Printf("Feed Name: %v\n", follow.FeedName)
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeedsWithUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error parsing feeds %v", err)
	}
	for _, feed := range feeds {
		fmt.Printf("Feed Name: %v\nFeed URL: %v\nOwner: %v\n\n", feed.Name, feed.Url, feed.UserName.String)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("not enough args")
	}

	feedURL := cmd.args[0]
	feed, err := s.db.GetFeed(context.Background(), feedURL)
	if err != nil {
		return fmt.Errorf("error retrieving feed: %v\n", err)
	}

	unfollowParams := database.UnfollowParams{UserID: user.ID, FeedID: feed.ID}
	err = s.db.Unfollow(context.Background(), unfollowParams)
	if err != nil {
		return fmt.Errorf("error unfollowing: %v\n", err)
	}
	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	titleLimit := 2
	var err error
	if len(cmd.args) == 1 {
		titleLimit, err = strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("please enter a valid number of posts to browse: %v\n", err)
		}
	}

	postUserParam := database.GetPostsForUserParams{UserID: user.ID, Limit: int32(titleLimit)}

	userPosts, err := s.db.GetPostsForUser(context.Background(), postUserParam)
	if err != nil {
		return fmt.Errorf("error retrieving posts for user: %v\n", err)
	}

	fmt.Printf("Printing posts for %s\n", user.Name)
	for _, post := range userPosts {
		fmt.Println(post.Title)
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
