package database

import (
	"context"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewClient(t *testing.T) {
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		t.Skip("database DSN not set")
	}

	db, err := NewClient(context.Background(), dsn)
	require.NoError(t, err)

	var message string
	err = db.QueryRow("SELECT 'hello'").Scan(&message)
	require.NoError(t, err)
	assert.Equal(t, "hello", message)
}
