package handler_test

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"chat-app/back-end/internal/handler"
	"chat-app/back-end/internal/model"
	"chat-app/back-end/internal/repository"

	"github.com/google/uuid"
)

// ---------- mock ----------

type mockRoomService struct {
	CreateRoomFn func(ctx context.Context, createdBy uuid.UUID, req *model.CreateRoomRequest) (*model.Room, error)
	GetRoomsFn   func(ctx context.Context) ([]*model.Room, error)
}

func (m *mockRoomService) CreateRoom(ctx context.Context, createdBy uuid.UUID, req *model.CreateRoomRequest) (*model.Room, error) {
	return m.CreateRoomFn(ctx, createdBy, req)
}
func (m *mockRoomService) GetRooms(ctx context.Context) ([]*model.Room, error) {
	return m.GetRoomsFn(ctx)
}

func fakeRoom(name string, createdBy uuid.UUID) *model.Room {
	return &model.Room{
		ID:        uuid.New(),
		Name:      name,
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}
}

// ---------- CreateRoom ----------

func TestCreateRoom_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockRoomService{
		CreateRoomFn: func(_ context.Context, by uuid.UUID, req *model.CreateRoomRequest) (*model.Room, error) {
			return fakeRoom(req.Name, by), nil
		},
	}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms",
		jsonBody(t, model.CreateRoomRequest{Name: "general"}))
	req.Header.Set("Content-Type", "application/json")
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.CreateRoom(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateRoom_NoClaims(t *testing.T) {
	svc := &mockRoomService{}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms",
		jsonBody(t, model.CreateRoomRequest{Name: "general"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.CreateRoom(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestCreateRoom_NameTaken(t *testing.T) {
	userID := uuid.New()
	svc := &mockRoomService{
		CreateRoomFn: func(_ context.Context, _ uuid.UUID, _ *model.CreateRoomRequest) (*model.Room, error) {
			return nil, repository.ErrRoomNameTaken
		},
	}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms",
		jsonBody(t, model.CreateRoomRequest{Name: "general"}))
	req.Header.Set("Content-Type", "application/json")
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.CreateRoom(w, req)

	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

func TestCreateRoom_ValidationFail(t *testing.T) {
	userID := uuid.New()
	svc := &mockRoomService{}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms",
		bytes.NewReader([]byte(`{}`))) // name is required
	req.Header.Set("Content-Type", "application/json")
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.CreateRoom(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCreateRoom_BadJSON(t *testing.T) {
	userID := uuid.New()
	svc := &mockRoomService{}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms", bytes.NewReader([]byte("bad")))
	req.Header.Set("Content-Type", "application/json")
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.CreateRoom(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestCreateRoom_InternalError(t *testing.T) {
	userID := uuid.New()
	svc := &mockRoomService{
		CreateRoomFn: func(_ context.Context, _ uuid.UUID, _ *model.CreateRoomRequest) (*model.Room, error) {
			return nil, errors.New("db error")
		},
	}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodPost, "/api/rooms",
		jsonBody(t, model.CreateRoomRequest{Name: "general"}))
	req.Header.Set("Content-Type", "application/json")
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.CreateRoom(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// ---------- GetRooms ----------

func TestGetRooms_Success(t *testing.T) {
	userID := uuid.New()
	svc := &mockRoomService{
		GetRoomsFn: func(_ context.Context) ([]*model.Room, error) {
			return []*model.Room{fakeRoom("general", userID)}, nil
		},
	}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.GetRooms(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGetRooms_InternalError(t *testing.T) {
	userID := uuid.New()
	svc := &mockRoomService{
		GetRoomsFn: func(_ context.Context) ([]*model.Room, error) {
			return nil, errors.New("db error")
		},
	}
	h := handler.NewRoomHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/rooms", nil)
	req = injectClaims(req, userID)
	w := httptest.NewRecorder()

	h.GetRooms(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
