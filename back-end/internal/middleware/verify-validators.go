package middleware

import (
	"chat-app/back-end/internal/util"
	"context"
	"net/http"
	"strings"
)

func VerifyJWT(jwt *util.JWTManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var tokenStr string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				tokenStr = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// this for mobile
				tokenStr = r.URL.Query().Get("token")
			}

			if tokenStr == "" {
				util.Error(w, http.StatusUnauthorized, "Missing or Invalid authorization header")
				return
			}

			claims, err := jwt.ValidateToken(tokenStr)
			if err != nil {
				util.Error(w, http.StatusUnauthorized, "Invalid or expired token")
				return
			}

			if claims.TokenType != util.AccessToken {
				util.Error(w, http.StatusUnauthorized, "Invalid token type")
				return
			}

			ctx := context.WithValue(r.Context(), "claims", claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
