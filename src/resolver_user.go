package main

import (
	"context"
	"time"

	graphql "github.com/graph-gophers/graphql-go"
)

type User struct {
	UserID        graphql.ID
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Email         string
	EmailVerified bool
	AuthProvider  string
	PhotoURL      *string
	DisplayName   *string
	Username      *string
}

type UserResolver struct{ user *User }

func (r *UserResolver) UserID() graphql.ID {
	return r.user.UserID
}

func (r *UserResolver) CreatedAt() string {
	return r.user.CreatedAt.UTC().Format(PostgresTimestamptzFormat)
}

func (r *UserResolver) UpdatedAt() string {
	return r.user.UpdatedAt.UTC().Format(PostgresTimestamptzFormat)
}

func (r *UserResolver) Email() string {
	return r.user.Email
}

func (r *UserResolver) EmailVerified() bool {
	return r.user.EmailVerified
}

func (r *UserResolver) AuthProvider() string {
	return r.user.AuthProvider
}

func (r *UserResolver) PhotoURL() *string {
	// if r.user.PhotoURL == nil {
	// 	return nil
	// }
	return r.user.PhotoURL
}

func (r *UserResolver) DisplayName() *string {
	// if r.user.DisplayName == nil {
	// 	return nil
	// }
	return r.user.DisplayName
}

func (r *UserResolver) Username() *string {
	// if r.user.Username == nil {
	// 	return nil
	// }
	return r.user.Username
}

func (r *UserResolver) Notes(ctx context.Context, args struct{ Limit, Offset *int32 }) ([]NoteResolver, error) {
	return RootRx.Notes(ctx, args)
}
