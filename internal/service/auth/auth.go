package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"test-auth/internal/service"
	"test-auth/internal/storage"
	"time"
)

type PasswordHasher interface {
	Hash(string) (string, error)
	CheckHash(string, string) bool
}

type Repository interface {
	GetUser(context.Context, string) (storage.User, error)
	SetSession(context.Context, string, storage.Session) error
}

type TokenManager interface {
	NewJWT(string, time.Duration, string) (string, error)
	Parse(accessToken string) (service.AccessToken, error)
	NewRefreshToken() (string, error)
	GetRandomString(int) (string, error)
}

type ServiceAuth struct {
	hasher          PasswordHasher
	db              Repository
	tokenManager    TokenManager
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type Tokens struct {
	AccessToken  string `json:"access_token" binding:"required"`
	RefreshToken string `json:"refresh_token" binding:"required"`
}

var (
	ErrValidationError = errors.New("validation error")
)

func NewServiceAuth(hasher PasswordHasher, db Repository, tokenManager TokenManager, accessTokenTTL, refreshTokenTTL time.Duration) *ServiceAuth {
	return &ServiceAuth{
		hasher:          hasher,
		db:              db,
		tokenManager:    tokenManager,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *ServiceAuth) SignIn(ctx context.Context, guid string) (Tokens, error) {
	user, err := s.db.GetUser(ctx, guid)
	if err != nil {
		return Tokens{}, err
	}

	return s.createSession(ctx, user)
}

func (s *ServiceAuth) RefreshTokens(ctx context.Context, accessToken, refreshToken string) (Tokens, error) {
	accessTokenClaims, err := s.tokenManager.Parse(accessToken)
	if err != nil {
		return Tokens{}, err
	}

	user, err := s.db.GetUser(ctx, accessTokenClaims.Subject)
	if err != nil {
		return Tokens{}, err
	}

	session, ok := findSession(accessTokenClaims.Key, user.Sessions)
	if !ok || !s.hasher.CheckHash(refreshToken, session.RefreshToken) {
		return Tokens{}, ErrValidationError
	}
	return s.createSession(ctx, user)
}

func findSession(key string, sessions []storage.Session) (storage.Session, bool) {
	for _, session := range sessions {
		if session.Key == key {
			return session, true
		}
	}
	return storage.Session{}, false
}

func (s *ServiceAuth) createSession(ctx context.Context, user storage.User) (Tokens, error) {
	var (
		res Tokens
		err error
	)

	key, err := s.tokenManager.GetRandomString(10)
	if err != nil {
		return Tokens{}, err
	}

	refreshToken, err := s.tokenManager.NewRefreshToken()
	if err != nil {
		return Tokens{}, err
	}

	refreshTokenDB, err := s.hasher.Hash(refreshToken)
	if err != nil {
		return Tokens{}, err
	}

	session := storage.Session{
		RefreshToken: refreshTokenDB,
		ExpiresAt:    time.Now().Add(s.refreshTokenTTL),
		Key:          key,
	}

	res.AccessToken, err = s.tokenManager.NewJWT(user.Guid, s.accessTokenTTL, key)
	if err != nil {
		return Tokens{}, err
	}

	err = s.db.SetSession(ctx, user.Guid, session)
	if err != nil {
		return Tokens{}, err
	}

	buf := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, buf)
	defer encoder.Close()
	_, err = encoder.Write([]byte(refreshToken))
	if err != nil {
		return Tokens{}, err
	}
	res.RefreshToken = buf.String()

	return res, err
}
