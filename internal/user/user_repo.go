package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"VK/internal/utils/hashutils"
)

var (
	ErrUserNotFound = errors.New("No user record found")
	ErrBadPass      = errors.New("Bad password")
	ErrUserExists   = errors.New("User Exists")
)

type UserDB struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserDB {
	return &UserDB{
		db: db,
	}
}

func (u *UserDB) Create(ctx context.Context, login, passIn string) (*User, error) {

	pass := hashutils.HashPassword(passIn)

	user := &User{
		Login: login,
	}

	err := u.db.QueryRow(ctx, `
		INSERT INTO users (login, password) 
		VALUES ($1, $2)
		RETURNING id, login
	`, login, pass).Scan(&user.ID, &user.Login)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrUserExists
		}
		return nil, fmt.Errorf("insert error: %w", err)
	}

	return user, nil
}

func (u *UserDB) CheckPasswordByLogin(ctx context.Context, login, pass string) (*User, error) {
	row := u.db.QueryRow(ctx, "SELECT id, login, password FROM users WHERE login = $1", login)
	return u.passwordIsValid(pass, row)
}

func (u *UserDB) passwordIsValid(pass string, row pgx.Row) (*User, error) {
	var (
		dbPass string
		user   = &User{}
	)

	err := row.Scan(&user.ID, &user.Login, &dbPass)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("db scan error: %w", err)
	}

	ok, err := hashutils.VerifyPassword(pass, dbPass)
	if err != nil {
		return nil, fmt.Errorf("verify error: %w", err)
	}
	if !ok {
		return nil, ErrBadPass
	}

	return user, nil
}
