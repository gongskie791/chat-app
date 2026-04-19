package hub

import (
	"chat-app/back-end/internal/model"
	"chat-app/back-end/internal/repository"
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	UserID   uuid.UUID
	Username string
	RoomID   uuid.UUID
	MsgRepo  *repository.MessageRepository
}

// ReadPump pumps messages from the WebSocket to the hub.
// Runs in its own goroutine per client.
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, raw, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		var evt model.IncomingEvent
		if err := json.Unmarshal(raw, &evt); err != nil {
			continue
		}

		switch evt.Type {
		case model.EventMessage:
			c.handleMessage(evt.Content)
		case model.EventTyping:
			c.handleTyping()
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket.
// Runs in its own goroutine per client.
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

func (c *Client) handleMessage(content string) {
	if content == "" {
		return
	}

	msg := &model.ChatMessage{
		ID:        uuid.New().String(),
		RoomID:    c.RoomID.String(),
		UserID:    c.UserID.String(),
		Username:  c.Username,
		Content:   content,
		Timestamp: time.Now().UnixMilli(),
	}

	// persist to Redis — fire and forget, don't block the pump
	go c.MsgRepo.SaveMessage(context.Background(), msg)

	data, _ := json.Marshal(model.OutgoingEvent{
		Type:      model.EventMessage,
		Content:   msg,
		UserID:    c.UserID.String(),
		Username:  c.Username,
		Timestamp: msg.Timestamp,
	})

	c.Hub.Broadcast(&BroadcastMsg{RoomID: c.RoomID, Data: data})
}

func (c *Client) handleTyping() {
	data, _ := json.Marshal(model.OutgoingEvent{
		Type:     model.EventTyping,
		UserID:   c.UserID.String(),
		Username: c.Username,
	})

	// Except: c — don't send the typing event back to the typer
	c.Hub.Broadcast(&BroadcastMsg{RoomID: c.RoomID, Data: data, Except: c})
}
