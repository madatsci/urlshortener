package jwt

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

var errInvalidToken = errors.New("invalid JWT token")

type (
	JWT struct {
		Secret   []byte
		Duration time.Duration
		claims   Claims
	}

	Claims struct {
		jwt.RegisteredClaims
		UserID string
	}

	Options struct {
		Secret   []byte
		Duration time.Duration
		Issuer   string
	}
)

func New(opts Options) *JWT {
	return &JWT{
		Secret:   opts.Secret,
		Duration: opts.Duration,
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
		return "", err
	}

	if !token.Valid {
		return "", errInvalidToken
	}

	return j.claims.UserID, nil
}
