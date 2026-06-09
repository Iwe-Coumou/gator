package main

import (
	"fmt"
)

type command struct {
	name string
	args []string
}

type commands struct {
	handlerMap map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlerMap[cmd.name]
	if !ok {
		return fmt.Errorf("handler for %v does not exist in commands", cmd.name)
	}

	if err := handler(s, cmd); err != nil {
		return fmt.Errorf("%v: %w", cmd.name, err)
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlerMap[name] = f
}
