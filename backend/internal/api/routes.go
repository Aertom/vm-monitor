package api

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/Aertom/vm-monitor/backend/internal/store"
)

type API struct {
	Store *store.Store
}

func (a *API) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/api/groups", a.handleGetGroups)
	r.Post("/api/groups/{groupID}/checkout", a.handleCheckout)
	r.Post("/api/groups/{groupID}/checkin", a.handleCheckin)

	return r
}

func (a *API) handleGetGroups(w http.ResponseWriter, r *http.Request) {
	groups := a.Store.GetGroups()
	writeJSON(w, http.StatusOK, groups)
}

type checkoutRequest struct {
	User string `json:"user"`
}

func (a *API) handleCheckout(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")

	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.User == "" {
		http.Error(w, "invalid request: 'user' is required", http.StatusBadRequest)
		return
	}

	err := a.Store.Checkout(groupID, req.User)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, store.ErrAlreadyCheckedOut):
		http.Error(w, err.Error(), http.StatusConflict)
	case errors.Is(err, store.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (a *API) handleCheckin(w http.ResponseWriter, r *http.Request) {
	groupID := chi.URLParam(r, "groupID")

	err := a.Store.Checkin(groupID)
	switch {
	case err == nil:
		w.WriteHeader(http.StatusNoContent)
	case errors.Is(err, store.ErrGroupNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
