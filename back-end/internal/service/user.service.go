package service

import (
	"chat-app/back-end/internal/model"
	"chat-app/back-end/internal/repository"
	"chat-app/back-end/internal/util"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

var ErrInvalidCredentials = errors.New("invalid email or password")

type UserService struct {
	userRepo   *repository.UserRepository
	jwtManager *util.JWTManager
	rdb        *redis.Client
}

func NewUserService(userRepo *repository.UserRepository, jwtManager *util.JWTManager, rdb *redis.Client) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
		rdb:        rdb,
	}
}

func (s *UserService) RegisterAccount(ctx context.Context, data *model.CreateUserRequest) (*model.UserLoginResponse, error) {
	emailExists, err := s.userRepo.IsEmailExists(ctx, data.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, repository.ErrEmailTaken
	}

	usernameExists, err := s.userRepo.IsUsernameExists(ctx, data.Username)
	if err != nil {
		return nil, err
	}
	if usernameExists {
		return nil, repository.ErrUsernameTaken
	}

	hashed, err := util.HashPassword(data.Password)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: data.Username,
		Email:    data.Email,
		Password: hashed,
	}

	if err := s.userRepo.CreateUser(ctx, user); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	if err := util.StoreToken(s.rdb, user.ID.String(), refreshToken, 7*24*time.Hour); err != nil {
		return nil, err
	}

	return &model.UserLoginResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) UserLogin(ctx context.Context, data *model.UserLoginRequest) (*model.UserLoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, data.Email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if !util.CheckPassword(data.Password, user.Password) {
		return nil, ErrInvalidCredentials
	}

	accessToken, refreshToken, err := s.generateTokenPair(user.ID)
	if err != nil {
		return nil, err
	}

	if err := util.StoreToken(s.rdb, user.ID.String(), refreshToken, 7*24*time.Hour); err != nil {
		return nil, err
	}

	return &model.UserLoginResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *UserService) UserLogout(ctx context.Context, userID uuid.UUID) error {
	return util.DeleteToken(s.rdb, userID.String())
}

func (s *UserService) generateTokenPair(userID uuid.UUID) (accessToken, refreshToken string, err error) {
	accessToken, err = s.jwtManager.GenerateAccessToken(userID, util.UserTypeUser, "")
	if err != nil {
		return
	}
	refreshToken, err = s.jwtManager.GenerateRefreshToken(userID, util.UserTypeUser, "")
	return
}
