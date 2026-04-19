package handler

import (
	"chat-app/back-end/internal/service"
	"chat-app/back-end/internal/util"
	"context"
	"errors"
	"net/http"
)

// AuthServiceI is the interface the AuthHandler depends on.
type AuthServiceI interface {
	RefreshToken(ctx context.Context, token string) (*service.RefreshTokenResponse, error)
}

// compile-time check
var _ AuthServiceI = (*service.AuthService)(nil)

type AuthHandler struct {
	authService AuthServiceI
}

func NewAuthHandler(authService AuthServiceI) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// POST /api/auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		util.Error(w, http.StatusUnauthorized, "No refresh token found")
		return
	}

	resp, err := h.authService.RefreshToken(r.Context(), cookie.Value)
	if err != nil {
		if errors.Is(err, service.ErrRefreshTokenInvalid) {
			util.ClearRefreshTokenCookie(w)
			util.Error(w, http.StatusUnauthorized, "Invalid or expired refresh token")
			return
		}
		util.Error(w, http.StatusInternalServerError, "Token refresh failed")
		return
	}

	util.SetRefreshTokenCookie(w, resp.RefreshToken)
	util.Success(w, http.StatusOK, "Token refreshed successfully", resp.AccessToken)
}
