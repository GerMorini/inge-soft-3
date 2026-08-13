package repository

import (
	"context"
	"errors"
	"fmt"

	identityerrors "github.com/gmorini/inge-soft-3/backend/internal/identity/errors"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CreateUserParams struct {
	FirstName    string
	LastName     string
	Phone        string
	Street       string
	StreetNumber string
	Apartment    *string
	City         string
	Province     string
	Username     string
	Email        string
	PasswordHash string
}

type CreatedUser struct {
	ID       int64
	Username string
	Email    string
}

type Credentials struct {
	ID           int64
	Username     string
	PasswordHash string
}

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (CreatedUser, error) {
	const query = `
		INSERT INTO users (
			first_name,
			last_name,
			phone,
			street,
			street_number,
			apartment,
			city,
			province,
			username,
			email,
			password_hash
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id, username, email`

	var created CreatedUser
	err := r.pool.QueryRow(
		ctx,
		query,
		params.FirstName,
		params.LastName,
		params.Phone,
		params.Street,
		params.StreetNumber,
		params.Apartment,
		params.City,
		params.Province,
		params.Username,
		params.Email,
		params.PasswordHash,
	).Scan(&created.ID, &created.Username, &created.Email)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			switch postgresError.ConstraintName {
			case "users_username_key":
				return CreatedUser{}, &identityerrors.ConflictError{Field: "username"}
			case "users_email_key":
				return CreatedUser{}, &identityerrors.ConflictError{Field: "email"}
			}
		}
		return CreatedUser{}, fmt.Errorf("insert user: %w", err)
	}
	return created, nil
}

func (r *Repository) FindCredentialsByUsername(ctx context.Context, username string) (Credentials, error) {
	const query = `
		SELECT id, username, password_hash
		FROM users
		WHERE username = $1`

	var credentials Credentials
	err := r.pool.QueryRow(ctx, query, username).Scan(
		&credentials.ID,
		&credentials.Username,
		&credentials.PasswordHash,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Credentials{}, identityerrors.ErrUserNotFound
	}
	if err != nil {
		return Credentials{}, fmt.Errorf("find credentials by username: %w", err)
	}
	return credentials, nil
}
