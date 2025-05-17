package router

import (
	"net/http"

	"coffee-loyalty-system/pkg/api"
	"coffee-loyalty-system/pkg/service"
)

type Router struct {
	mux *http.ServeMux
}

func NewRouter(userService *service.UserService) *Router {
	r := &Router{
		mux: http.NewServeMux(),
	}

	// Health check route
	r.mux.HandleFunc("/health", api.HealthCheck)

	// User routes with prefix
	userHandler := api.NewUserHandler(userService)
	r.mux.Handle("/api/users/", http.StripPrefix("/api/users", userHandler))

	return r
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.mux.ServeHTTP(w, req)
}
