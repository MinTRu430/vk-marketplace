package session

import (
	"VK/internal/utils/randutils"
	"context"
	"database/sql"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SessionsDB struct {
	DB *pgxpool.Pool
}

func NewSessionsDB(db *pgxpool.Pool) *SessionsDB {
	return &SessionsDB{
		DB: db,
	}
}

func (sm *SessionsDB) Check(r *http.Request) (*Session, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		log.Println("CheckSession no Authorization header")
		return nil, ErrNoAuth
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		log.Println("CheckSession invalid Authorization header format")
		return nil, ErrNoAuth
	}
	token := parts[1]

	sess := &Session{}
	row := sm.DB.QueryRow(r.Context(), `SELECT user_id FROM sessions WHERE id = $1`, token)
	err := row.Scan(&sess.UserID)
	if err == sql.ErrNoRows {
		log.Println("CheckSession no rows")
		return nil, ErrNoAuth
	} else if err != nil {
		log.Println("CheckSession err:", err)
		return nil, err
	}

	sess.ID = token
	return sess, nil
}

func (sm *SessionsDB) Create(ctx context.Context, w http.ResponseWriter, user UserInterface) error {
	sessID, err := randutils.GenerateSessionID()
	if err != nil {
		log.Println("CreateSession failed to generate session ID:", err)
		return err
	}

	_, err = sm.DB.Exec(ctx, "INSERT INTO sessions(id, user_id) VALUES ($1, $2)", sessID, user.GetID())
	if err != nil {
		log.Println("CreateSession db error:", err)
		return err
	}

	w.Header().Set("Authorization", "Bearer "+sessID)
	w.Header().Set("X-Session-Expires", time.Now().Add(24*time.Hour).Format(time.RFC3339))

	return nil
}

func (sm *SessionsDB) DestroyCurrent(w http.ResponseWriter, r *http.Request) error {
	sess, err := sm.Check(r) // Use Check to get session from Authorization header
	if err != nil {
		log.Println("DestroyCurrent no valid session:", err)
		return err
	}

	_, err = sm.DB.Exec(r.Context(), "DELETE FROM sessions WHERE id = $1", sess.ID)
	if err != nil {
		log.Println("DestroyCurrent db error:", err)
		return err
	}

	w.Header().Set("Authorization", "")

	return nil
}
func (sm *SessionsDB) DestroyAll(ctx context.Context, w http.ResponseWriter, user UserInterface) error {
	_, err := sm.DB.Exec(ctx, "DELETE FROM sessions WHERE user_id = $1", user.GetID())
	if err != nil {
		log.Println("DestroyAll db error:", err)
		return err
	}

	w.Header().Set("Authorization", "")

	return nil
}
