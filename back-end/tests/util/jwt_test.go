package util_test

import (
	"testing"
	"time"

	"chat-app/back-end/internal/util"

	"github.com/google/uuid"
)

func newManager(accessExpiry, refreshExpiry time.Duration) *util.JWTManager {
	return util.NewJWTManager("test-secret-key", accessExpiry, refreshExpiry)
}

func TestGenerateAndValidateAccessToken(t *testing.T) {
	mgr := newManager(15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()

	token, err := mgr.GenerateAccessToken(userID, util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("GenerateAccessToken: %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
	if claims.TokenType != util.AccessToken {
		t.Errorf("expected token type %q, got %q", util.AccessToken, claims.TokenType)
	}
	if claims.UserType != util.UserTypeUser {
		t.Errorf("expected user type %q, got %q", util.UserTypeUser, claims.UserType)
	}
}

func TestGenerateAndValidateRefreshToken(t *testing.T) {
	mgr := newManager(15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()

	token, err := mgr.GenerateRefreshToken(userID, util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("GenerateRefreshToken: %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if claims.TokenType != util.RefreshToken {
		t.Errorf("expected token type %q, got %q", util.RefreshToken, claims.TokenType)
	}
	if claims.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, claims.UserID)
	}
}

func TestValidateToken_WrongSecret(t *testing.T) {
	mgr1 := newManager(15*time.Minute, 7*24*time.Hour)
	mgr2 := util.NewJWTManager("different-secret", 15*time.Minute, 7*24*time.Hour)

	token, err := mgr1.GenerateAccessToken(uuid.New(), util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = mgr2.ValidateToken(token)
	if err == nil {
		t.Fatal("expected error for wrong secret, got nil")
	}
}

func TestValidateToken_Expired(t *testing.T) {
	mgr := newManager(-1*time.Second, 7*24*time.Hour)

	token, err := mgr.GenerateAccessToken(uuid.New(), util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = mgr.ValidateToken(token)
	if err != util.ErrExpireToken {
		t.Errorf("expected ErrExpireToken, got %v", err)
	}
}

func TestValidateToken_Tampered(t *testing.T) {
	mgr := newManager(15*time.Minute, 7*24*time.Hour)

	token, err := mgr.GenerateAccessToken(uuid.New(), util.UserTypeUser, "")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	_, err = mgr.ValidateToken(token + "x")
	if err == nil {
		t.Fatal("expected error for tampered token, got nil")
	}
}

func TestValidateToken_GarbageInput(t *testing.T) {
	mgr := newManager(15*time.Minute, 7*24*time.Hour)

	_, err := mgr.ValidateToken("this.is.not.a.jwt")
	if err == nil {
		t.Fatal("expected error for garbage token, got nil")
	}
}

func TestAccessToken_RolePreserved(t *testing.T) {
	mgr := newManager(15*time.Minute, 7*24*time.Hour)
	userID := uuid.New()

	token, err := mgr.GenerateAccessToken(userID, util.UserTypeAdmin, "moderator")
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	claims, err := mgr.ValidateToken(token)
	if err != nil {
		t.Fatalf("validate: %v", err)
	}
	if claims.Role != "moderator" {
		t.Errorf("expected role %q, got %q", "moderator", claims.Role)
	}
	if claims.UserType != util.UserTypeAdmin {
		t.Errorf("expected user type %q, got %q", util.UserTypeAdmin, claims.UserType)
	}
}
