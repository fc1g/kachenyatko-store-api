package auth

import (
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type TokenManager struct {
	privateKey ed25519.PrivateKey
	publicKey  ed25519.PublicKey
	issuer     string
	ttl        time.Duration
}

func NewTokenManager(privateKeyBase64, issuer string, ttl time.Duration) (*TokenManager, error) {
	pemBytes, err := base64.StdEncoding.DecodeString(privateKeyBase64)
	if err != nil {
		return nil, fmt.Errorf("decode private key: %w", err)
	}

	privateKey, err := jwt.ParseEdPrivateKeyFromPEM(pemBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key: %w", err)
	}

	edKey, ok := privateKey.(ed25519.PrivateKey)
	if !ok {
		return nil, errors.New("private key is not ed25519")
	}

	publicKey := edKey.Public().(ed25519.PublicKey)

	return &TokenManager{
		privateKey: edKey,
		publicKey:  publicKey,
		issuer:     issuer,
		ttl:        ttl,
	}, nil
}

func (m *TokenManager) Issue(userID uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.ttl)

	token := jwt.NewWithClaims(jwt.SigningMethodEdDSA, jwt.RegisteredClaims{
		Issuer:    m.issuer,
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
	})

	signed, err := token.SignedString(m.privateKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign token: %w", err)
	}

	return signed, expiresAt, nil
}

var ErrInvalidToken = errors.New("invalid token")

func (m *TokenManager) Verify(tokenString string) (uuid.UUID, error) {
	var claims jwt.RegisteredClaims

	_, err := jwt.ParseWithClaims(tokenString, &claims,
		func(*jwt.Token) (any, error) { return m.publicKey, nil },
		jwt.WithValidMethods([]string{jwt.SigningMethodEdDSA.Alg()}),
		jwt.WithIssuer(m.issuer),
		jwt.WithExpirationRequired(),
	)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: invalid subject: %w", ErrInvalidToken, err)
	}

	return userID, nil
}
