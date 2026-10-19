package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/zsoltdzsugan/gatorcli/internal/database"
)

func handlerAgg(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %v <time_between_reqs>", cmd.Name)
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	log.Printf("Collecting feeds every %s...\n", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}

func scrapeFeeds(s *state) {
	nextFeed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		log.Fatalf("couldn't get next feed to fetch: %s", err)
		return
	}

	scrapeFeed(s.db, nextFeed)
}

func scrapeFeed(db *database.Queries, feed database.Feed) {
	_, err := db.MarkFetchedFeed(context.Background(), feed.ID)
	if err != nil {
		log.Fatalf("couldn't mark feed as fetched: %s", err)
		return
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		log.Fatalln(err)
	}

	existingPosts, err := getPostsMap(db)
	if err != nil {
		log.Printf("couldn't get existing posts: %v", err)
	}

	for _, item := range rssFeed.Channel.Item {
		fmt.Printf("Title: %v\n", item.Title)

		if _, exists := existingPosts[item.Link]; exists {
			continue
		}

		var description sql.NullString
		if item.Description != "" {
			description = sql.NullString{
				String: item.Description,
				Valid:  true,
			}
		}

		publishedAt := parseDate(item.PubDate)

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Url:         item.Link,
			Description: description,
			PublishedAt: publishedAt,
			FeedID:      feed.ID,
		})
		if err != nil {
			log.Fatal("error in creating post")
		}

		existingPosts[item.Link] = struct{}{}
	}
}

func getPostsMap(db *database.Queries) (map[string]struct{}, error) {
	posts, err := db.GetPosts(context.Background())
	if err != nil {
		return nil, err
	}

	existingPosts := make(map[string]struct{}, len(posts))
	for _, post := range posts {
		existingPosts[post.Url] = struct{}{}
	}

	return existingPosts, nil
}

func parseDate(value string) sql.NullTime {
	value = strings.TrimSpace(value)
	if value == "" {
		return sql.NullTime{}
	}

	layouts := []string{
		time.RFC1123Z,
		time.RFC1123,
		time.RFC822Z,
		time.RFC822,
		time.RFC3339,
		time.RFC3339Nano,
		"Mon, 2 Jan 2006 15:04:05 MST",
		"Mon, 02 Jan 2006 15:04:05 -0700",
	}

	for _, layout := range layouts {
		parsed, err := time.Parse(layout, value)
		if err == nil {
			return sql.NullTime{
				Time:  parsed.UTC(),
				Valid: true,
			}
		}
	}

	return sql.NullTime{}
}
