package session

import (
	"context"
	"errors"
	"net/http"
)

var (
	ErrNoAuth = errors.New("no session found")
)

type Session struct {
	UserID  uint32
	SquadID uint32
	ID      string
}

type UserInterface interface { // Потом подумать о его существовании, может он не нужен будет
	GetID() uint32
}

type SessionManager interface {
	Check(*http.Request) (*Session, error)
	Create(context.Context, http.ResponseWriter, UserInterface) error
	DestroyCurrent(http.ResponseWriter, *http.Request) error
	DestroyAll(context.Context, http.ResponseWriter, UserInterface) error
}
