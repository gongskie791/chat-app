package service

import (
	"chat-app/back-end/internal/util"
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrRefreshTokenInvalid = errors.New("invalid or expired refresh token")

type RefreshTokenResponse struct {
	AccessToken  string
	RefreshToken string
}

type AuthService struct {
	rdb        *redis.Client
	jwtManager *util.JWTManager
}

func NewAuthService(rdb *redis.Client, jwtManager *util.JWTManager) *AuthService {
	return &AuthService{rdb: rdb, jwtManager: jwtManager}
}

func (s *AuthService) RefreshToken(ctx context.Context, incomingRefreshToken string) (*RefreshTokenResponse, error) {
	claims, err := s.jwtManager.ValidateToken(incomingRefreshToken)
	if err != nil {
		return nil, ErrRefreshTokenInvalid
	}

	if claims.TokenType != util.RefreshToken {
		return nil, ErrRefreshTokenInvalid
	}

	userID := claims.UserID.String()
	stored, err := util.GetToken(s.rdb, userID)
	if err != nil || stored == "" || stored != incomingRefreshToken {
		return nil, ErrRefreshTokenInvalid
	}

	newAccess, err := s.jwtManager.GenerateAccessToken(claims.UserID, claims.UserType, claims.Role)
	if err != nil {
		return nil, err
	}

	newRefresh, err := s.jwtManager.GenerateRefreshToken(claims.UserID, claims.UserType, claims.Role)
	if err != nil {
		return nil, err
	}

	if err := util.StoreToken(s.rdb, userID, newRefresh, 7*24*time.Hour); err != nil {
		return nil, err
	}

	return &RefreshTokenResponse{
		AccessToken:  newAccess,
		RefreshToken: newRefresh,
	}, nil
}
