package tokenmanager

import (
	"fmt"
	"math/rand"
	"test-auth/internal/service"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Manager struct {
	signingKey string
}

func NewManager(signingKey string) (*Manager, error) {
	return &Manager{signingKey: signingKey}, nil
}

func (m *Manager) NewJWT(userId string, ttl time.Duration, refreshTokenHash string) (string, error) {
	type CustomClaims struct {
		RefreshHash string `json:"key"`
		jwt.StandardClaims
	}

	c := CustomClaims{
		refreshTokenHash,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(ttl).Unix(),
			Subject:   userId,
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS512, c)

	return token.SignedString([]byte(m.signingKey))
}

func (m *Manager) Parse(accessToken string) (service.AccessToken, error) {
	token, err := jwt.Parse(accessToken, func(token *jwt.Token) (i interface{}, err error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(m.signingKey), nil
	})
	if err != nil {
		return service.AccessToken{}, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return service.AccessToken{}, fmt.Errorf("error get user claims from token")
	}
	res := service.AccessToken{
		ExpiresAt: claims["exp"].(float64),
		Subject:   claims["sub"].(string),
		Key:       claims["key"].(string),
	}
	return res, nil
}

func (m *Manager) NewRefreshToken() (string, error) {
	b := make([]byte, 24)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}

func (m *Manager) GetRandomString(l int) (string, error) {
	b := make([]byte, l)

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	if _, err := r.Read(b); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", b), nil
}
