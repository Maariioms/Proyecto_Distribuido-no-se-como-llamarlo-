package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"soc-mvp-distribuido/events/internal/repository"
)

type EventController struct {
	repo repository.EventRepository
}

func NewEventController(repo repository.EventRepository) *EventController {
	return &EventController{repo: repo}
}

type createEventRequest struct {
	Type     string `json:"type"`
	Severity string `json:"severity"`
}

func (c *EventController) Create(w http.ResponseWriter, r *http.Request) {
	var req createEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Type == "" || req.Severity == "" {
		http.Error(w, "type and severity are required", http.StatusBadRequest)
		return
	}

	event, err := c.repo.Create(r.Context(), req.Type, req.Severity)
	if err != nil {
		http.Error(w, "could not create event", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusCreated, event)
}

func (c *EventController) List(w http.ResponseWriter, r *http.Request) {
	events, err := c.repo.List(r.Context())
	if err != nil {
		http.Error(w, "could not list events", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, events)
}

func (c *EventController) Get(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	event, err := c.repo.Get(r.Context(), uint(id))
	if errors.Is(err, repository.ErrNotFound) {
		http.Error(w, "event not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "could not fetch event", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
