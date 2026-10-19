package main

import (
	"context"
	"fmt"
	"strconv"
)

func handlerGetPosts(s *state, cmd command) error {
	limit := 2

	if len(cmd.Args) > 0 {
		parsedLimit, err := strconv.Atoi(cmd.Args[0])
		if err != nil {
			return fmt.Errorf("invalid limit %q: must be an integer", cmd.Args[0])
		}
		if parsedLimit <= 0 {
			return fmt.Errorf("limit must be greater than 0")
		}
		limit = parsedLimit
	}
	posts, err := s.db.GetPosts(context.Background(), int32(limit))
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Println(post)
	}
	return nil
}
