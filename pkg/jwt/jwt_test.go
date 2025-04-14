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
	secret := []byte("secret_key")

	t.Run("success", func(t *testing.T) {
		jwt := New(Options{
			Secret:   secret,
			Duration: time.Hour,
		})

		userID := uuid.NewString()

		tokenString, err := jwt.GetString(userID)
		require.NoError(t, err)
		assert.NotEmpty(t, tokenString)

		decodedUserID, err := jwt.GetUserID(tokenString)
		require.NoError(t, err)
		assert.Equal(t, userID, decodedUserID)
	})

	t.Run("expired token", func(t *testing.T) {
		jwt := New(Options{
			Secret:   secret,
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
	})

	t.Run("invalid signature", func(t *testing.T) {
		jwt := New(Options{
			Secret:   secret,
			Duration: time.Hour,
		})

		token := j.NewWithClaims(j.SigningMethodHS256, j.RegisteredClaims{})
		tokenString, err := token.SignedString(secret)
		require.NoError(t, err)

		decodedUserID, err := jwt.GetUserID(tokenString + "invalid_data")
		assert.Empty(t, decodedUserID)
		assert.ErrorIs(t, err, j.ErrTokenSignatureInvalid)
	})
}

func BenchmarkGetString(b *testing.B) {
	jwt := New(Options{
		Secret:   []byte("secret_key"),
		Duration: time.Hour,
	})

	for i := 0; i < b.N; i++ {
		b.StopTimer()
		userID := uuid.NewString()
		b.StartTimer()

		jwt.GetString(userID)
	}
}
