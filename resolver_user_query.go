package main

import (
	"context"
	"fmt"
)

func (r *RootResolver) Me(ctx context.Context) (*UserResolver, error) {
	userID, ok := ctx.Value(UserIDKey).(string)
	if !ok {
		return nil, ErrUserMustBeAuth
	}
	user := &User{}
	err := db.QueryRow(`
		select
			user_id,
			created_at,
			updated_at,
			email,
			email_verified,
			auth_provider,
			photo_url,
			display_name,
			username
		from users
		where user_id = $1
	`, userID).Scan(&user.UserID, &user.CreatedAt, &user.UpdatedAt, &user.Email, &user.EmailVerified, &user.AuthProvider, &user.PhotoURL, &user.DisplayName, &user.Username)
	if err != nil {
		return nil, fmt.Errorf("query me: %w", err)
	}
	return &UserResolver{user}, nil
}
