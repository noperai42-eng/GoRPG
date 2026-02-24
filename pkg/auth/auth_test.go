package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"rpg-game/pkg/db"
)

const testSecret = "test-secret-key-for-jwt-signing"

// setupTestStore creates a temporary SQLite database and returns a Store
// along with a cleanup function that removes the temporary file.
func setupTestStore(t *testing.T) (*db.Store, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := db.NewStore(dbPath)
	if err != nil {
		t.Fatalf("failed to create test store: %v", err)
	}
	cleanup := func() {
		store.Close()
		os.Remove(dbPath)
	}
	return store, cleanup
}

func TestRegisterAndLogin(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	svc := NewAuthService(store, testSecret)

	// Register a new user.
	accountID, err := svc.Register("hero_one", "secret123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if accountID <= 0 {
		t.Fatalf("expected positive account ID, got %d", accountID)
	}

	// Login with correct credentials.
	tokenStr, err := svc.Login("hero_one", "secret123")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if tokenStr == "" {
		t.Fatal("expected non-empty token string")
	}

	// Validate the token.
	gotID, gotUsername, err := svc.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if gotID != accountID {
		t.Errorf("expected account ID %d, got %d", accountID, gotID)
	}
	if gotUsername != "hero_one" {
		t.Errorf("expected username %q, got %q", "hero_one", gotUsername)
	}
}

func TestInvalidPassword(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	svc := NewAuthService(store, testSecret)

	// Register a user.
	_, err := svc.Register("warrior", "goodpass123")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Login with wrong password.
	_, err = svc.Login("warrior", "wrongpass")
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}

	// Login with non-existent user.
	_, err = svc.Login("nobody", "anypass")
	if err == nil {
		t.Fatal("expected error for non-existent user, got nil")
	}
	if err != ErrInvalidCredentials {
		t.Errorf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestDuplicateUsername(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	svc := NewAuthService(store, testSecret)

	// Register first user.
	_, err := svc.Register("mage_123", "password1")
	if err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	// Attempt to register with the same username.
	_, err = svc.Register("mage_123", "password2")
	if err == nil {
		t.Fatal("expected error for duplicate username, got nil")
	}
	if err != ErrUsernameExists {
		t.Errorf("expected ErrUsernameExists, got: %v", err)
	}
}

func TestTokenValidation(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	svc := NewAuthService(store, testSecret)

	// Register and login.
	_, err := svc.Register("rogue", "sneaky99")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	tokenStr, err := svc.Login("rogue", "sneaky99")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	// Valid token should pass validation.
	id, username, err := svc.ValidateToken(tokenStr)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if id <= 0 {
		t.Errorf("expected positive account ID, got %d", id)
	}
	if username != "rogue" {
		t.Errorf("expected username %q, got %q", "rogue", username)
	}

	// Completely invalid token should fail.
	_, _, err = svc.ValidateToken("not.a.valid.token")
	if err == nil {
		t.Fatal("expected error for invalid token, got nil")
	}
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got: %v", err)
	}

	// Token signed with a different secret should fail.
	otherSvc := NewAuthService(store, "different-secret")
	_, _, err = otherSvc.ValidateToken(tokenStr)
	if err == nil {
		t.Fatal("expected error for token signed with different secret, got nil")
	}
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got: %v", err)
	}

	// Token with missing username claim should fail.
	claims := jwt.MapClaims{
		"sub": "1",
		"exp": jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		"iat": jwt.NewNumericDate(time.Now()),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	noUserToken, _ := token.SignedString([]byte(testSecret))
	_, _, err = svc.ValidateToken(noUserToken)
	if err == nil {
		t.Fatal("expected error for token without username claim, got nil")
	}
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestExpiredToken(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	svc := NewAuthService(store, testSecret)

	// Register a user.
	_, err := svc.Register("paladin", "holylight")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	// Manually create an expired token.
	acct, err := store.GetAccountByUsername("paladin")
	if err != nil || acct == nil {
		t.Fatalf("failed to get account: %v", err)
	}

	pastTime := time.Now().Add(-48 * time.Hour)
	claims := jwt.MapClaims{
		"sub":      "1",
		"username": "paladin",
		"exp":      jwt.NewNumericDate(pastTime.Add(24 * time.Hour)), // expired 24h ago
		"iat":      jwt.NewNumericDate(pastTime),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString([]byte(testSecret))
	if err != nil {
		t.Fatalf("failed to sign expired token: %v", err)
	}

	// Validate should fail for expired token.
	_, _, err = svc.ValidateToken(expiredToken)
	if err == nil {
		t.Fatal("expected error for expired token, got nil")
	}
	if err != ErrInvalidToken {
		t.Errorf("expected ErrInvalidToken, got: %v", err)
	}
}

func TestRegisterValidation(t *testing.T) {
	store, cleanup := setupTestStore(t)
	defer cleanup()

	svc := NewAuthService(store, testSecret)

	tests := []struct {
		name     string
		username string
		password string
		wantErr  error
	}{
		{"username too short", "ab", "validpass", ErrInvalidUsername},
		{"username too long", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "validpass", ErrInvalidUsername}, // 33 chars
		{"username with spaces", "bad name", "validpass", ErrInvalidUsername},
		{"username with special chars", "bad@name!", "validpass", ErrInvalidUsername},
		{"empty username", "", "validpass", ErrInvalidUsername},
		{"password too short", "goodname", "12345", ErrInvalidPassword},
		{"empty password", "goodname", "", ErrInvalidPassword},
		{"valid underscore username", "good_name_1", "validpass", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.Register(tt.username, tt.password)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if err != tt.wantErr {
					t.Errorf("expected error %v, got: %v", tt.wantErr, err)
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got: %v", err)
				}
			}
		})
	}
}
