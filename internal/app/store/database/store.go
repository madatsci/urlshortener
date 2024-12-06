package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/madatsci/urlshortener/internal/app/models"
	"github.com/madatsci/urlshortener/internal/app/store"
	"github.com/pressly/goose/v3"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

// Store is an implementation of store.Store interface which interacts with database.
type Store struct {
	conn *sql.DB
}

// New creates a new database-driven storage.
func New(ctx context.Context, conn *sql.DB) (*Store, error) {
	store := &Store{conn: conn}
	if err := store.bootstrap(); err != nil {
		return nil, err
	}

	return store, nil
}

func (s *Store) CreateUser(ctx context.Context, user models.User) error {
	_, err := s.conn.ExecContext(
		ctx,
		"INSERT INTO users (id, created_at) VALUES ($1, $2)",
		user.ID,
		user.CreatedAt,
	)

	return err
}

func (s *Store) GetUser(ctx context.Context, userID string) (models.User, error) {
	var user models.User

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, created_at FROM users WHERE id = $1",
		userID,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		return user, err
	}

	return user, nil
}

func (s *Store) CreateURL(ctx context.Context, url models.URL) error {
	originalURL, err := s.getByOriginalURL(ctx, url.Original)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			_, err = s.conn.ExecContext(
				ctx,
				"INSERT INTO urls (id, user_id, correlation_id, slug, original_url, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
				url.ID,
				newNullString(url.UserID),
				url.CorrelationID,
				url.Slug,
				url.Original,
				url.CreatedAt,
			)
			if err != nil {
				return err
			}

			return s.linkURLtoUser(ctx, url, url.UserID)
		}

		return err
	}

	if err = s.linkURLtoUser(ctx, originalURL, url.UserID); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			return &store.AlreadyExistsError{
				Err: pgErr,
				URL: originalURL,
			}
		}
	}

	return err
}

func (s *Store) BatchCreateURL(ctx context.Context, urls []models.URL) error {
	tx, err := s.conn.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback() //nolint:errcheck

	urlStmt, err := tx.PrepareContext(
		ctx,
		"INSERT INTO urls (id, user_id, correlation_id, slug, original_url, created_at) VALUES ($1, $2, $3, $4, $5, $6)",
	)
	if err != nil {
		return err
	}
	defer urlStmt.Close()

	userUrlStmt, err := tx.PrepareContext(
		ctx,
		"INSERT INTO user_urls (id, user_id, url_id, is_deleted, created_at) VALUES ($1, $2, $3, $4, $5)",
	)
	if err != nil {
		return err
	}
	defer userUrlStmt.Close()

	for _, url := range urls {
		_, err := urlStmt.ExecContext(ctx, url.ID, newNullString(url.UserID), url.CorrelationID, url.Slug, url.Original, url.CreatedAt)
		if err != nil {
			return err
		}

		_, err = userUrlStmt.ExecContext(ctx, uuid.NewString(), url.UserID, url.ID, false, time.Now())
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) GetURL(ctx context.Context, slug string) (models.URL, error) {
	var url models.URL
	var userID sql.NullString

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, user_id, correlation_id, slug, original_url, created_at, is_deleted FROM urls WHERE slug = $1",
		slug,
	).Scan(&url.ID, &userID, &url.CorrelationID, &url.Slug, &url.Original, &url.CreatedAt, &url.Deleted)

	if err != nil {
		return url, err
	}

	url.UserID = userID.String

	return url, nil
}

func (s *Store) ListURLsByUserID(ctx context.Context, userID string) ([]models.URL, error) {
	res := make([]models.URL, 0)

	rows, err := s.conn.QueryContext(
		ctx,
		"SELECT id, user_id, correlation_id, slug, original_url, created_at, is_deleted FROM urls WHERE id IN (SELECT url_id FROM user_urls WHERE user_id = $1 AND NOT is_deleted)",
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var url models.URL
		var userID sql.NullString
		err = rows.Scan(&url.ID, &userID, &url.CorrelationID, &url.Slug, &url.Original, &url.CreatedAt, &url.Deleted)
		if err != nil {
			return nil, err
		}
		url.UserID = userID.String
		res = append(res, url)
	}

	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (s *Store) ListAllUrls(ctx context.Context) map[string]models.URL {
	// TODO implement later (currently this is used only for testing purposes)
	return nil
}

func (s *Store) SoftDeleteURL(ctx context.Context, userID string, slugs []string) error {
	inArgs := make([]string, len(slugs))
	for i := 0; i < len(slugs); i++ {
		inArgs[i] = fmt.Sprintf("$%d", i+2)
	}
	inArg := strings.Join(inArgs, ",")

	args := make([]interface{}, len(slugs)+1)
	args[0] = userID
	for i, slug := range slugs {
		args[i+1] = slug
	}

	_, err := s.conn.ExecContext(
		ctx,
		"UPDATE urls SET is_deleted = true WHERE user_id = $1 and slug IN("+inArg+")",
		args...,
	)

	return err
}

func (s *Store) Ping(ctx context.Context) error {
	return s.conn.PingContext(ctx)
}

func (s *Store) bootstrap() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	if err := goose.Up(s.conn, "migrations"); err != nil {
		return err
	}

	return nil
}

func (s *Store) getByOriginalURL(ctx context.Context, originalURL string) (models.URL, error) {
	var url models.URL
	var userID sql.NullString

	err := s.conn.QueryRowContext(
		ctx,
		"SELECT id, user_id, correlation_id, slug, original_url, created_at, is_deleted FROM urls WHERE original_url = $1",
		originalURL,
	).Scan(&url.ID, &userID, &url.CorrelationID, &url.Slug, &url.Original, &url.CreatedAt, &url.Deleted)

	if err != nil {
		return url, err
	}

	url.UserID = userID.String

	return url, nil
}

func (s *Store) linkURLtoUser(ctx context.Context, url models.URL, userID string) error {
	userURL := models.UserURL{
		ID:        uuid.NewString(),
		UserID:    userID,
		UrlID:     url.ID,
		Deleted:   false,
		CreatedAt: time.Now(),
	}

	_, err := s.conn.ExecContext(
		ctx,
		"INSERT INTO user_urls (id, user_id, url_id, is_deleted, created_at) VALUES ($1, $2, $3, $4, $5)",
		userURL.ID,
		userURL.UserID,
		userURL.UrlID,
		userURL.Deleted,
		userURL.CreatedAt,
	)

	return err
}

func newNullString(s string) sql.NullString {
	if len(s) == 0 {
		return sql.NullString{}
	}
	return sql.NullString{
		String: s,
		Valid:  true,
	}
}
