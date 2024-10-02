package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/madatsci/urlshortener/pkg/jwt"
	"go.uber.org/zap"
)

const defaultCookieName = "auth_token"

// AuthenticatedUserKey should be used to read userID from context.
const AuthenticatedUserKey ctxKey = 0

type (
	Auth struct {
		cookieName string
		jwt        *jwt.JWT
		log        *zap.SugaredLogger
		userID     string
	}

	Options struct {
		CookieName string
		JWT        *jwt.JWT
		Log        *zap.SugaredLogger
	}

	ctxKey int
)

func NewAuth(opts Options) *Auth {
	cookieName := opts.CookieName
	if cookieName == "" {
		cookieName = defaultCookieName
	}

	return &Auth{
		cookieName: cookieName,
		jwt:        opts.JWT,
		log:        opts.Log,
	}
}

func (a *Auth) PublicAPIAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var userID string
		var err error

		cookie, err := r.Cookie(a.cookieName)
		if err != nil {
			if err == http.ErrNoCookie {
				a.log.Debug("cookie header not found, issue new token")
				userID, err = a.registerNewUser(w)
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
				userID, err = a.registerNewUser(w)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		a.userID = userID
		a.continueWithUser(w, r, next)
	})
}

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

		a.userID = userID
		a.continueWithUser(w, r, next)
	})
}

func (a *Auth) registerNewUser(w http.ResponseWriter) (string, error) {
	userID := uuid.NewString()
	token, err := a.jwt.GetString(userID)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{Name: a.cookieName, Value: token})

	a.log.With("userID", userID).Debug("registered new user")

	return userID, nil
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
