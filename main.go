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
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/devjoshbrown/gator/internal/config"
	"github.com/devjoshbrown/gator/internal/database"
	_ "github.com/lib/pq"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type commands struct {
	commandDict map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.commandDict[cmd.name]
	if !ok {
		return fmt.Errorf("command invalid: %s", cmd.name)
	}
	return handler(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.commandDict[name] = f
}

type command struct {
	name string
	args []string
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login requires a username argument")
	}
	if len(cmd.args) >= 2 {
		return fmt.Errorf("too many arguments, login requires a username argument")
	}

	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("user does not exist %w", err)
	}

	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return err
	}

	fmt.Printf("username set to %s\n", s.cfg.User)
	return nil

}
func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register requires a username argument")
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	})
	if err != nil {
		return fmt.Errorf("could not create user %w:", err)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Println("User Created Successfully")
	fmt.Printf("%+v\n", user)
	return nil
}
func handlerAddfeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 2 {
		return fmt.Errorf("Error: must pass exactly 2 args")
	}

	feed, err := s.db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	})
	if err != nil {
		return fmt.Errorf("could not create feed: %w", err)
	}

	_, err = s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})

	if err != nil {
		return fmt.Errorf("failed to CreateFeedFollow: %w", err)
	}
	fmt.Printf("%v\n", feed)

	return nil

}
func handlerFeeds(s *state, cmd command) error {
	if len(cmd.args) > 0 {
		return fmt.Errorf("Error: Feeds command takes 0 arguments")
	}

	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get feeds : %w", err)
	}

	for _, feed := range feeds {
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(feed.UserName)
	}
	return nil
}
func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Follow must have only 1 argument - url")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("GetFeedByURL failed in handler Follow : %w", err)
	}
	followers, err := s.db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to CreateFeedFollow: %w", err)
	}
	fmt.Println(followers)
	return nil
}
func handlerReset(s *state, cmd command) error {

	err := s.db.Reset(context.Background())
	if err != nil {
		return fmt.Errorf("could not reset database %w:", err)
	}

	fmt.Println("Users Reset Successfully")
	return nil
}
func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("could not collect users data %w:", err)
	}
	for _, user := range users {
		if user.Name == s.cfg.User {
			fmt.Printf("%v (current)\n", user.Name)
		} else {
			fmt.Printf("%v\n", user.Name)
		}
	}
	return nil
}
func handlerGetFollowers(s *state, cmd command, user database.User) error {

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get feed follows for user: %w", err)
	}

	for _, follow := range follows {
		fmt.Println(follow.FeedName)
	}
	return nil

}
func handlerAgg(s *state, cmd command) error {
	index := "https://www.wagslane.dev/index.xml"
	feed, err := fetchFeed(context.Background(), index)
	if err != nil {
		return fmt.Errorf("failed to fetch feed at wags/index")
	}
	fmt.Println(feed)
	return nil
}
func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	request, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to generate the feed request %w", err)
	}

	request.Header.Set("User-Agent", "gator")

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to send the feed request %w", err)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("Failed to read response body : %w", err)
	}

	var feed RSSFeed
	err = xml.Unmarshal(data, &feed)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal xml file %w", err)
	}

	feed.Channel.Title = html.UnescapeString(feed.Channel.Title)
	feed.Channel.Description = html.UnescapeString(feed.Channel.Description)

	for i, item := range feed.Channel.Item {
		feed.Channel.Item[i].Title = html.UnescapeString(item.Title)
		feed.Channel.Item[i].Description = html.UnescapeString(item.Description)
	}

	return &feed, nil
}
func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) != 1 {
		return fmt.Errorf("Follow must have only 1 argument - url")
	}

	feed, err := s.db.GetFeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("failed to get feed by URL: %w", err)
	}

	err = s.db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete follower: %w", err)
	}
	return nil
}
func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {
		user, err := s.db.GetUser(context.Background(), s.cfg.User)
		if err != nil {
			return fmt.Errorf("user not logged in: %w", err)
		}
		return handler(s, cmd, user)
	}
}

/*
Terminal:  go run . register lane

	   ↑         ↑
	command     argument
*/
func main() {

	/*2.  Read config from ~/.gatorconfig.json */
	Newcfg, err := config.Read()
	if err != nil {
		log.Fatal(err)
	}

	/* 3. Open connection to Postgres database */
	db, err := sql.Open("postgres", Newcfg.Url)
	if err != nil {
		log.Fatal(err)
	}

	dbQueries := database.New(db)

	/*
		4. Pack both into "state" — this is the toolbox every command gets
		  state = { db: ..., cfg: ... }
	*/
	s := state{
		db:  dbQueries,
		cfg: &Newcfg,
	}

	/* 5. Register all available commands (login, register, reset, etc.) */
	c := commands{
		commandDict: make(map[string]func(*state, command) error),
	}
	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerGetUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddfeed))
	c.register("feeds", handlerFeeds)
	c.register("following", middlewareLoggedIn(handlerGetFollowers))
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))

	/* Look at what the user typed (os.Args) */
	a := os.Args

	if len(a) < 2 {
		log.Fatal("Not enough arguments")
	}

	/* 7. Find the matching handler function and call it */
	command_name := a[1]
	command_args := a[2:]

	cmd := command{
		name: command_name,
		args: command_args,
	}

	/* 8. The handler does its work using state, then exits */
	err = c.run(&s, cmd)
	if err != nil {
		log.Fatal(err)
	}

}
