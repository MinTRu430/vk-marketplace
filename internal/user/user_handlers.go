package user

import (
	"VK/internal/session"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"
)

type UserHandler struct {
	SessionsDB session.SessionManager
	UserDB     *UserDB
}

var (
	loginRE = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	passRE  = regexp.MustCompile(`^[\x21-\x7E]{8,64}$`)
)

type regAndAuthRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type registrationResponse struct {
	ID    uint32 `json:"id"`
	Login string `json:"login"`
}

func (uh *UserHandler) Registration(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req regAndAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if !loginRE.MatchString(req.Login) {
		http.Error(w, "Invalid login format", http.StatusBadRequest)
		return
	}
	if !passRE.MatchString(req.Password) {
		http.Error(w, "Invalid password format", http.StatusBadRequest)
		return
	}

	user, err := uh.UserDB.Create(r.Context(), req.Login, req.Password)
	if err != nil {
		switch err {
		case ErrUserExists:
			http.Error(w, "User already exists", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	resp := registrationResponse{
		ID:    user.ID,
		Login: user.Login,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (uh *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req regAndAuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Login == "" || req.Password == "" {
		http.Error(w, "Login and password required", http.StatusBadRequest)
		return
	}

	user, err := uh.UserDB.CheckPasswordByLogin(r.Context(), req.Login, req.Password)
	if err != nil {
		http.Error(w, "User Not Found", http.StatusUnauthorized)
		return
	}

	sessID, err := uh.SessionsDB.Create(r.Context(), w, user)
	if err != nil {
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}
	fmt.Println(sessID)

	w.Header().Set("Authorization", "Bearer "+sessID)
	w.Header().Set("X-Session-Expires", time.Now().Add(time.Duration(time.Now().Year())).Format(time.RFC3339))
	fmt.Println("Response headers:")
	for k, v := range w.Header() {
		fmt.Println(k, v)
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"login": user.Login,
	})
}
