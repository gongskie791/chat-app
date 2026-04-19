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

// UserServiceI is the interface the UserHandler depends on.
type UserServiceI interface {
	RegisterAccount(ctx context.Context, data *model.CreateUserRequest) (*model.UserLoginResponse, error)
	UserLogin(ctx context.Context, data *model.UserLoginRequest) (*model.UserLoginResponse, error)
	UserLogout(ctx context.Context, userID uuid.UUID) error
}

// compile-time check
var _ UserServiceI = (*service.UserService)(nil)

type UserHandler struct {
	userService UserServiceI
}

func NewUserHandler(userService UserServiceI) *UserHandler {
	return &UserHandler{userService: userService}
}

// POST /api/auth/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if fields := util.ValidateStruct(&req); fields != nil {
		util.ValidationFailed(w, fields)
		return
	}

	resp, err := h.userService.RegisterAccount(r.Context(), &req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrEmailTaken):
			util.Error(w, http.StatusConflict, "Email already registered")
		case errors.Is(err, repository.ErrUsernameTaken):
			util.Error(w, http.StatusConflict, "Username already taken")
		default:
			util.Error(w, http.StatusInternalServerError, "Register failed")
		}
		return
	}

	util.SetRefreshTokenCookie(w, resp.RefreshToken)
	resp.RefreshToken = ""

	util.Success(w, http.StatusCreated, "Registration successful", resp)
}

// POST /api/auth/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.UserLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if fields := util.ValidateStruct(&req); fields != nil {
		util.ValidationFailed(w, fields)
		return
	}

	resp, err := h.userService.UserLogin(r.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			util.Error(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		util.Error(w, http.StatusInternalServerError, "Login failed")
		return
	}

	util.SetRefreshTokenCookie(w, resp.RefreshToken)
	resp.RefreshToken = ""

	util.Success(w, http.StatusOK, "Login successful", resp)
}

// POST /api/auth/logout
func (h *UserHandler) Logout(w http.ResponseWriter, r *http.Request) {
	claims := util.ClaimsFromContext(r)
	if claims == nil {
		util.Error(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	if err := h.userService.UserLogout(r.Context(), claims.UserID); err != nil {
		util.Error(w, http.StatusInternalServerError, "Logout failed")
		return
	}

	util.ClearRefreshTokenCookie(w)
	util.Success[any](w, http.StatusOK, "Logout successful", nil)
}
