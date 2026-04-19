package middleware

import (
	"chat-app/back-end/internal/util"
	"net/http"
)

// RequireRole allows access only to users whose role is in the allowed list.
// R must be a string-based type (e.g. model.UserRole).
// Must be used after VerifyJWT.
func RequireRole[R ~string](roles ...R) func(http.Handler) http.Handler {
	allowed := make(map[R]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value("claims").(*util.Claims)
			if !ok || claims == nil {
				util.Error(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			if _, ok := allowed[R(claims.Role)]; !ok {
				util.Error(w, http.StatusForbidden, "Forbidden: insufficient role")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
