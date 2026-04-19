package service

import (
	"chat-app/back-end/internal/model"
	"chat-app/back-end/internal/repository"
	"context"

	"github.com/google/uuid"
)

type RoomService struct {
	roomRepo *repository.RoomRepository
}

func NewRoomService(roomRepo *repository.RoomRepository) *RoomService {
	return &RoomService{roomRepo: roomRepo}
}

func (s *RoomService) CreateRoom(ctx context.Context, createdBy uuid.UUID, req *model.CreateRoomRequest) (*model.Room, error) {
	taken, err := s.roomRepo.IsRoomNameTaken(ctx, req.Name)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, repository.ErrRoomNameTaken
	}

	room := &model.Room{
		Name:      req.Name,
		CreatedBy: createdBy,
	}
	if err := s.roomRepo.CreateRoom(ctx, room); err != nil {
		return nil, err
	}
	return room, nil
}

func (s *RoomService) GetRooms(ctx context.Context) ([]*model.Room, error) {
	return s.roomRepo.GetRooms(ctx)
}
