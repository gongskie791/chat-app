package handler

import (
	"chat-app/back-end/internal/hub"
	"chat-app/back-end/internal/model"
	"chat-app/back-end/internal/repository"
	"chat-app/back-end/internal/util"
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Origin check is handled by the CORS middleware
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub      *hub.Hub
	userRepo *repository.UserRepository
	msgRepo  *repository.MessageRepository
}

func NewWSHandler(h *hub.Hub, userRepo *repository.UserRepository, msgRepo *repository.MessageRepository) *WSHandler {
	return &WSHandler{hub: h, userRepo: userRepo, msgRepo: msgRepo}
}

// GET /api/ws/{roomID}
// Token is passed as a query param: ?token=<access_token>
// (VerifyJWT middleware already supports this)
func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	claims := util.ClaimsFromContext(r)
	if claims == nil {
		util.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	roomID, err := uuid.Parse(r.PathValue("roomID"))
	if err != nil {
		util.Error(w, http.StatusBadRequest, "Invalid room ID")
		return
	}

	user, err := h.userRepo.GetUserByID(r.Context(), claims.UserID)
	if err != nil {
		util.Error(w, http.StatusNotFound, "User not found")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// upgrader already wrote the error response
		return
	}

	client := &hub.Client{
		Hub:      h.hub,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		UserID:   claims.UserID,
		Username: user.Username,
		RoomID:   roomID,
		MsgRepo:  h.msgRepo,
	}

	// Push history before registering so it arrives first in WritePump
	history, err := h.msgRepo.GetRecentMessages(r.Context(), roomID.String(), 50)
	if err == nil && len(history) > 0 {
		evt, _ := json.Marshal(model.OutgoingEvent{
			Type:    model.EventHistory,
			Content: history,
		})
		client.Send <- evt
	}

	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
