package middleware

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/madatsci/urlshortener/pkg/jwt"
	"go.uber.org/zap"
)

const DefaultCookieName = "auth_token"

// AuthenticatedUserKey should be used to read userID from context.
const AuthenticatedUserKey ctxKey = 0

type (
	// Auth is an authentication middleware.
	//
	// User NewAuth to create a new Auth instance.
	Auth struct {
		cookieName string
		jwt        *jwt.JWT
		store      store.Store
		log        *zap.SugaredLogger
		userID     string
	}

	// Options represents dependencies required for Auth.
	Options struct {
		CookieName string
		JWT        *jwt.JWT
		Store      store.Store
		Log        *zap.SugaredLogger
	}

	ctxKey int
)

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

		a.userID = userID
		a.continueWithUser(w, r, next)
	})
}

// PublicAPIAuth defines authentication handler for private API scope.
func (a *Auth) PrivateAPIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(a.cookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				a.handleUnauthorized(w, errors.New("no authorisation cookie"))
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
		if userID == "" {
			a.handleUnauthorized(w, errors.New("token does not contain user ID"))
			return
		}
		if _, err := a.store.GetUser(r.Context(), userID); err != nil {
			a.handleUnauthorized(w, errors.New("got unregistered user from auth token"))
			return
		}

		a.userID = userID
		a.continueWithUser(w, r, next)
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

func (a *Auth) continueWithUser(w http.ResponseWriter, r *http.Request, next http.Handler) {
	a.log.With("userID", a.userID).Debug("add userID to request context")
	ctx := context.WithValue(r.Context(), AuthenticatedUserKey, a.userID)
	next.ServeHTTP(w, r.WithContext(ctx))
}
