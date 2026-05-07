package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/dws-1-2026-green/subscriptions/internal/domain/subscription"
	subscriptionusecase "github.com/dws-1-2026-green/subscriptions/internal/usecase/subscription"
	"github.com/go-chi/chi/v5"
)

// Handler handles HTTP requests for subscriptions
type Handler struct {
	service subscriptionusecase.Service
}

// NewHandler creates a new HTTP handler
func NewHandler(service subscriptionusecase.Service) Handler {
	return Handler{service: service}
}

// RegisterRoutes registers the API routes
func (h Handler) RegisterRoutes(r chi.Router) {
	r.Route("/api/v1/subscriptions", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}

// Create handles POST /api/v1/subscriptions
func (h Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req subscription.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.service.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, subscription.ErrSourceRequired) ||
			errors.Is(err, subscription.ErrEventTypeRequired) ||
			errors.Is(err, subscription.ErrDestinationURLRequired) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to create subscription")
		return
	}

	h.respondJSON(w, http.StatusCreated, sub)
}

// GetByID handles GET /api/v1/subscriptions/{id}
func (h Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	sub, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to get subscription")
		return
	}

	h.respondJSON(w, http.StatusOK, sub)
}

// List handles GET /api/v1/subscriptions
func (h Handler) List(w http.ResponseWriter, r *http.Request) {
	source := r.URL.Query().Get("source")
	eventType := r.URL.Query().Get("event_type")

	subs, err := h.service.List(r.Context(), source, eventType)
	if err != nil {
		h.respondError(w, http.StatusInternalServerError, "failed to list subscriptions")
		return
	}

	h.respondJSON(w, http.StatusOK, subs)
}

// Update handles PUT /api/v1/subscriptions/{id}
func (h Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	var req subscription.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	sub, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		if errors.Is(err, subscription.ErrNothingToUpdate) {
			h.respondError(w, http.StatusBadRequest, err.Error())
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to update subscription")
		return
	}

	h.respondJSON(w, http.StatusOK, sub)
}

// Delete handles DELETE /api/v1/subscriptions/{id}
func (h Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		h.respondError(w, http.StatusBadRequest, "id is required")
		return
	}

	err := h.service.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, subscription.ErrNotFound) {
			h.respondError(w, http.StatusNotFound, "subscription not found")
			return
		}
		h.respondError(w, http.StatusInternalServerError, "failed to delete subscription")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h Handler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (h Handler) respondError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}
