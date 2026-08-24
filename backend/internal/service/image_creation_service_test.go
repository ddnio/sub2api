package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/stretchr/testify/require"
)

type imageCreationRepositoryStub struct {
	created  *ImageCreationTemplate
	homeIDs  []int64
	homeETag string
}

func (s *imageCreationRepositoryStub) StoreAsset(context.Context, *ImageCreationAsset) (*ImageCreationAsset, bool, error) {
	return nil, false, nil
}
func (s *imageCreationRepositoryStub) GetAsset(context.Context, string, bool) (*ImageCreationAsset, error) {
	return nil, nil
}
func (s *imageCreationRepositoryStub) CreateTemplate(_ context.Context, template *ImageCreationTemplate) (*ImageCreationTemplate, error) {
	s.created = template
	template.ID = 7
	return template, nil
}
func (s *imageCreationRepositoryStub) GetTemplate(context.Context, int64, bool) (*ImageCreationTemplate, error) {
	return nil, nil
}
func (s *imageCreationRepositoryStub) ListTemplates(context.Context, int64, pagination.PaginationParams, ImageCreationTemplateListFilters, bool) ([]ImageCreationTemplate, *pagination.PaginationResult, error) {
	return nil, nil, nil
}
func (s *imageCreationRepositoryStub) UpdateDraft(context.Context, int64, int, ImageCreationTemplateDocument, *string, int64) (*ImageCreationTemplate, error) {
	return nil, nil
}
func (s *imageCreationRepositoryStub) PublishTemplate(context.Context, int64, int, int64) (*ImageCreationTemplate, error) {
	return nil, nil
}
func (s *imageCreationRepositoryStub) ArchiveTemplate(context.Context, int64, int64) (*ImageCreationTemplate, error) {
	return nil, nil
}
func (s *imageCreationRepositoryStub) RestoreTemplate(context.Context, int64, int64) (*ImageCreationTemplate, error) {
	return nil, nil
}
func (s *imageCreationRepositoryStub) GetHomeFeatured(context.Context) (*ImageCreationHomeFeatured, error) {
	return nil, nil
}
func (s *imageCreationRepositoryStub) ReplaceHomeFeatured(_ context.Context, etag string, ids []int64, _ int64) (*ImageCreationHomeFeatured, error) {
	s.homeETag = etag
	s.homeIDs = ids
	return &ImageCreationHomeFeatured{TemplateIDs: ids}, nil
}
func (s *imageCreationRepositoryStub) SetFavorite(context.Context, int64, int64, bool) error {
	return nil
}
func (s *imageCreationRepositoryStub) ApplyTemplate(context.Context, int64, int64, int) (*ImageCreationTemplateApplication, error) {
	return nil, nil
}

func TestImageCreationServiceCreatesValidatedDraft(t *testing.T) {
	repo := &imageCreationRepositoryStub{}
	svc := NewImageCreationService(repo)
	doc := validImageCreationDocument()

	created, err := svc.CreateTemplate(context.Background(), doc, nil, 9)
	require.NoError(t, err)
	require.Equal(t, int64(7), created.ID)
	require.Equal(t, ImageCreationTemplateStateDraft, repo.created.State)
	require.Equal(t, 1, repo.created.Revision)
	require.Equal(t, int64(9), repo.created.CreatedBy)

	doc.Title = ""
	_, err = svc.CreateTemplate(context.Background(), doc, nil, 9)
	require.Error(t, err)
}

func TestImageCreationServiceRejectsInvalidHomeFeaturedSet(t *testing.T) {
	repo := &imageCreationRepositoryStub{}
	svc := NewImageCreationService(repo)

	_, err := svc.ReplaceHomeFeatured(context.Background(), "etag", []int64{1, 1}, 9)
	require.Error(t, err)
	_, err = svc.ReplaceHomeFeatured(context.Background(), "etag", []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 9)
	require.NoError(t, err)
	require.Len(t, repo.homeIDs, 20)

	_, err = svc.ReplaceHomeFeatured(context.Background(), "etag", []int64{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21}, 9)
	require.Error(t, err)

	_, err = svc.ReplaceHomeFeatured(context.Background(), "etag", []int64{3, 7}, 9)
	require.NoError(t, err)
	require.Equal(t, []int64{3, 7}, repo.homeIDs)
}

func TestImageCreationServiceNormalizesCompressedHomeFeaturedETag(t *testing.T) {
	repo := &imageCreationRepositoryStub{}
	svc := NewImageCreationService(repo)

	_, err := svc.ReplaceHomeFeatured(context.Background(), `"version-zstd"`, []int64{3, 7}, 9)
	require.NoError(t, err)
	require.Equal(t, `"version"`, repo.homeETag)
}
