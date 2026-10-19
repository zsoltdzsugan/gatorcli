package main

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/zsoltdzsugan/gatorcli/internal/database"
)

func handlerLogin(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	name := cmd.Args[0]

	_, err := s.db.GetUserByName(context.Background(), name)
	if err != nil {
		return fmt.Errorf("user does not exists: %w", err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}

	fmt.Println("User switched successfully!")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.Args) != 1 {
		return fmt.Errorf("usage: %s <name>", cmd.Name)
	}

	name := cmd.Args[0]

	user, err := s.db.GetUserByName(context.Background(), name)
	if err == nil {
		return fmt.Errorf("user %q already exists", name)
	}

	id := uuid.New()

	newUser := database.CreateUserParams{
		ID:        id,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		Name:      name,
	}

	user, err = s.db.CreateUser(context.Background(), newUser)
	if err != nil {
		return fmt.Errorf("couldn't create user: %w", err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("couldn't set current user: %w", err)
	}
	fmt.Println("User created successfully!")
	fmt.Println(user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return fmt.Errorf("couldn't reset table: %w", err)
	}
	s.cfg.SetUser("")
	return nil
}

func handlerGetUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("no users in 'users' table: %w", err)
	}

	for i := range users {
		line := "* " + users[i]
		if users[i] == s.cfg.CurrentUserName {
			line += " (current)"
		}
		fmt.Println(line)
	}

	return nil
}
