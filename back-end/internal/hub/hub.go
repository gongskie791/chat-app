package hub

import (
	"chat-app/back-end/internal/model"
	"encoding/json"
	"sync"

	"github.com/google/uuid"
)

// BroadcastMsg carries a payload destined for a room.
// Set Except to skip the sender (e.g. typing indicators).
type BroadcastMsg struct {
	RoomID uuid.UUID
	Data   []byte
	Except *Client
}

type Hub struct {
	// rooms[roomID] → set of connected clients
	rooms      map[uuid.UUID]map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMsg
	mu         sync.RWMutex
}

func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[uuid.UUID]map[*Client]bool),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan *BroadcastMsg, 256),
	}
}

// Run is the hub's main event loop. Call it in a goroutine.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.mu.Lock()
			if h.rooms[c.RoomID] == nil {
				h.rooms[c.RoomID] = make(map[*Client]bool)
			}
			h.rooms[c.RoomID][c] = true
			h.mu.Unlock()
			h.notifyPresence(c.RoomID, c, model.EventJoin)

		case c := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[c.RoomID]; ok {
				if _, ok := clients[c]; ok {
					delete(clients, c)
					close(c.Send)
					if len(clients) == 0 {
						delete(h.rooms, c.RoomID)
					}
				}
			}
			h.mu.Unlock()
			// client is already removed, so remaining clients get updated list
			h.notifyPresence(c.RoomID, c, model.EventLeave)

		case msg := <-h.broadcast:
			h.sendToRoom(msg.RoomID, msg.Data, msg.Except)
		}
	}
}

func (h *Hub) Register(c *Client) {
	h.register <- c
}

func (h *Hub) Unregister(c *Client) {
	h.unregister <- c
}

func (h *Hub) Broadcast(msg *BroadcastMsg) {
	h.broadcast <- msg
}

// sendToRoom pushes data to every client in the room except one.
func (h *Hub) sendToRoom(roomID uuid.UUID, data []byte, except *Client) {
	h.mu.RLock()
	clients := h.rooms[roomID]
	h.mu.RUnlock()

	for c := range clients {
		if c == except {
			continue
		}
		select {
		case c.Send <- data:
		default:
			// client buffer full — drop the message for this client
		}
	}
}

// notifyPresence sends a join/leave announcement and a fresh online-users list
// to everyone currently in the room.
func (h *Hub) notifyPresence(roomID uuid.UUID, actor *Client, eventType model.EventType) {
	h.mu.RLock()
	clients := h.rooms[roomID]
	h.mu.RUnlock()

	// build current online list (actor is already removed on leave)
	online := make([]model.OnlineUser, 0, len(clients))
	for c := range clients {
		online = append(online, model.OnlineUser{
			UserID:   c.UserID.String(),
			Username: c.Username,
		})
	}

	onlineEvt, _ := json.Marshal(model.OutgoingEvent{
		Type:    model.EventOnlineUsers,
		Content: online,
	})

	announceEvt, _ := json.Marshal(model.OutgoingEvent{
		Type:     eventType,
		UserID:   actor.UserID.String(),
		Username: actor.Username,
	})

	for c := range clients {
		select {
		case c.Send <- onlineEvt:
		default:
		}
		select {
		case c.Send <- announceEvt:
		default:
		}
	}
}
