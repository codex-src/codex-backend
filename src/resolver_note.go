package main

import (
	"time"

	graphql "github.com/graph-gophers/graphql-go"
)

type Note struct {
	UserID    graphql.ID
	NoteID    graphql.ID
	CreatedAt time.Time
	UpdatedAt time.Time
	Data      string
}

type NoteResolver struct{ note *Note }

func (r *NoteResolver) UserID() graphql.ID {
	return r.note.UserID
}

func (r *NoteResolver) NoteID() graphql.ID {
	return r.note.NoteID
}

func (r *NoteResolver) CreatedAt() string {
	return r.note.CreatedAt.UTC().Format(PostgresTimestamptzFormat)
}

func (r *NoteResolver) UpdatedAt() string {
	return r.note.UpdatedAt.UTC().Format(PostgresTimestamptzFormat)
}

func (r *NoteResolver) Data() string {
	return r.note.Data
}

// func (r *NoteResolver) DataShort() string {
// 	return fmt.Sprintf("%0.*s", 250, r.note.Data)
// }
//
// func (r *NoteResolver) DataMedium() string {
// 	return fmt.Sprintf("%0.*s", 500, r.note.Data)
// }
//
// func (r *NoteResolver) DataLong() string {
// 	return fmt.Sprintf("%0.*s", 1e3, r.note.Data)
// }
