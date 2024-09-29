package jwt

import (
	"testing"
	"time"

	j "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWT(t *testing.T) {
	jwt := New(Options{
		Secret:   []byte("secret_key"),
		Duration: time.Hour,
	})

	userID := uuid.NewString()

	tokenString, err := jwt.GetString(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	decodedUserID, err := jwt.GetUserID(tokenString)
	require.NoError(t, err)
	assert.Equal(t, userID, decodedUserID)
}

func TestDuration(t *testing.T) {
	jwt := New(Options{
		Secret:   []byte("secret_key"),
		Duration: -time.Hour,
	})

	userID := uuid.NewString()

	tokenString, err := jwt.GetString(userID)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	decodedUserID, err := jwt.GetUserID(tokenString)
	assert.Empty(t, decodedUserID)
	var targetErr *j.ValidationError
	assert.ErrorAs(t, err, &targetErr)
}

func TestInvalidSignature(t *testing.T) {
	secretKey := []byte("secret_key")

	jwt := New(Options{
		Secret:   secretKey,
		Duration: time.Hour,
	})

	token := j.NewWithClaims(j.SigningMethodHS256, j.RegisteredClaims{})
	tokenString, err := token.SignedString(secretKey)
	require.NoError(t, err)

	decodedUserID, err := jwt.GetUserID(tokenString + "invalid_data")
	assert.Empty(t, decodedUserID)
	assert.ErrorIs(t, err, j.ErrTokenSignatureInvalid)
}
