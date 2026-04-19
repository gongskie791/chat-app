package repository

import (
	"chat-app/back-end/internal/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

var ErrRoomNotFound  = errors.New("room not found")
var ErrRoomNameTaken = errors.New("room name already taken")

type RoomRepository struct {
	db *sqlx.DB
}

func NewRoomRepository(db *sqlx.DB) *RoomRepository {
	return &RoomRepository{db: db}
}

func (r *RoomRepository) CreateRoom(ctx context.Context, room *model.Room) error {
	query := `
		INSERT INTO rooms (name, created_by)
		VALUES ($1, $2)
		RETURNING id, created_at`

	return r.db.QueryRowContext(ctx, query, room.Name, room.CreatedBy).
		Scan(&room.ID, &room.CreatedAt)
}

func (r *RoomRepository) GetRooms(ctx context.Context) ([]*model.Room, error) {
	var rooms []*model.Room
	err := r.db.SelectContext(ctx, &rooms,
		`SELECT id, name, created_by, created_at FROM rooms ORDER BY created_at DESC`)
	return rooms, err
}

func (r *RoomRepository) GetRoomByID(ctx context.Context, id uuid.UUID) (*model.Room, error) {
	var room model.Room
	err := r.db.GetContext(ctx, &room,
		`SELECT id, name, created_by, created_at FROM rooms WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrRoomNotFound
	}
	return &room, err
}

func (r *RoomRepository) IsRoomNameTaken(ctx context.Context, name string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM rooms WHERE name = $1)`, name).Scan(&exists)
	return exists, err
}
