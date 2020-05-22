package plugins

import (
	"context"
)

type FinancePlugin struct {
	Ctx      context.Context
	Commands map[string]Plugin
	Closed   chan struct{}
}

func New() *FinancePlugin {
	return &FinancePlugin{
		Ctx:      context.Background(),
		Commands: make(map[string]Plugin),
		Closed:   make(chan struct{}),
	}
}
