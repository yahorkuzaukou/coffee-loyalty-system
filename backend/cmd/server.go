package cmd

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"coffee-loyalty-system/pkg/router"
	"coffee-loyalty-system/pkg/service"
	"coffee-loyalty-system/pkg/storage"

	"github.com/sirupsen/logrus"
)

var log = logrus.New()

type Server struct {
	*http.Server
	db storage.Database
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.RequestURI, time.Since(start))
	})
}

func NewServer(dbOptional ...storage.Database) (*Server, error) {
	var db storage.Database
	if len(dbOptional) > 0 && dbOptional[0] != nil {
		db = dbOptional[0]
	} else {
		cfg := storage.Config{
			Host:     getEnvOrDefault("DB_HOST", "localhost"),
			Port:     getEnvIntOrDefault("DB_PORT", 5432),
			User:     getEnvOrDefault("DB_USER", "postgres"),
			Password: getEnvOrDefault("DB_PASSWORD", "postgres"),
			DBName:   getEnvOrDefault("DB_NAME", "coffee_loyalty"),
			SSLMode:  getEnvOrDefault("DB_SSLMODE", "disable"),
		}
		var err error
		db, err = storage.NewDatabase(cfg)
		if err != nil {
			return nil, err
		}
	}

	userService := service.NewUserService(db)
	r := router.NewRouter(userService)

	server := &Server{
		Server: &http.Server{
			Addr:         ":8080",
			Handler:      loggingMiddleware(r),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		db: db,
	}

	return server, nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.Server.Shutdown(ctx); err != nil {
		return err
	}
	s.db.Close()
	return nil
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func Run(server *Server) error {
	serverErrors := make(chan error, 1)

	go func() {
		log.Info("Server listening on :8080")
		serverErrors <- server.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.WithError(err).Error("Server error")
		return err

	case sig := <-shutdown:
		log.WithField("signal", sig).Info("Shutdown signal received")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(ctx); err != nil {
			log.WithError(err).Error("Could not stop server gracefully")
			server.Close()
			return err
		}
	}

	return nil
}
