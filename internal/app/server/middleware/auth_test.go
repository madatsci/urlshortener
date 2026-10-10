package middleware

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/madatsci/urlshortener/internal/app/models"
	memstore "github.com/madatsci/urlshortener/internal/app/store/memory"
	"github.com/madatsci/urlshortener/pkg/jwt"
)

const testTokenSecret = "secret_key"

// userIDFromContext returns the user ID that the middleware put into the
// request context.
func userIDFromContext(r *http.Request) string {
	userID, _ := r.Context().Value(AuthenticatedUserKey).(string)
	return userID
}

func newTestAuth(store *memstore.Store) *Auth {
	return NewAuth(Options{
		JWT: jwt.New(jwt.Options{
			Secret:   []byte(testTokenSecret),
			Duration: time.Hour,
			Issuer:   "urlshortener",
		}),
		Store: store,
		Log:   zap.NewNop().Sugar(),
	})
}

// TestPrivateAPIAuthIsolatesConcurrentUsers checks that concurrent requests are
// attributed to the user from their own token and never to another request's
// user.
func TestPrivateAPIAuthIsolatesConcurrentUsers(t *testing.T) {
	store := memstore.New()
	auth := newTestAuth(store)

	const users = 24

	tokens := make([]string, users)
	wantIDs := make([]string, users)

	for i := range users {
		user := models.User{ID: uuid.NewString(), CreatedAt: time.Now()}
		require.NoError(t, store.CreateUser(t.Context(), user))

		token, err := auth.jwt.GetString(user.ID)
		require.NoError(t, err)

		tokens[i] = token
		wantIDs[i] = user.ID
	}

	// The barrier keeps every request inside the handler until all of them have
	// arrived, which is exactly the situation that used to mix users up.
	var barrier sync.WaitGroup
	barrier.Add(users)

	gate := make(chan struct{})
	seen := make(chan string, users)

	handler := auth.PrivateAPIAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		barrier.Done()
		<-gate

		seen <- userIDFromContext(r)
		w.WriteHeader(http.StatusOK)
	}))

	var wg sync.WaitGroup
	for i := range users {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
			req.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: tokens[i]})
			handler.ServeHTTP(httptest.NewRecorder(), req)
		}(i)
	}

	barrier.Wait()
	close(gate)

	wg.Wait()
	close(seen)

	got := make([]string, 0, users)
	for userID := range seen {
		got = append(got, userID)
	}

	assert.ElementsMatch(t, wantIDs, got,
		"each request must be attributed to the user from its own token")
}

// TestPublicAPIAuthIsolatesConcurrentUsers checks that concurrently registered
// visitors each receive a cookie that resolves back to their own new user.
func TestPublicAPIAuthIsolatesConcurrentUsers(t *testing.T) {
	store := memstore.New()
	auth := newTestAuth(store)

	const visitors = 24

	var mu sync.Mutex
	contextUserIDs := make([]string, 0, visitors)
	tokenUserIDs := make([]string, 0, visitors)

	var wg sync.WaitGroup
	for range visitors {
		wg.Add(1)

		go func() {
			defer wg.Done()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", nil)

			auth.PublicAPIAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				contextUserIDs = append(contextUserIDs, userIDFromContext(r))
				mu.Unlock()

				w.WriteHeader(http.StatusCreated)
			})).ServeHTTP(rec, req)

			var token string
			for _, c := range rec.Result().Cookies() {
				if c.Name == DefaultCookieName {
					token = c.Value
				}
			}
			require.NotEmpty(t, token, "a new visitor must be issued a cookie")

			// The cookie handed to this visitor must resolve to that same visitor.
			decoded, err := auth.jwt.GetUserID(token)
			require.NoError(t, err)

			mu.Lock()
			tokenUserIDs = append(tokenUserIDs, decoded)
			mu.Unlock()
		}()
	}

	wg.Wait()

	require.Len(t, contextUserIDs, visitors)
	require.Len(t, tokenUserIDs, visitors)

	// Every visitor got a distinct user, and the user in the cookie is exactly
	// the user the handler was given for that same request.
	assert.ElementsMatch(t, contextUserIDs, tokenUserIDs,
		"the issued cookie must identify the same user the request was attributed to")

	distinct := make(map[string]struct{}, visitors)
	for _, userID := range contextUserIDs {
		distinct[userID] = struct{}{}
	}
	assert.Len(t, distinct, visitors, "every visitor must get a distinct user ID")
}

// TestPublicAPIAuthIgnoresTokenWithoutUserID checks that a token which does not
// identify a user is never resolved to another user's session.
func TestPublicAPIAuthIgnoresTokenWithoutUserID(t *testing.T) {
	store := memstore.New()
	auth := newTestAuth(store)
	ctx := t.Context()

	victim := models.User{ID: uuid.NewString(), CreatedAt: time.Now()}
	require.NoError(t, store.CreateUser(ctx, victim))

	// Authenticate the victim first, so that a leaky implementation would have
	// its claims lying around.
	victimToken, err := auth.jwt.GetString(victim.ID)
	require.NoError(t, err)

	decoded, err := auth.jwt.GetUserID(victimToken)
	require.NoError(t, err)
	require.Equal(t, victim.ID, decoded)

	// A token signed with the correct secret but carrying no UserID claim.
	noUserIDToken, err := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, jwtlib.RegisteredClaims{
		Issuer:    "urlshortener",
		ExpiresAt: jwtlib.NewNumericDate(time.Now().Add(time.Hour)),
	}).SignedString([]byte(testTokenSecret))
	require.NoError(t, err)

	// The signer itself must reject such a token.
	_, err = auth.jwt.GetUserID(noUserIDToken)
	require.Error(t, err)

	var gotUserID string
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.AddCookie(&http.Cookie{Name: DefaultCookieName, Value: noUserIDToken})

	auth.PublicAPIAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotUserID = userIDFromContext(r)
	})).ServeHTTP(rec, req)

	assert.NotEqual(t, victim.ID, gotUserID,
		"a token without a user ID must not resolve to another user")
}

// TestPublicAPIAuthSetsDistinctCookies checks that concurrently registered
// visitors never receive the same token value.
func TestPublicAPIAuthSetsDistinctCookies(t *testing.T) {
	store := memstore.New()
	auth := newTestAuth(store)

	const visitors = 24

	var mu sync.Mutex
	cookies := make(map[string]string, visitors)

	var wg sync.WaitGroup
	for i := range visitors {
		wg.Add(1)

		go func(i int) {
			defer wg.Done()

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", nil)

			auth.PublicAPIAuth(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(http.StatusCreated)
			})).ServeHTTP(rec, req)

			for _, c := range rec.Result().Cookies() {
				if c.Name == DefaultCookieName {
					mu.Lock()
					cookies[c.Value] = userIDFromContext(req)
					mu.Unlock()
				}
			}
		}(i)
	}

	wg.Wait()
	assert.Len(t, cookies, visitors, "every visitor must receive a distinct cookie")
}
