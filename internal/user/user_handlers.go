package user

import (
	"VK/internal/session"
	"encoding/json"
	"log"
	"net/http"
	"regexp"
)

type UserHandler struct {
	// Tmpl      Templater
	Sessions session.SessionManager
	UserDB   *UserDB
}

var (
	loginRE = regexp.MustCompile(`^[a-zA-Z0-9_-]{3,32}$`)
	passRE  = regexp.MustCompile(`^[a-zA-Z0-9_-]{8,32}$`) //Возможно стоит переделать, оба валидатора
)

type registrationRequest struct {
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

	var req registrationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	login := req.Login
	pass := req.Password

	if !loginRE.MatchString(login) {
		http.Error(w, "Invalid login format", http.StatusBadRequest)
		return
	}
	if !passRE.MatchString(pass) {
		http.Error(w, "Invalid password format", http.StatusBadRequest)
		return
	}

	user, err := uh.UserDB.Create(r.Context(), login, pass)
	if err != nil {
		switch err {
		case ErrUserExists:
			http.Error(w, "User already exists", http.StatusConflict)
		default:
			log.Println("DB error during registration:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if err := uh.Sessions.Create(r.Context(), w, user); err != nil {
		log.Println("Failed to create session:", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
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

	var req registrationRequest
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
		log.Println("auth error:", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := uh.Sessions.Create(r.Context(), w, user); err != nil {
		log.Println("session creation error:", err)
		http.Error(w, "Session error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":    user.ID,
		"login": user.Login,
	})
}

func (uh *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := uh.Sessions.DestroyCurrent(w, r)
	if err != nil {
		log.Println("Logout error:", err)
		http.Error(w, "Unauthorized or session not found", http.StatusUnauthorized)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}
