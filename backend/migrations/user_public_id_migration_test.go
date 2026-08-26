package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUserPublicIDMigrationBackfillsWithoutExposingSequentialIDs(t *testing.T) {
	content, err := FS.ReadFile("240_add_user_public_id.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "ADD COLUMN IF NOT EXISTS public_id UUID")
	require.Contains(t, sql, "UPDATE users SET public_id = gen_random_uuid() WHERE public_id IS NULL")
	require.Contains(t, sql, "ALTER COLUMN public_id SET DEFAULT gen_random_uuid()")
	require.Contains(t, sql, "ALTER COLUMN public_id SET NOT NULL")
	require.NotContains(t, sql, "id::text")
	require.NotContains(t, sql, "email")
}

func TestUserPublicIDUniqueIndexIsCreatedOnline(t *testing.T) {
	content, err := FS.ReadFile("241_add_user_public_id_unique_index_notx.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "CREATE UNIQUE INDEX CONCURRENTLY IF NOT EXISTS users_public_id_unique")
	require.Contains(t, sql, "ON users (public_id)")
	require.NotContains(t, sql, "DROP INDEX")
}
