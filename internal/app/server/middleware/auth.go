package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/pkg/jwt"
)

// DefaultCookieName is the default cookie name for authentication token.
const DefaultCookieName = "auth_token"

// AuthenticatedUserKey should be used to read userID from context.
const AuthenticatedUserKey ctxKey = 0

// Auth is an authentication middleware.
//
// An Auth instance is shared by all requests, so it holds no request-scoped
// data: the authenticated user ID is carried from the authentication step to
// the handler as a local variable and then via the request context.
//
// User NewAuth to create a new Auth instance.
type Auth struct {
	cookieName string
	jwt        *jwt.JWT
	store      store.Store
	log        *zap.SugaredLogger
}

// Options represents dependencies required for Auth.
type Options struct {
	CookieName string
	JWT        *jwt.JWT
	Store      store.Store
	Log        *zap.SugaredLogger
}

type ctxKey int

// NewAuth creates a new Auth middleware.
func NewAuth(opts Options) *Auth {
	cookieName := opts.CookieName
	if cookieName == "" {
		cookieName = DefaultCookieName
	}

	return &Auth{
		cookieName: cookieName,
		jwt:        opts.JWT,
		store:      opts.Store,
		log:        opts.Log,
	}
}

// PublicAPIAuth defines authentication handler for public API scope.
func (a *Auth) PublicAPIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		var err error

		cookie, err := r.Cookie(a.cookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				a.log.Debug("cookie header not found, issue new token")
				userID, err = a.registerNewUser(r.Context(), w)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		if cookie != nil {
			a.log.With("cookie", cookie).Debug("got cookie from request")
			userID, err = a.jwt.GetUserID(cookie.Value)
			if err != nil {
				userID, err = a.registerNewUser(r.Context(), w)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}

			if _, err := a.store.GetUser(r.Context(), userID); err != nil {
				a.handleUnauthorized(w, errors.New("got unregistered user from auth token"))
				return
			}
		}

		a.continueWithUser(w, r, next, userID)
	})
}

// PrivateAPIAuth defines authentication handler for private API scope.
func (a *Auth) PrivateAPIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(a.cookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				a.handleUnauthorized(w, errors.New("no authorization cookie"))
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}
			return
		}

		userID, err := a.jwt.GetUserID(cookie.Value)
		if err != nil {
			a.handleUnauthorized(w, err)
			return
		}
		if _, err := a.store.GetUser(r.Context(), userID); err != nil {
			a.handleUnauthorized(w, errors.New("got unregistered user from auth token"))
			return
		}

		a.continueWithUser(w, r, next, userID)
	})
}

func (a *Auth) registerNewUser(ctx context.Context, w http.ResponseWriter) (string, error) {
	user := models.User{
		ID:        uuid.NewString(),
		CreatedAt: time.Now(),
	}

	if err := a.store.CreateUser(ctx, user); err != nil {
		return user.ID, err
	}

	token, err := a.jwt.GetString(user.ID)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{Name: a.cookieName, Value: token})

	a.log.With("userID", user.ID).Info("registered new user")

	return user.ID, nil
}

func (a *Auth) handleUnauthorized(w http.ResponseWriter, err error) {
	a.log.Debugf("unauthorized attempt to access private API: %s", err)
	w.WriteHeader(http.StatusUnauthorized)
}

func (a *Auth) continueWithUser(w http.ResponseWriter, r *http.Request, next http.Handler, userID string) {
	a.log.With("userID", userID).Debug("add userID to request context")
	ctx := context.WithValue(r.Context(), AuthenticatedUserKey, userID)
	next.ServeHTTP(w, r.WithContext(ctx))
}
