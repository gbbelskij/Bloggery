package storage

import (
	"context"
	"errors"
	"fmt"
	"new_service/internal/handlers/structs"
	"new_service/internal/models"
	custom_errors "new_service/internal/repository"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Storage struct {
	Conn *pgxpool.Pool
}

func New(connectString string) (*Storage, error) {
	const op = "repository.storage.New"

	conn, err := pgxpool.New(context.Background(), connectString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{Conn: conn}, nil

}

func (s *Storage) SaveUser(email string, password string, username string) error {
	const op = "repository.storage.SaveUser"

	hash_password, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	if username != "" {
		_, err = s.Conn.Exec(context.Background(),
			`INSERT INTO users(email, username, password_hash) VALUES($1, $2, $3)`,
			email, username, hash_password,
		)
	} else {
		_, err = s.Conn.Exec(context.Background(),
			`INSERT INTO users(email, password_hash) VALUES($1, $2)`,
			email, hash_password,
		)
	}
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserPasswordByEmail(email string) (string, uuid.UUID, error) {
	const op = "repository.storage.GetUser"

	row := s.Conn.QueryRow(
		context.Background(),
		`SELECT password_hash, user_id FROM users WHERE email = $1`,
		email,
	)
	var password_hash string
	var user_id uuid.UUID

	err := row.Scan(&password_hash, &user_id)
	if err != nil {
		return "", user_id, fmt.Errorf("%s: %w", op, err)
	}
	return password_hash, user_id, nil
}

func (s *Storage) GetUserPasswordByUsername(username string) (string, uuid.UUID, error) {
	const op = "repository.storage.GetUser"

	row := s.Conn.QueryRow(
		context.Background(),
		`SELECT password_hash, user_id FROM users WHERE username = $1`,
		username,
	)
	var password_hash string
	var user_id uuid.UUID

	err := row.Scan(&password_hash, &user_id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", user_id, custom_errors.ErrUserDoesNotExist
		}
		return "", user_id, fmt.Errorf("%s: %w", op, err)
	}
	return password_hash, user_id, nil
}

func (s *Storage) SavePost(post *models.Post) error {
	const op = "repository.storage.SavePost"

	_, err := s.Conn.Exec(
		context.Background(),
		`INSERT INTO posts(user_id, title, content) VALUES($1, $2, $3)`,
		post.UserId, post.Title, post.Content,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) GetNextPosts(userId uuid.UUID, paginationParams structs.PaginationParams) ([]models.DbPost, error) {
	const op = "repository.storage.GetNextPosts"

	var query string
	if paginationParams.Reverse {
		query = `SELECT * FROM (
				SELECT * FROM posts
				WHERE user_id = $1 AND created_at > $2
				ORDER BY created_at ASC
				LIMIT $3
				) AS subquery
				ORDER BY created_at DESC;`
	} else {
		query = `SELECT * FROM posts
				WHERE user_id = $1 AND created_at < $2
				ORDER BY created_at DESC
				LIMIT $3;`
	}

	posts, err := s.Conn.Query(
		context.Background(),
		query,
		userId, paginationParams.Cursor, paginationParams.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer posts.Close()

	parsedPosts, err := pgx.CollectRows(posts, pgx.RowToStructByName[models.DbPost])
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return parsedPosts, nil
}

func (s *Storage) GetPost(postId uuid.UUID) (models.DbPost, error) {
	const op = "repository.storage.GetPost"

	post, err := s.Conn.Query(
		context.Background(),
		`SELECT * FROM posts WHERE post_id = $1`,
		postId,
	)
	if err != nil {
		return models.DbPost{}, fmt.Errorf("%s: %w", op, err)
	}

	parsed_post, err := pgx.CollectOneRow(post, pgx.RowToStructByName[models.DbPost])
	if err != nil {
		return models.DbPost{}, fmt.Errorf("%s: %w", op, err)
	}

	return parsed_post, nil
}

func (s *Storage) DeletePost(postId uuid.UUID) error {
	const op = "repository.storage.DeletePost"

	_, err := s.Conn.Exec(
		context.Background(),
		`DELETE FROM posts WHERE post_id = $1`,
		postId,
	)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
