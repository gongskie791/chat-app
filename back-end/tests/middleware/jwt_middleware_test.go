package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chat-app/back-end/internal/middleware"
	"chat-app/back-end/internal/util"

	"github.com/google/uuid"
)

func newJWT() *util.JWTManager {
	return util.NewJWTManager("test-secret", 15*time.Minute, 7*24*time.Hour)
}

func okHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

func TestVerifyJWT_MissingToken(t *testing.T) {
	mgr := newJWT()
	var called bool
	h := middleware.VerifyJWT(mgr)(okHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if called {
		t.Error("next handler should not be called")
	}
}

func TestVerifyJWT_InvalidToken(t *testing.T) {
	mgr := newJWT()
	var called bool
	h := middleware.VerifyJWT(mgr)(okHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer not.a.real.token")
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if called {
		t.Error("next handler should not be called")
	}
}

func TestVerifyJWT_ExpiredToken(t *testing.T) {
	mgr := util.NewJWTManager("test-secret", -1*time.Second, 7*24*time.Hour)
	token, err := mgr.GenerateAccessToken(uuid.New(), util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var called bool
	h := middleware.VerifyJWT(mgr)(okHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if called {
		t.Error("next handler should not be called")
	}
}

func TestVerifyJWT_RefreshTokenUsedAsAccess(t *testing.T) {
	mgr := newJWT()
	token, err := mgr.GenerateRefreshToken(uuid.New(), util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var called bool
	h := middleware.VerifyJWT(mgr)(okHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
	if called {
		t.Error("next handler should not be called")
	}
}

func TestVerifyJWT_ValidBearerHeader(t *testing.T) {
	mgr := newJWT()
	token, err := mgr.GenerateAccessToken(uuid.New(), util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var called bool
	h := middleware.VerifyJWT(mgr)(okHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Error("next handler should be called")
	}
}

func TestVerifyJWT_ValidQueryParam(t *testing.T) {
	mgr := newJWT()
	token, err := mgr.GenerateAccessToken(uuid.New(), util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var called bool
	h := middleware.VerifyJWT(mgr)(okHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/?token="+token, nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Error("next handler should be called")
	}
}

func TestVerifyJWT_ClaimsInjectedIntoContext(t *testing.T) {
	mgr := newJWT()
	userID := uuid.New()
	token, err := mgr.GenerateAccessToken(userID, util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	var gotClaims *util.Claims
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotClaims = util.ClaimsFromContext(r)
		w.WriteHeader(http.StatusOK)
	})
	h := middleware.VerifyJWT(mgr)(next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	if gotClaims == nil {
		t.Fatal("expected claims in context, got nil")
	}
	if gotClaims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, gotClaims.UserID)
	}
}
