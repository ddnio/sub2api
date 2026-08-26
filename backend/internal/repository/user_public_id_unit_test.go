package repository

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUserRepositoryCreatesStableUniquePublicIDs(t *testing.T) {
	repo, _ := newUserEntRepo(t)
	ctx := context.Background()

	first := &service.User{
		Email:        "first-public-id@example.com",
		Username:     "first-public-id",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, first))
	require.NotEmpty(t, first.PublicID)
	_, err := uuid.Parse(first.PublicID)
	require.NoError(t, err)

	loaded, err := repo.GetByID(ctx, first.ID)
	require.NoError(t, err)
	require.Equal(t, first.PublicID, loaded.PublicID)

	second := &service.User{
		Email:        "second-public-id@example.com",
		Username:     "second-public-id",
		PasswordHash: "hash",
		Role:         service.RoleUser,
		Status:       service.StatusActive,
	}
	require.NoError(t, repo.Create(ctx, second))
	require.NotEqual(t, first.PublicID, second.PublicID)
}
