package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Post struct {
	ID        int64     `json:"id"`
	Content   string    `json:"content"`
	Title     string    `json:"title"`
	UserID    int64     `json:"user_id"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Version   int64     `json:"version"`
	Comments  []Comment `json:"comments"`
	User      User      `json:"user"`
}

type PostWithMetaData struct {
	Post
	CommentCount int64 `json:"comment_count"`
}

type PostStore struct {
	db *sql.DB
}

func (s *PostStore) GetUserFeed(ctx context.Context, userID int64, fq PaginatedFeedQuery) ([]PostWithMetaData, error) {
	sort := "DESC"
	if fq.Sort == "asc" {
		sort = "ASC"
	}
	// Base query with placeholders for limit/offset/search
	query := `
		SELECT 
			p.id, p.user_id, p.title, p.content, p.tags, p.created_at, p.version,
			COUNT(c.id) AS comments_count
		FROM posts p
		LEFT JOIN comments c ON c.post_id = p.id
		LEFT JOIN followers f ON f.user_id = $1 AND f.follower_id = p.user_id
		WHERE 
			(f.follower_id IS NOT NULL OR p.user_id = $1)
			AND (p.title ILIKE '%' || $4 || '%' OR p.content ILIKE '%' || $4 || '%')
	`

	args := []any{userID, fq.Limit, fq.Offset, fq.Search}
	paramCounter := 4

	// Add tags filter ONLY if tags slice is non-empty
	if len(fq.Tags) > 0 {
		paramCounter++
		query += fmt.Sprintf(" AND p.tags @> $%d", paramCounter)
		args = append(args, pq.Array(fq.Tags))
	}

	// Add since filter only if provided
	if fq.Since != nil {
		paramCounter++
		query += fmt.Sprintf(" AND p.created_at > $%d", paramCounter)
		args = append(args, *fq.Since)
	}

	// Add until filter only if provided
	if fq.Until != nil {
		paramCounter++
		query += fmt.Sprintf(" AND p.created_at < $%d", paramCounter)
		args = append(args, *fq.Until)
	}

	// Complete query
	query += `
		GROUP BY p.id
		ORDER BY p.created_at ` + sort + `
		LIMIT $2 OFFSET $3
	`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feed []PostWithMetaData
	for rows.Next() {
		var post PostWithMetaData
		err := rows.Scan(
			&post.ID,
			&post.UserID,
			&post.Title,
			&post.Content,
			pq.Array(&post.Tags),
			&post.CreatedAt,
			&post.Version,
			&post.CommentCount,
		)
		if err != nil {
			return nil, err
		}
		feed = append(feed, post)
	}

	return feed, rows.Err()
}
func (s *PostStore) Create(ctx context.Context, post *Post) error {
	query := `INSERT INTO posts(content, title, user_id, tags, created_at, updated_at) VALUES ($1, $2, $3, $4,$5,$6) RETURNING id, created_at, updated_at`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, post.Content, post.Title, post.UserID, pq.Array(post.Tags), time.Now(), time.Now()).Scan(&post.ID, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		return err
	}
	return nil
}
func (s *PostStore) GetByID(ctx context.Context, id int64) (*Post, error) {
	query := `
        SELECT id, title, content, user_id, tags, created_at, updated_at ,version
        FROM posts 
        WHERE id = $1`

	var post Post
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&post.ID,
		&post.Title,
		&post.Content,
		&post.UserID,
		pq.Array(&post.Tags),
		&post.CreatedAt,
		&post.UpdatedAt,
		&post.Version,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrNotFound
		default:
			return nil, err
		}
	}

	return &post, nil
}
func (s *PostStore) DeleteByID(ctx context.Context, id int64) error {
	query := `DELETE FROM posts WHERE id = $1`
	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	res, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
func (s *PostStore) UpdateByID(ctx context.Context, post *Post) error {
	exists, err := s.exists(ctx, post.ID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	query := `
        UPDATE posts
        SET content = $2, title = $3, tags = $4, version = version + 1, updated_at = $6
        WHERE id = $1 AND version = $5
        RETURNING version`

	ctx, cancel := context.WithTimeout(ctx, QueryTimeoutDuration)
	defer cancel()
	err = s.db.QueryRowContext(ctx, query,
		post.ID,
		post.Content,
		post.Title,
		pq.Array(post.Tags),
		post.Version,
		time.Now(),
	).Scan(&post.Version)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrConflict
		default:
			return err
		}
	}
	return nil
}
