package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Follower struct {
	UserID     int64 `json:"user_id"`
	FollowerID int64 `json:"follower_id"`
	CreatedAt  int64 `json:"created_at"`
}
type FollowerStore struct {
	db *sql.DB
}

func (s *FollowerStore) Follow(ctx context.Context, FollowerId int64, userID int64) error {
	query := `
INSERT INTO followers (user_id, follower_id)
VALUES ($1, $2)
RETURNING follower_id;
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	_, err := s.db.ExecContext(ctx, query, userID, FollowerId)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code.Name() == "unique_violation" {
			return ErrConflict
		}
	}
	return nil
}
func (s *FollowerStore) Unfollow(ctx context.Context, FollowerId int64, UserId int64) error {
	query := `
	DELETE FROM followers 
    WHERE user_id = $1 AND follower_id = $2;
`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	_, err := s.db.ExecContext(ctx, query, UserId, FollowerId)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			return ErrConflict
		}
		return err
	}
	return nil
}
