package model

type EventType string

const (
	EventMessage     EventType = "message"
	EventTyping      EventType = "typing"
	EventJoin        EventType = "join"
	EventLeave       EventType = "leave"
	EventOnlineUsers EventType = "online_users"
	EventHistory     EventType = "history"
)

// IncomingEvent is what the client sends over the WebSocket.
type IncomingEvent struct {
	Type    EventType `json:"type"`
	Content string    `json:"content"` // used for message text
}

// OutgoingEvent is what the server broadcasts to clients.
type OutgoingEvent struct {
	Type      EventType `json:"type"`
	Content   any       `json:"content,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Timestamp int64     `json:"timestamp,omitempty"`
}

// ChatMessage is stored in Redis and sent as OutgoingEvent.Content.
type ChatMessage struct {
	ID        string `json:"id"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Content   string `json:"content"`
	Timestamp int64  `json:"timestamp"`
}

// OnlineUser is an entry in the online_users event content list.
type OnlineUser struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}
