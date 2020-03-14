package main

import (
	"context"
	"errors"
)

const PostgresTimestamptzFormat = "2006-01-02 15:04:05.000000Z"

var (
	ErrUserMustBeUnauth = errors.New("user must be unauthenticated")
	ErrUserMustBeAuth   = errors.New("user must be authenticated")
)

var RootRx = RootResolver{}

type RootResolver struct{}

func (r *RootResolver) Ping(ctx context.Context) string {
	return "pong!"
}
