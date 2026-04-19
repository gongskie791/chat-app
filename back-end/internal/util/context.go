package util

import "net/http"

// ClaimsFromContext extracts JWT claims stored by the VerifyJWT middleware.
func ClaimsFromContext(r *http.Request) *Claims {
	claims, _ := r.Context().Value("claims").(*Claims)
	return claims
}
