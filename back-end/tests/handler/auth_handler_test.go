package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"chat-app/back-end/internal/handler"
	"chat-app/back-end/internal/service"
)

// ---------- mock ----------

type mockAuthService struct {
	RefreshTokenFn func(ctx context.Context, token string) (*service.RefreshTokenResponse, error)
}

func (m *mockAuthService) RefreshToken(ctx context.Context, token string) (*service.RefreshTokenResponse, error) {
	return m.RefreshTokenFn(ctx, token)
}

// ---------- RefreshToken ----------

func TestRefreshToken_Success(t *testing.T) {
	svc := &mockAuthService{
		RefreshTokenFn: func(_ context.Context, _ string) (*service.RefreshTokenResponse, error) {
			return &service.RefreshTokenResponse{
				AccessToken:  "new.access.token",
				RefreshToken: "new.refresh.token",
			}, nil
		},
	}
	h := handler.NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old.refresh.token"})
	w := httptest.NewRecorder()

	h.RefreshToken(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var hasCookie bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "refresh_token" {
			hasCookie = true
		}
	}
	if !hasCookie {
		t.Error("expected refresh_token cookie to be refreshed")
	}
}

func TestRefreshToken_NoCookie(t *testing.T) {
	svc := &mockAuthService{}
	h := handler.NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	w := httptest.NewRecorder()

	h.RefreshToken(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestRefreshToken_InvalidToken(t *testing.T) {
	svc := &mockAuthService{
		RefreshTokenFn: func(_ context.Context, _ string) (*service.RefreshTokenResponse, error) {
			return nil, service.ErrRefreshTokenInvalid
		},
	}
	h := handler.NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bad.token"})
	w := httptest.NewRecorder()

	h.RefreshToken(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	var cleared bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "refresh_token" && c.MaxAge < 0 {
			cleared = true
		}
	}
	if !cleared {
		t.Error("expected refresh_token cookie to be cleared on invalid token")
	}
}

func TestRefreshToken_InternalError(t *testing.T) {
	svc := &mockAuthService{
		RefreshTokenFn: func(_ context.Context, _ string) (*service.RefreshTokenResponse, error) {
			return nil, errors.New("db error")
		},
	}
	h := handler.NewAuthHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "some.token"})
	w := httptest.NewRecorder()

	h.RefreshToken(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
