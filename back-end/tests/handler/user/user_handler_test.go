package user_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chat-app/back-end/internal/handler"
	"chat-app/back-end/internal/model"
	"chat-app/back-end/internal/repository"
	"chat-app/back-end/internal/service"
	"chat-app/back-end/internal/util"

	"github.com/google/uuid"
)

// ---------- mock ----------

type mockUserService struct {
	RegisterAccountFn func(ctx context.Context, data *model.CreateUserRequest) (*model.UserLoginResponse, error)
	UserLoginFn       func(ctx context.Context, data *model.UserLoginRequest) (*model.UserLoginResponse, error)
	UserLogoutFn      func(ctx context.Context, userID uuid.UUID) error
}

func (m *mockUserService) RegisterAccount(ctx context.Context, data *model.CreateUserRequest) (*model.UserLoginResponse, error) {
	return m.RegisterAccountFn(ctx, data)
}
func (m *mockUserService) UserLogin(ctx context.Context, data *model.UserLoginRequest) (*model.UserLoginResponse, error) {
	return m.UserLoginFn(ctx, data)
}
func (m *mockUserService) UserLogout(ctx context.Context, userID uuid.UUID) error {
	return m.UserLogoutFn(ctx, userID)
}

// ---------- helpers ----------

func fakeLoginResp() *model.UserLoginResponse {
	return &model.UserLoginResponse{
		User: &model.UserResponse{
			ID:        uuid.New(),
			Username:  "testuser",
			Email:     "test@example.com",
			CreatedAt: time.Now(),
		},
		AccessToken:  "access.token",
		RefreshToken: "refresh.token",
	}
}

func jsonBody(t *testing.T, v any) *bytes.Reader {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return bytes.NewReader(b)
}

func injectClaims(r *http.Request, userID uuid.UUID) *http.Request {
	claims := &util.Claims{UserID: userID, UserType: util.UserTypeUser, TokenType: util.AccessToken}
	return r.WithContext(context.WithValue(r.Context(), "claims", claims))
}

// ---------- Register ----------

func TestRegister_Success(t *testing.T) {
	svc := &mockUserService{
		RegisterAccountFn: func(_ context.Context, _ *model.CreateUserRequest) (*model.UserLoginResponse, error) {
			return fakeLoginResp(), nil
		},
	}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		jsonBody(t, model.CreateUserRequest{Username: "testuser", Email: "test@example.com", Password: "password123"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
	var hasCookie bool
	for _, c := range w.Result().Cookies() {
		if c.Name == "refresh_token" {
			hasCookie = true
		}
	}
	if !hasCookie {
		t.Error("expected refresh_token cookie to be set")
	}
}

func TestRegister_EmailTaken(t *testing.T) {
	svc := &mockUserService{
		RegisterAccountFn: func(_ context.Context, _ *model.CreateUserRequest) (*model.UserLoginResponse, error) {
			return nil, repository.ErrEmailTaken
		},
	}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		jsonBody(t, model.CreateUserRequest{Username: "testuser", Email: "test@example.com", Password: "password123"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestRegister_UsernameTaken(t *testing.T) {
	svc := &mockUserService{
		RegisterAccountFn: func(_ context.Context, _ *model.CreateUserRequest) (*model.UserLoginResponse, error) {
			return nil, repository.ErrUsernameTaken
		},
	}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		jsonBody(t, model.CreateUserRequest{Username: "testuser", Email: "test@example.com", Password: "password123"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestRegister_ValidationFail(t *testing.T) {
	svc := &mockUserService{}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		jsonBody(t, map[string]string{"username": "x"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestRegister_BadJSON(t *testing.T) {
	svc := &mockUserService{}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Register(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------- Login ----------

func TestLogin_Success(t *testing.T) {
	svc := &mockUserService{
		UserLoginFn: func(_ context.Context, _ *model.UserLoginRequest) (*model.UserLoginResponse, error) {
			return fakeLoginResp(), nil
		},
	}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		jsonBody(t, model.UserLoginRequest{Email: "test@example.com", Password: "password123"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

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
		t.Error("expected refresh_token cookie to be set")
	}
}

func TestLogin_InvalidCredentials(t *testing.T) {
	svc := &mockUserService{
		UserLoginFn: func(_ context.Context, _ *model.UserLoginRequest) (*model.UserLoginResponse, error) {
			return nil, service.ErrInvalidCredentials
		},
	}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		jsonBody(t, model.UserLoginRequest{Email: "test@example.com", Password: "wrong"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestLogin_ValidationFail(t *testing.T) {
	svc := &mockUserService{}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login",
		jsonBody(t, map[string]string{"email": "not-an-email"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestLogin_BadJSON(t *testing.T) {
	svc := &mockUserService{}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader([]byte("{bad")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.Login(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ---------- Logout ----------

func TestLogout_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockUserService{
		UserLogoutFn: func(_ context.Context, id uuid.UUID) error {
			if id != userID {
				t.Errorf("unexpected userID: %v", id)
			}
			return nil
		},
	}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestLogout_NoClaims(t *testing.T) {
	svc := &mockUserService{}
	h := handler.NewUserHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	w := httptest.NewRecorder()

	h.Logout(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
