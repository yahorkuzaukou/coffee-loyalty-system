package cmd

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type mockDB struct{}

func (m *mockDB) Pool() *pgxpool.Pool {
	return nil
}

func (m *mockDB) Close() {}

func TestNewServer(t *testing.T) {
	server, err := NewServer(&mockDB{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	if server == nil {
		t.Fatal("Expected server to be created, got nil")
	}

	if server.Addr != ":8080" {
		t.Errorf("Expected server address to be :8080, got %s", server.Addr)
	}

	if server.Handler == nil {
		t.Error("Expected server handler to be set, got nil")
	}
}

func TestHealthCheck(t *testing.T) {
	server, err := NewServer(&mockDB{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	server.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}

	if w.Body.String() != "OK" {
		t.Errorf("Expected body 'OK', got '%s'", w.Body.String())
	}
}

func TestServerShutdown(t *testing.T) {
	server, err := NewServer(&mockDB{})
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	go func() {
		server.ListenAndServe()
	}()

	err = server.Shutdown(ctx)
	if err != nil {
		t.Errorf("Expected no error during shutdown, got %v", err)
	}
}

func TestLoggingMiddleware(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler := loggingMiddleware(testHandler)
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Code)
	}
}
