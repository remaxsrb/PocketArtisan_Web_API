package auth

import (
	"PocketArtisan/config"
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTService interface {
	Generate(identity Identity) (string, error)
	Validate(token string) (*Identity, error)
}

type jwtService struct {
	secret []byte
	ttl    time.Duration
}

var (
	instance JWTService
	once     sync.Once
)

func InitJWTService(ttl time.Duration) {
	once.Do(func() {
		secret := config.GetCrypto().JwtKeySecret

		instance = &jwtService{
			secret: []byte(secret),
			ttl:    ttl,
		}

	})
}

func GetJWTService() JWTService {
	if instance == nil {
		panic("JWTService not initialized")
	}
	return instance
}

type claims struct {
	UserID      string `json:"uid"`
	Role        string `json:"role"`
	CraftsmanID string `json:"craftsman_id,omitempty"`
	jwt.RegisteredClaims
}

func (j *jwtService) Generate(identity Identity) (string, error) {
	now := time.Now()

	claims := claims{
		UserID:      identity.ID,
		Role:        identity.Role,
		CraftsmanID: identity.CraftsmanID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(j.ttl)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString(j.secret)
}

func (j *jwtService) Validate(tokenStr string) (*Identity, error) {
	token, err := jwt.ParseWithClaims(
		tokenStr,
		&claims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return j.secret, nil
		},
	)

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	claims, ok := token.Claims.(*claims)
	if !ok {
		return nil, errors.New("invalid claims")
	}

	return &Identity{
		ID:          claims.UserID,
		Role:        claims.Role,
		CraftsmanID: claims.CraftsmanID,
	}, nil
}
