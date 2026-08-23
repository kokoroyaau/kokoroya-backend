package jwtauth

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("invalid token")

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type Manager struct {
	secret []byte
	ttl    time.Duration
}

func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

func (m *Manager) Generate(userID int64, role string) (token string, jti string, expiresAt time.Time, err error) {
	jti = uuid.NewString()
	expiresAt = time.Now().Add(m.ttl)

	claims := Claims{
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "kokoroya",
			Subject:   strconv.FormatInt(userID, 10),
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(m.secret)
	return token, jti, expiresAt, err
}

// Parse verifies a token's signature, issuer, and expiry, returning its subject claims.
func (m *Manager) Parse(tokenStr string) (userID int64, role string, jti string, err error) {
	var claims Claims
	token, err := jwt.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("kokoroya"))
	if err != nil || !token.Valid {
		return 0, "", "", ErrInvalidToken
	}

	userID, err = strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return 0, "", "", ErrInvalidToken
	}

	return userID, claims.Role, claims.ID, nil
}
