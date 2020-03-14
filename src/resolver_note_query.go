package main

import (
	"context"
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
	if err != nil {
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

// func (r *RootResolver) Note(ctx context.Context, args struct{ NoteID graphql.ID }) (*NoteResolver, error) {
// 	currUser := CurrentSessionFromContext(ctx)
// 	if !currUser.IsAuth() {
// 		return nil, ErrUserMustBeAuth
// 	}
// 	note := &Note{}
// 	err := DB.QueryRow(`
// 		select
// 			user_id,
// 			note_id,
// 			created_at,
// 			updated_at,
// 			title_utf8_count,
// 			title,
// 			data_utf8_count,
// 			data
// 		from notes
// 		where
// 			note_id = $1 and
// 			( select user_id = $2 from notes where note_id = $3 )
// 	`, args.NoteID, currUser.UserID, args.NoteID).Scan(&note.UserID, &note.NoteID, &note.CreatedAt, &note.UpdatedAt, &note.TitleUTF8Count, &note.Title, &note.DataUTF8Count, &note.Data)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &NoteResolver{note}, nil
// }
