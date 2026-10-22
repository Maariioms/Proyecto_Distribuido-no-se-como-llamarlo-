package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"soc-mvp-distribuido/backend/internal/events"
	"soc-mvp-distribuido/backend/internal/repository"
)

type AccountController struct {
	repo repository.AccountRepository
	// como avisamos a events que paso algo, ej un login o un registro
	// es una interfaz igual que repo y el controller no sabe si es grpc o kafka
	publisher events.Publisher
}

func NewAccountController(repo repository.AccountRepository, publisher events.Publisher) *AccountController {
	return &AccountController{repo: repo, publisher: publisher}
}

type credentialsRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (c *AccountController) Register(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Username == "" || len(req.Password) < 6 {
		http.Error(w, "username is required and password must be at least 6 characters", http.StatusBadRequest)
		return
	}

	account, err := c.repo.Create(r.Context(), req.Username, req.Password)
	if errors.Is(err, repository.ErrUsernameTaken) {
		http.Error(w, "username already taken", http.StatusConflict)
		return
	}
	if err != nil {
		http.Error(w, "could not create account", http.StatusInternalServerError)
		return
	}
	c.publisher.Publish(r.Context(), "account_created", "low")
	writeJSON(w, http.StatusCreated, account)
}

func (c *AccountController) Login(w http.ResponseWriter, r *http.Request) {
	var req credentialsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	account, err := c.repo.Authenticate(r.Context(), req.Username, req.Password)
	if errors.Is(err, repository.ErrInvalidCredentials) {
		c.publisher.Publish(r.Context(), "login_failed", "medium")
		http.Error(w, "invalid username or password", http.StatusUnauthorized)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	c.publisher.Publish(r.Context(), "login_success", "low")
	writeJSON(w, http.StatusOK, account)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
