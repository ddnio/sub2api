package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageCreationFoundationMigration(t *testing.T) {
	content, err := FS.ReadFile("238_image_creation_foundation.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	for _, table := range []string{
		"image_creation_assets",
		"image_creation_templates",
		"image_creation_user_template_states",
		"image_creation_change_logs",
	} {
		require.Contains(t, sql, "CREATE TABLE IF NOT EXISTS "+table)
	}

	require.Contains(t, sql, "content BYTEA NOT NULL")
	require.Contains(t, sql, "CHECK (octet_length(content) = byte_size)")
	require.Contains(t, sql, "CHECK (home_position IS NULL OR state = 'published')")
	require.Contains(t, sql, "CREATE UNIQUE INDEX IF NOT EXISTS image_creation_templates_home_position_unique")
	require.Contains(t, sql, "WHERE home_position IS NOT NULL")
	require.Contains(t, sql, "PRIMARY KEY (user_id, template_id)")
	require.Contains(t, sql, "ON DELETE CASCADE")
	require.NotContains(t, sql, "REFERENCES settings")
}
