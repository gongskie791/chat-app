package handler

import (
	"chat-app/back-end/internal/model"
	"chat-app/back-end/internal/repository"
	"chat-app/back-end/internal/service"
	"chat-app/back-end/internal/util"
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/google/uuid"
)

// RoomServiceI is the interface the RoomHandler depends on.
type RoomServiceI interface {
	CreateRoom(ctx context.Context, createdBy uuid.UUID, req *model.CreateRoomRequest) (*model.Room, error)
	GetRooms(ctx context.Context) ([]*model.Room, error)
}

// compile-time check
var _ RoomServiceI = (*service.RoomService)(nil)

type RoomHandler struct {
	roomService RoomServiceI
}

func NewRoomHandler(roomService RoomServiceI) *RoomHandler {
	return &RoomHandler{roomService: roomService}
}

// POST /api/rooms
func (h *RoomHandler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	claims := util.ClaimsFromContext(r)
	if claims == nil {
		util.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	var req model.CreateRoomRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if fields := util.ValidateStruct(&req); fields != nil {
		util.ValidationFailed(w, fields)
		return
	}

	room, err := h.roomService.CreateRoom(r.Context(), claims.UserID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrRoomNameTaken) {
			util.Error(w, http.StatusConflict, "Room name already taken")
			return
		}
		util.Error(w, http.StatusInternalServerError, "Failed to create room")
		return
	}

	util.Success(w, http.StatusCreated, "Room created", room)
}

// GET /api/rooms
func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.roomService.GetRooms(r.Context())
	if err != nil {
		util.Error(w, http.StatusInternalServerError, "Failed to get rooms")
		return
	}

	util.Success(w, http.StatusOK, "Rooms fetched", rooms)
}
