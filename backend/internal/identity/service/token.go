package service

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	identityerrors "github.com/gmorini/inge-soft-3/backend/internal/identity/errors"
	"github.com/golang-jwt/jwt/v5"
)

const tokenLifetime = 30 * time.Minute

type tokenClaims struct {
	Username string `json:"username"`
	jwt.RegisteredClaims
}

type TokenManager struct {
	secret []byte
	now    func() time.Time
}

type AuthenticatedIdentity struct {
	UserID   int64
	Username string
}

func NewTokenManager(secret string) (*TokenManager, error) {
	return newTokenManager(secret, time.Now)
}

func newTokenManager(secret string, now func() time.Time) (*TokenManager, error) {
	if len([]byte(secret)) < 32 {
		return nil, errors.New("JWT secret must contain at least 32 bytes")
	}
	if now == nil {
		return nil, errors.New("token clock is required")
	}
	return &TokenManager{secret: []byte(secret), now: now}, nil
}

func (m *TokenManager) Issue(userID int64, username string) (string, error) {
	issuedAt := m.now().UTC().Truncate(time.Second)
	claims := tokenClaims{
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			IssuedAt:  jwt.NewNumericDate(issuedAt),
			ExpiresAt: jwt.NewNumericDate(issuedAt.Add(tokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	if err != nil {
		return "", fmt.Errorf("sign access token: %w", err)
	}
	return signed, nil
}

func (m *TokenManager) Validate(tokenValue string) (AuthenticatedIdentity, error) {
	claims := &tokenClaims{}
	token, err := jwt.ParseWithClaims(
		tokenValue,
		claims,
		func(token *jwt.Token) (any, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, identityerrors.ErrInvalidToken
			}
			return m.secret, nil
		},
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithExpirationRequired(),
		jwt.WithIssuedAt(),
		jwt.WithTimeFunc(m.now),
	)
	if err != nil || !token.Valid || claims.ExpiresAt == nil || claims.IssuedAt == nil {
		return AuthenticatedIdentity{}, identityerrors.ErrInvalidToken
	}

	userID, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil || userID < 1 || strings.TrimSpace(claims.Username) == "" {
		return AuthenticatedIdentity{}, identityerrors.ErrInvalidToken
	}
	return AuthenticatedIdentity{UserID: userID, Username: claims.Username}, nil
}
