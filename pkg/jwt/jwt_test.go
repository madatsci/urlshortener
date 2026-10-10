package jwt

import (
	"fmt"
	"sync"
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

	t.Run("token without user ID", func(t *testing.T) {
		jwt := New(Options{
			Secret:   secret,
			Duration: time.Hour,
		})

		// A token signed with the correct secret but carrying no UserID claim
		// must not resolve to any user.
		token := j.NewWithClaims(j.SigningMethodHS256, j.RegisteredClaims{
			ExpiresAt: j.NewNumericDate(time.Now().Add(time.Hour)),
		})
		tokenString, err := token.SignedString(secret)
		require.NoError(t, err)

		decodedUserID, err := jwt.GetUserID(tokenString)
		assert.Error(t, err)
		assert.Empty(t, decodedUserID)
		assert.ErrorIs(t, err, errNoUserID)
	})

	t.Run("token without user ID does not inherit a previous user ID", func(t *testing.T) {
		jwt := New(Options{
			Secret:   secret,
			Duration: time.Hour,
		})

		// Parse a valid token first so that any state left behind by it
		// would be visible to the next call.
		validToken, err := jwt.GetString("user-from-previous-request")
		require.NoError(t, err)
		decoded, err := jwt.GetUserID(validToken)
		require.NoError(t, err)
		require.Equal(t, "user-from-previous-request", decoded)

		// A UserID-less token must be rejected, not resolved to the user above.
		emptyToken := j.NewWithClaims(j.SigningMethodHS256, j.RegisteredClaims{
			ExpiresAt: j.NewNumericDate(time.Now().Add(time.Hour)),
		})
		emptyTokenString, err := emptyToken.SignedString(secret)
		require.NoError(t, err)

		decodedUserID, err := jwt.GetUserID(emptyTokenString)
		assert.Error(t, err)
		assert.Empty(t, decodedUserID)
		assert.ErrorIs(t, err, errNoUserID)
	})

	t.Run("unexpected signing method", func(t *testing.T) {
		jwt := New(Options{
			Secret:   secret,
			Duration: time.Hour,
		})

		// "alg: none" with an empty signature.
		const noneToken = "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0." +
			"eyJVc2VySUQiOiJhdHRhY2tlciJ9."

		decodedUserID, err := jwt.GetUserID(noneToken)
		assert.Error(t, err)
		assert.Empty(t, decodedUserID)
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

// TestConcurrentUse checks that a single JWT instance can serve concurrent
// callers without tokens leaking between users.
//
// It is meaningful under `go test -race`, which reports a data race on the
// previous implementation.
func TestConcurrentUse(t *testing.T) {
	jwt := New(Options{
		Secret:   []byte("secret_key"),
		Duration: time.Hour,
	})

	const goroutines = 32
	const iterations = 50

	var wg sync.WaitGroup
	errCh := make(chan error, goroutines*iterations)

	for g := 0; g < goroutines; g++ {
		wg.Add(1)

		go func(g int) {
			defer wg.Done()

			for i := 0; i < iterations; i++ {
				userID := fmt.Sprintf("user-%02d-%02d", g, i)

				tokenString, err := jwt.GetString(userID)
				if err != nil {
					errCh <- fmt.Errorf("signing token for %s: %w", userID, err)
					continue
				}

				decodedUserID, err := jwt.GetUserID(tokenString)
				if err != nil {
					errCh <- fmt.Errorf("parsing token for %s: %w", userID, err)
					continue
				}

				if decodedUserID != userID {
					errCh <- fmt.Errorf("token issued for %s decoded to %s", userID, decodedUserID)
				}
			}
		}(g)
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
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
