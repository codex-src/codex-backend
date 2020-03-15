package main

import (
	"context"
	"errors"

	graphql "github.com/graph-gophers/graphql-go"
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

// // NOTE: Unprotected
// func (r *RootResolver) User(ctx context.Context, args struct{ UserID graphql.ID }) (*UserResolver, error) {
// 	// userID, ok := ctx.Value(UserIDKey).(string)
// 	// if !ok {
// 	// 	return nil, ErrUserMustBeAuth
// 	// }
// 	user := &User{}
// 	err := db.QueryRow(`
// 		select
// 			user_id,
// 			note_id,
// 			created_at,
// 			updated_at,
// 			data
// 		from notes
// 		where note_id = $1
// 	`, args.NoteID).Scan(&note.UserID, &note.NoteID, &note.CreatedAt, &note.UpdatedAt, &note.Data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &NoteResolver{note}, nil
// }

// NOTE: Unprotected
func (r *RootResolver) User(ctx context.Context, args struct{ UserID graphql.ID }) (*UserResolver, error) {
	// userID, ok := ctx.Value(UserIDKey).(string)
	// if !ok {
	// 	return nil, ErrUserMustBeAuth
	// }
	user := &User{}
	err := db.QueryRow(`
		select
			user_id,
			-- created_at,
			-- updated_at,
			-- email,
			-- email_verified,
			-- auth_provider,
			photo_url,
			display_name,
			username
		from users
		where user_id = $1
	`, args.UserID).Scan(&user.UserID, &user.PhotoURL, &user.DisplayName, &user.Username)
	if err != nil {
		return nil, err
	}
	return &UserResolver{user}, nil
}

// NOTE: Unprotected
func (r *RootResolver) Note(ctx context.Context, args struct{ NoteID graphql.ID }) (*NoteResolver, error) {
	// userID, ok := ctx.Value(UserIDKey).(string)
	// if !ok {
	// 	return nil, ErrUserMustBeAuth
	// }
	note := &Note{}
	err := db.QueryRow(`
		select
			user_id,
			note_id,
			created_at,
			updated_at,
			data
		from notes
		where note_id = $1
	`, args.NoteID).Scan(&note.UserID, &note.NoteID, &note.CreatedAt, &note.UpdatedAt, &note.Data)
	if err != nil {
		return nil, err
	}
	return &NoteResolver{note}, nil
}
