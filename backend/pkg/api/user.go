package api

import (
	"coffee-loyalty-system/pkg/service"
	"context"
	"encoding/json"
	"net/http"
)

type UserHandler struct {
	userService *service.UserService
	mux         *http.ServeMux
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	h := &UserHandler{
		userService: userService,
		mux:         http.NewServeMux(),
	}

	h.mux.HandleFunc("/", h.ListUsers) // GET /api/users/

	return h
}

func (h *UserHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	users, err := h.userService.ListUsers(context.Background())
	if err != nil {
		http.Error(w, "Error fetching users", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}
