package main

import (
	"context"
	"database/sql"
)

type NotesArgs struct {
	Limit     *int32
	Offset    *int32
	Direction *string
}

func (r *RootResolver) Notes(ctx context.Context, args NotesArgs) ([]*NoteResolver, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUserMustBeAuth
	}
	var rxs []*NoteResolver
	rows, err := db.Query(`
		select
			user_id,
			note_id,
			created_at,
			updated_at,
			data
		from notes
		where user_id = $1
		order by
			case when $4::text is null or $4 = 'desc' then updated_at end desc,
			case when $4 =  'asc' then updated_at end asc
		limit coalesce( $2, 25 )
		offset $3
	`, userID, args.Limit, args.Offset, args.Direction)
	// NOTE: Guard sql.ErrNoRows
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		note := &Note{}
		err := rows.Scan(&note.UserID, &note.NoteID, &note.CreatedAt, &note.UpdatedAt, &note.Data)
		if err != nil {
			return nil, err
		}
		rxs = append(rxs, &NoteResolver{note})
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return rxs, nil
}
