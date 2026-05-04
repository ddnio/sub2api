package repository

import (
	"context"
	"database/sql"
	"testing"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func newUserEmailLookupRepo(t *testing.T) (*userRepository, *dbent.Client) {
	t.Helper()

	db, err := sql.Open("sqlite", "file:user_repo_email_lookup?mode=memory&cache=shared")
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	_, err = db.Exec("PRAGMA foreign_keys = ON")
	require.NoError(t, err)

	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	t.Cleanup(func() { _ = client.Close() })

	return newUserRepositoryWithSQL(client, db), client
}

func TestUserRepositoryGetByEmailNormalizesLegacySpacingAndCase(t *testing.T) {
	repo, _ := newUserEmailLookupRepo(t)
	ctx := context.Background()

	err := repo.Create(ctx, &service.User{
		Email:        " Legacy@Example.com ",
		Username:     "legacy-user",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	})
	require.NoError(t, err)

	got, err := repo.GetByEmail(ctx, "legacy@example.com")
	require.NoError(t, err)
	require.Equal(t, " Legacy@Example.com ", got.Email)
}

func TestUserRepositoryGetByEmailRejectsNormalizedDuplicates(t *testing.T) {
	repo, _ := newUserEmailLookupRepo(t)
	ctx := context.Background()

	err := repo.Create(ctx, &service.User{
		Email:        " Legacy@Example.com ",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	})
	require.NoError(t, err)
	err = repo.Create(ctx, &service.User{
		Email:        "legacy@example.com",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	})
	require.NoError(t, err)

	_, err = repo.GetByEmail(ctx, " legacy@example.com ")
	require.Error(t, err)
	require.Contains(t, err.Error(), "normalized email lookup matched multiple users")
}

func TestUserRepositoryExistsByEmailNormalizesLegacySpacingAndCase(t *testing.T) {
	repo, _ := newUserEmailLookupRepo(t)
	ctx := context.Background()

	err := repo.Create(ctx, &service.User{
		Email:        " Legacy@Example.com ",
		Username:     "legacy-user",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	})
	require.NoError(t, err)

	exists, err := repo.ExistsByEmail(ctx, "  LEGACY@example.com  ")
	require.NoError(t, err)
	require.True(t, exists)
}
