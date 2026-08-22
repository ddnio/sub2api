//go:build integration

package repository

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestImageCreationRepositoryLifecycle(t *testing.T) {
	ctx := context.Background()
	client := testEntClient(t)
	user := mustCreateUser(t, client, &service.User{Email: fmt.Sprintf("image-creation-%d@example.com", time.Now().UnixNano())})
	t.Cleanup(func() {
		_, _ = integrationDB.ExecContext(context.Background(), "DELETE FROM image_creation_user_template_states WHERE user_id = $1", user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), "DELETE FROM image_creation_change_logs WHERE actor_user_id = $1", user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), "DELETE FROM image_creation_templates WHERE created_by = $1", user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), "DELETE FROM image_creation_assets WHERE created_by = $1", user.ID)
		_, _ = integrationDB.ExecContext(context.Background(), "DELETE FROM users WHERE id = $1", user.ID)
	})

	var imageData bytes.Buffer
	imageFile := image.NewRGBA(image.Rect(0, 0, 2, 2))
	imageFile.Set(0, 0, color.RGBA{R: 20, G: 160, B: 140, A: 255})
	require.NoError(t, png.Encode(&imageData, imageFile))
	svc := service.NewImageCreationService(NewImageCreationRepository(client))
	asset, created, err := svc.UploadAsset(ctx, imageData.Bytes(), "uploaded", "", "", user.ID)
	require.NoError(t, err)
	require.True(t, created)

	title := fmt.Sprintf("portrait-%d", time.Now().UnixNano())
	template, err := svc.CreateTemplate(ctx, service.ImageCreationTemplateDocument{
		SchemaVersion: 1,
		Title:         title,
		Summary:       "studio portrait",
		Category:      "portrait",
		Tags:          []string{"portrait", "studio"},
		Prompt:        "soft window light portrait",
		InputMode:     "text",
		CoverAlt:      "studio portrait sample",
		Defaults: domain.ImageCreationTemplateDefaults{
			Size: "1024x1024", Quality: "high", OutputFormat: "png",
		},
	}, &asset.ID, user.ID)
	require.NoError(t, err)

	params := pagination.PaginationParams{Page: 1, PageSize: 20}
	adminItems, _, err := svc.ListTemplates(ctx, 0, params, service.ImageCreationTemplateListFilters{Query: title}, true)
	require.NoError(t, err)
	require.Len(t, adminItems, 1)

	template, err = svc.PublishTemplate(ctx, template.ID, template.Revision, user.ID)
	require.NoError(t, err)
	require.Equal(t, 1, template.PublishedVersion)

	require.NoError(t, svc.SetFavorite(ctx, user.ID, template.ID, true))
	favorites, _, err := svc.ListTemplates(ctx, user.ID, params, service.ImageCreationTemplateListFilters{Favorite: true}, false)
	require.NoError(t, err)
	require.Len(t, favorites, 1)

	applied, err := svc.ApplyTemplate(ctx, user.ID, template.ID, template.PublishedVersion)
	require.NoError(t, err)
	require.Equal(t, "soft window light portrait", applied.Prompt)
	recent, _, err := svc.ListTemplates(ctx, user.ID, params, service.ImageCreationTemplateListFilters{Recent: true}, false)
	require.NoError(t, err)
	require.Len(t, recent, 1)

	home, err := svc.GetHomeFeatured(ctx)
	require.NoError(t, err)
	home, err = svc.ReplaceHomeFeatured(ctx, home.ETag, []int64{template.ID}, user.ID)
	require.NoError(t, err)
	require.Equal(t, []int64{template.ID}, home.TemplateIDs)
}
