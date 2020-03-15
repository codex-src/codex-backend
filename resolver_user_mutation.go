package main

import (
	"context"
	"log"

	graphql "github.com/graph-gophers/graphql-go"
)

type RegisterUserInput struct {
	UserID        graphql.ID
	Email         string
	EmailVerified bool
	AuthProvider  string
	PhotoURL      *string
	DisplayName   *string
}

func (r *RootResolver) RegisterUser(ctx context.Context, args struct{ UserInput RegisterUserInput }) (*UserResolver, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
		insert into users (
			user_id,
			email,
			email_verified,
			auth_provider,
			photo_url,
			display_name
		) values ( $1, $2, $3, $4, $5, $6 )
	`, args.UserInput.UserID, args.UserInput.Email, args.UserInput.EmailVerified, args.UserInput.AuthProvider, args.UserInput.PhotoURL, args.UserInput.DisplayName)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	log.Printf("registered user userID=%s", args.UserInput.UserID)
	return RootRx.Me(ctx)
}
