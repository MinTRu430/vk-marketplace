package session

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrNoAuth = errors.New("No session found")
)

type Session struct {
	UserID uint32
	ID     string
}

type UserInterface interface {
	GetID() uint32
}

type SessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(context.Context, http.ResponseWriter, UserInterface) error
	DestroyCurrent(http.ResponseWriter, *http.Request) error
	DestroyAll(context.Context, http.ResponseWriter, UserInterface) error
}

func SessionFromContext(ctx context.Context) (*Session, error) {
	sess, ok := ctx.Value(1).(*Session)
	if !ok {
		return nil, ErrNoAuth
	}
	return sess, nil
}
