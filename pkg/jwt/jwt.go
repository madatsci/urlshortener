// Package jwt implements operations with JWT tokens.
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

var errInvalidToken = errors.New("invalid JWT token")

type (
	// JWT represents data required to create and sign JWT token.
	//
	// Use New to create a new instance of JWT.
	JWT struct {
		Secret []byte
		claims Claims
	}

	// Claims represents JWT token claims.
	Claims struct {
		jwt.RegisteredClaims
		UserID string
	}

	// Options is used to initialize a new JWT.
	Options struct {
		Secret   []byte
		Duration time.Duration
		Issuer   string
	}
)

// New creates a new instance of JWT.
func New(opts Options) *JWT {
	return &JWT{
		Secret: opts.Secret,
		claims: Claims{
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(opts.Duration)),
				Issuer:    opts.Issuer,
			},
		},
	}
}

// GetString returns signed JWT token as string.
func (j *JWT) GetString(userID string) (string, error) {
	j.claims.UserID = userID
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, j.claims)

	tokenString, err := token.SignedString(j.Secret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// GetUserID parses user ID from token.
func (j *JWT) GetUserID(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &j.claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return j.Secret, nil
	})
	if err != nil {
		return "", errors.Wrap(err, "token parsing error")
	}

	if !token.Valid {
		return "", errInvalidToken
	}

	return j.claims.UserID, nil
}
