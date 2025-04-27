package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"go-sso/internal/domain/models"
	"go-sso/internal/storage"

	"github.com/lib/pq"
	_ "github.com/lib/pq" // Importing pq for PostgreSQL driver
)

type Storage struct {
	db *sql.DB
}

// New создает новое подключение к базе данных PostgreSQL
func New(connStr string) (*Storage, error) {
	const op = "storage.postgres.New"

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

// SaveUser сохраняет пользователя в базе данных
func (s *Storage) SaveUser(
	ctx context.Context,
	email string,
	passHash []byte,
) (int64, error) {
	const op = "storage.postgres.SaveUser"

	query := `
        INSERT INTO users (email, pass_hash) 
        VALUES ($1, $2) 
        RETURNING id`

	var id int64
	err := s.db.QueryRowContext(ctx, query, email, passHash).Scan(&id)
	if err != nil {
		var psqlErr *pq.Error

		if errors.As(err, &psqlErr) && psqlErr.Code == storage.ErrUniqueViolation {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrUserExists)
		}

		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}

// User возвращает пользователя по его email
func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.postgres.User"

	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, email, pass_hash
		FROM users
		WHERE email = $1`)
	if err != nil {
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, email)

	var user models.User
	err = row.Scan(&user.ID, &user.Email, &user.PassHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, fmt.Errorf("%s: %w", op, storage.ErrUserNotFound)
		}

		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return user, nil
}

func (s *Storage) IsAdmin(ctx context.Context, uid int64) (bool, error) {
	const op = "storage.postgres.IsAdmin"

	stmt, err := s.db.PrepareContext(ctx, `
		SELECT is_admin
		FROM users 
		WHERE id = $1`)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, uid)

	var isAdmin bool
	err = row.Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

func (s *Storage) App(ctx context.Context, appID int) (models.App, error) {
	const op = "storage.postgres.App"

	stmt, err := s.db.PrepareContext(ctx, `
		SELECT id, name, secret
		FROM apps
		WHERE id = $1`)
	if err != nil {
		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}
	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, appID)

	var app models.App
	err = row.Scan(&app.ID, &app.Name, &app.Secret)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.App{}, fmt.Errorf("%s: %w", op, storage.ErrAppNotFound)
		}

		return models.App{}, fmt.Errorf("%s: %w", op, err)
	}

	return app, nil
}
