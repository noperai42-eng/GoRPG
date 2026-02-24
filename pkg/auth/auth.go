package auth

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"rpg-game/pkg/db"
)

var (
	ErrInvalidUsername    = errors.New("username must be 3-30 characters, alphanumeric and underscores only")
	ErrInvalidPassword   = errors.New("password must be at least 6 characters")
	ErrUsernameExists    = errors.New("username already exists")
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrInvalidToken      = errors.New("invalid or expired token")
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,30}$`)

// AuthService provides user registration, login, and JWT token management.
type AuthService struct {
	store     *db.Store
	jwtSecret []byte
}

// NewAuthService creates a new AuthService with the given store and JWT secret.
func NewAuthService(store *db.Store, jwtSecret string) *AuthService {
	return &AuthService{
		store:     store,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new user account after validating the username and password.
// Returns the new account ID on success.
func (a *AuthService) Register(username, password string) (int64, error) {
	if !usernameRegex.MatchString(username) {
		return 0, ErrInvalidUsername
	}
	if len(password) < 6 {
		return 0, ErrInvalidPassword
	}

	// Check if username already exists
	existing, err := a.store.GetAccountByUsername(username)
	if err != nil {
		return 0, fmt.Errorf("checking username: %w", err)
	}
	if existing != nil {
		return 0, ErrUsernameExists
	}

	hash, err := HashPassword(password)
	if err != nil {
		return 0, fmt.Errorf("hashing password: %w", err)
	}

	id, err := a.store.CreateAccount(username, hash)
	if err != nil {
		return 0, fmt.Errorf("creating account: %w", err)
	}
	return id, nil
}

// Login authenticates a user and returns a signed JWT token string.
func (a *AuthService) Login(username, password string) (string, error) {
	acct, err := a.store.GetAccountByUsername(username)
	if err != nil {
		return "", fmt.Errorf("looking up account: %w", err)
	}
	if acct == nil {
		return "", ErrInvalidCredentials
	}

	if !CheckPassword(acct.PasswordHash, password) {
		return "", ErrInvalidCredentials
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":      strconv.FormatInt(acct.ID, 10),
		"username": acct.Username,
		"exp":      jwt.NewNumericDate(now.Add(24 * time.Hour)),
		"iat":      jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(a.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("signing token: %w", err)
	}
	return tokenString, nil
}

// ValidateToken parses and validates a JWT token string.
// Returns the account ID, username, and any error.
func (a *AuthService) ValidateToken(tokenString string) (int64, string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return a.jwtSecret, nil
	})
	if err != nil {
		return 0, "", ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, "", ErrInvalidToken
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return 0, "", ErrInvalidToken
	}

	accountID, err := strconv.ParseInt(sub, 10, 64)
	if err != nil {
		return 0, "", ErrInvalidToken
	}

	username, ok := claims["username"].(string)
	if !ok || username == "" {
		return 0, "", ErrInvalidToken
	}

	return accountID, username, nil
}

// HashPassword hashes a plaintext password using bcrypt with cost 12.
func HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

// CheckPassword compares a bcrypt hash with a plaintext password.
// Returns true if they match.
func CheckPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
