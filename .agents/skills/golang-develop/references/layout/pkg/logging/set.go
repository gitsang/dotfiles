package logging

import (
	"errors"
	"fmt"
	"log/slog"
	"strings"
)

type LogHandlerSet map[string]slog.Handler

type LogHandlerSetConfig struct {
	Default   string
	Providers map[string][]LogConfig
}

func NewLogHandlerSet(c LogHandlerSetConfig) (LogHandlerSet, error) {
	set := make(map[string]slog.Handler)

	for name, configs := range c.Providers {
		set[strings.ToLower(name)] = NewFanoutLogHandler(configs...)
	}

	if _, ok := set[strings.ToLower(c.Default)]; !ok {
		return nil, errors.New("default log handler not found")
	}

	return set, nil
}

func (s LogHandlerSet) Get(name string) slog.Handler {
	handler, ok := s[strings.ToLower(name)]
	if !ok {
		panic(fmt.Sprintf("log handler %s not found", name))
	}
	return handler
}
