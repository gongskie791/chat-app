package util

import "net/http"

const (
	refreshTokenCookieName = "refresh_token"
	refreshTokenMaxAge     = 7 * 24 * 60 * 60
)

// setRefreshTokenCookie sets the refresh token as an HTTP-only	cookie
func SetRefreshTokenCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    token,
		Path:     "/",
		MaxAge:   refreshTokenMaxAge,
		HttpOnly: true, // Not accessible via JavaScript
		Secure:   false,
		SameSite: http.SameSiteLaxMode, // Prevents CSRF
	})
}

// clearRefreshTokenCookie removes the refresh token cookie
func ClearRefreshTokenCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     refreshTokenCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Immediately expire
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
}
