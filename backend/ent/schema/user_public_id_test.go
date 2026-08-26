package schema

import (
	"testing"

	"entgo.io/ent/entc/load"
	"entgo.io/ent/schema/field"
	"github.com/stretchr/testify/require"
)

func TestUserPublicIDIsStableAndNonEnumerable(t *testing.T) {
	spec, err := (&load.Config{Path: "."}).Load()
	require.NoError(t, err)

	schemas := map[string]*load.Schema{}
	for _, loaded := range spec.Schemas {
		schemas[loaded.Name] = loaded
	}
	publicID := requireSchemaField(t, requireSchema(t, schemas, "User"), "public_id")
	require.Equal(t, field.TypeUUID, publicID.Info.Type)
	require.True(t, publicID.Default)
	require.True(t, publicID.Immutable)
	require.True(t, publicID.Unique)
	require.False(t, publicID.Optional)
	require.False(t, publicID.Nillable)
}
