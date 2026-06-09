package main

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/Iwe-Coumou/gator/internal/database"
	"github.com/google/uuid"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username required")
	}

	userName := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), userName)
	if err != nil {
		return fmt.Errorf("user '%s' not found", userName)
	}

	if err = s.cfg.SetUser(userName); err != nil {
		return err
	}

	fmt.Println("User logged in successfully")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("username required")
	}
	userName := cmd.args[0]

	_, err := s.db.GetUser(context.Background(), userName)
	if err == nil {
		return fmt.Errorf("user '%s' already exists", userName)
	}

	params := database.CreateUserParams{ID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now(), Name: userName}
	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("creating user: %w", err)
	}

	if err := s.cfg.SetUser(userName); err != nil {
		return fmt.Errorf("setting new user as current: %w", err)
	}

	fmt.Println("User created")
	fmt.Printf("%+v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	if err := s.db.DeleteUsers(context.Background()); err != nil {
		return fmt.Errorf("resetting user table: %v", err)
	}

	fmt.Println("Successfully reset user table")
	return nil
}

func handlerListUser(s *state, cmd command) error {
	userNames, err := s.db.ListUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting user list")
	}

	current := s.cfg.CurrentUserName

	for _, userName := range userNames {
		var flag string
		if userName == current {
			flag = " (current)"
		} else {
			flag = ""
		}

		fmt.Printf("* %s%s\n", userName, flag)
	}
	return nil
}

func handlerAggregate(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("missing time between request argument")
	}

	fmt.Printf("Collecting feeds every %s", cmd.args[0])

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("parsing interval argument: %w", err)
	}

	ticker := time.NewTicker(duration)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("not enough arguments")
	}

	feedID := uuid.New()

	feedParams := database.AddFeedParams{
		ID:        feedID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	}

	if err := s.db.AddFeed(context.Background(), feedParams); err != nil {
		return fmt.Errorf("error adding feed %s@%s to user %s", cmd.args[0], cmd.args[1], user.Name)
	}

	feedFollowParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedID,
	}

	_, err := s.db.CreateFeedFollow(context.Background(), feedFollowParams)
	if err != nil {
		return fmt.Errorf("error creating feed follow")
	}

	fmt.Println("Feed added and followed")
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting feeds")
	}

	for _, feed := range feeds {
		fmt.Printf("Feed name: %s | Feed url: %s | Created by: %s\n", feed.Name, feed.Url, feed.Creator)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("feed url required")
	}

	feed, err := s.db.FeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("error getting feed by url")
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error creating feed follow")
	}

	fmt.Printf("Feed: %s | User: %s\n", feedFollow.FeedName, feedFollow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	feedFollows, err := s.db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("error getting follows for '%s'", user.Name)
	}

	fmt.Printf("Feeds followed by %s:\n", user.Name)
	for _, feedFollow := range feedFollows {
		fmt.Println(feedFollow.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("needs feed url to unfollow")
	}

	feed, err := s.db.FeedByURL(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("error getting feed by url")
	}

	if err = s.db.RemoveFeedFollow(context.Background(), database.RemoveFeedFollowParams{UserID: user.ID, FeedID: feed.ID}); err != nil {
		return fmt.Errorf("error removing feed follow from database")
	}

	fmt.Printf("%s unfollowed from %s\n", user.Name, feed.Name)
	return nil
}

func parseDate(s string) (time.Time, error) {
	formats := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC3339,
	}

	for _, format := range formats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date: %s", s)
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	var limit int32 = 2
	if len(cmd.args) > 0 {
		parsed, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = int32(parsed)
	}

	params := database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  limit,
	}

	posts, err := s.db.GetPostsForUser(context.Background(), params)
	if err != nil {
		return fmt.Errorf("error getting posts for user")
	}

	for _, post := range posts {
		printPost(post)
	}
	return nil
}

func handlerCurrentUser(_ *state, _ command, user database.User) error {
	fmt.Printf("current user: %s\n", user.Name)
	return nil
}

func printPost(p database.Post) {
	fmt.Printf("\033[1m%s\033[0m\t%s\n", p.Title, p.PublishedAt.Format("Jan 2, 2006"))
	fmt.Printf("%s\n", p.Description)
	fmt.Printf("%s\n", p.Url)
	fmt.Println()
}

func scrapeFeeds(s *state) error {

	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("error fetching next feed")
	}

	feed, err := fetchFeed(context.Background(), nextFeed.Url)
	if err != nil {
		return fmt.Errorf("error fetching RSS feed")
	}

	s.db.MarkFeedFetched(context.Background(), nextFeed.ID)

	for _, item := range feed.Channel.Item {
		published, err := parseDate(item.PubDate)
		if err != nil {
			fmt.Printf("error parsing date: %s\n", item.PubDate)
			continue
		}

		params := database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: published,
			FeedID:      nextFeed.ID,
		}

		if err = s.db.CreatePost(context.Background(), params); err != nil {
			fmt.Printf("error while creating post for %s@%v\n", item.Title, err)
			continue
		} else {
			fmt.Printf("Added %s to posts\n", item.Title)
		}

	}

	return nil
}
