package repositories

import (
	"context"
	"database/sql"
	"errors"

	"app-backend/models"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	CreateUser(ctx context.Context, email, passwordHash string) (models.User, error)
	GetByEmail(ctx context.Context, email string) (models.User, error)
	GetByID(ctx context.Context, id int64) (models.User, error)
}

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) CreateUser(ctx context.Context, email, passwordHash string) (models.User, error) {
	query := `
		INSERT INTO users (email, password_hash)
		VALUES ($1, $2)
		RETURNING id, email, password_hash, created_at`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email, passwordHash).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (models.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE email = $1`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, email).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	return user, nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id int64) (models.User, error) {
	query := `
		SELECT id, email, password_hash, created_at
		FROM users
		WHERE id = $1`

	var user models.User
	err := r.db.QueryRowContext(ctx, query, id).
		Scan(&user.ID, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.User{}, ErrUserNotFound
		}
		return models.User{}, err
	}

	return user, nil
}
