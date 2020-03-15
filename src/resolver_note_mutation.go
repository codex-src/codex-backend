package main

import (
	"context"
	"log"

	graphql "github.com/graph-gophers/graphql-go"
)

type NoteInput struct {
	NoteID graphql.ID
	Data   string
}

func (r *RootResolver) CreateNote(ctx context.Context, args struct{ NoteInput NoteInput }) (*bool, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUserMustBeAuth
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
		insert into notes (
			user_id,
			note_id,
			data
		) values ( $1, $2, $3 )
	`, userID, args.NoteInput.NoteID, args.NoteInput.Data)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (r *RootResolver) DeleteNote(ctx context.Context, args struct{ NoteID graphql.ID }) (*bool, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUserMustBeAuth
	}
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	_, err = tx.Exec(`
		delete
		from notes
		where
			user_id = $1 and
			note_id = $2
	`, userID, args.NoteID)
	if err != nil {
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		return nil, err
	}
	log.Printf("deleted note noteID=%s from user userID=%s", args.NoteID, userID)
	return nil, nil
}
