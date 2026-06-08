//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestIntegration_Register_Login_Flow(t *testing.T) {
	env := setupTestEnv(t)
	username := fmt.Sprintf("inttest_%d", time.Now().UnixNano())

	t.Run("register new user → 201 with token", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "integration_pass123",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["token"] == nil || resp["token"] == "" {
			t.Error("expected a non-empty JWT token in response")
		}
	})

	t.Run("register same username again → 409", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "integration_pass123",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("login with correct credentials → 200 with token", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "integration_pass123",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp map[string]interface{}
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["token"] == nil || resp["token"] == "" {
			t.Error("expected a non-empty JWT token on login")
		}
	})

	t.Run("login with wrong password → 401", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{
			"username": username,
			"password": "wrongpassword",
		})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("register with invalid body → 400", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"username": "ab"})
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		env.router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
		}
	})
}

func TestIntegration_ContextCancellation_DoesNotPanic(t *testing.T) {
	env := setupTestEnv(t)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	body, _ := json.Marshal(map[string]string{
		"username": "ctx_user",
		"password": "does_not_matter",
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handler panicked with cancelled context: %v", r)
		}
	}()

	env.router.ServeHTTP(w, req)
}

func TestIntegration_RequestTimeout_DoesNotPanic(t *testing.T) {
	env := setupTestEnv(t)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	body, _ := json.Marshal(map[string]string{
		"username": "timeout_user",
		"password": "does_not_matter",
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, "/api/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handler panicked with timed-out context: %v", r)
		}
	}()

	env.router.ServeHTTP(w, req)
}
