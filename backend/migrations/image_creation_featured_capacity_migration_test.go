package migrations

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageCreationFeaturedCapacityMigration(t *testing.T) {
	content, err := FS.ReadFile("239_expand_image_creation_featured_capacity.sql")
	require.NoError(t, err)

	sql := strings.Join(strings.Fields(string(content)), " ")
	require.Contains(t, sql, "DROP CONSTRAINT IF EXISTS image_creation_templates_home_position_check")
	require.Contains(t, sql, "CHECK (home_position BETWEEN 1 AND 20)")
}
