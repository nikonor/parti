package main

import (
	"context"
	"time"
)

var (
	testBody = TestMSG{
		Title:     "title",
		Message:   "testing",
		CreatedAt: time.Now().UTC(),
	}
)

// TestMSG - тестовая структура
type TestMSG struct {
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func rmqInit(ctx context.Context) (context.Context, error) {
	return ctx, nil
}
