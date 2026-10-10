// Package jwt implements operations with JWT tokens.
package jwt

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/pkg/errors"
)

var (
	errInvalidToken = errors.New("invalid JWT token")
	errNoUserID     = errors.New("token does not contain user ID")
)

// JWT represents data required to create and sign JWT token.
//
// A JWT holds only the data shared by all issued tokens (secret, lifetime,
// issuer). Claims are built and parsed on every call, therefore a single
// instance is safe for concurrent use by any number of goroutines and tokens
// issued for different users can never interfere with each other.
//
// Use New to create a new instance of JWT.
type JWT struct {
	Secret   []byte
	duration time.Duration
	issuer   string
}

// Claims represents JWT token claims.
type Claims struct {
	jwt.RegisteredClaims
	UserID string
}

// Options is used to initialize a new JWT.
type Options struct {
	Secret   []byte
	Duration time.Duration
	Issuer   string
}

// New creates a new instance of JWT.
func New(opts Options) *JWT {
	return &JWT{
		Secret:   opts.Secret,
		duration: opts.Duration,
		issuer:   opts.Issuer,
	}
}

// GetString returns signed JWT token as string.
func (j *JWT) GetString(userID string) (string, error) {
	// Claims must be local to this call: sharing them between calls would let
	// concurrent goroutines sign a token with somebody else's user ID.
	claims := Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.duration)),
			Issuer:    j.issuer,
		},
		UserID: userID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(j.Secret)
	if err != nil {
		return "", errors.Wrap(err, "token signing error")
	}

	return tokenString, nil
}

// GetUserID parses user ID from token.
//
// It returns an error if the token is malformed, expired, signed with an
// unexpected method, or does not identify a user.
func (j *JWT) GetUserID(tokenString string) (string, error) {
	// Claims must be local to this call: decoding into shared state would both
	// race with concurrent calls and make claims missing from the token (such
	// as UserID) silently inherit the values of a previously parsed token.
	var claims Claims

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (interface{}, error) {
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

	if claims.UserID == "" {
		return "", errNoUserID
	}

	return claims.UserID, nil
}
