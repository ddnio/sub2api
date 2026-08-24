package service

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
)

const (
	ImageCreationTemplateStateDraft     = "draft"
	ImageCreationTemplateStatePublished = "published"
	ImageCreationTemplateStateArchived  = "archived"
)

var (
	ErrImageCreationNotFound      = infraerrors.NotFound("IMAGE_CREATION_NOT_FOUND", "image creation resource not found")
	ErrImageCreationConflict      = infraerrors.New(http.StatusConflict, "IMAGE_CREATION_CONFLICT", "image creation resource changed; reload and try again")
	ErrImageCreationCoverRequired = infraerrors.BadRequest("IMAGE_CREATION_INVALID_INPUT", "template cover is required before publish")
	ErrImageCreationAssetTooLarge = infraerrors.New(http.StatusRequestEntityTooLarge, "IMAGE_CREATION_ASSET_TOO_LARGE", "image must not exceed 8 MiB")
	imageCreationAssetIDPattern   = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

type ImageCreationTemplateDocument = domain.ImageCreationTemplateDocument

type ImageCreationTemplate struct {
	ID                    int64
	State                 string
	DraftData             ImageCreationTemplateDocument
	PublishedData         *ImageCreationTemplateDocument
	Revision              int
	PublishedVersion      int
	DraftCoverAssetID     *string
	PublishedCoverAssetID *string
	HomePosition          *int16
	CreatedBy             int64
	UpdatedBy             int64
	CreatedAt             time.Time
	UpdatedAt             time.Time
	PublishedAt           *time.Time
	FavoritedAt           *time.Time
	LastUsedAt            *time.Time
}

type ImageCreationTemplateListFilters struct {
	Query    string
	Category string
	Tag      string
	State    string
	Favorite bool
	Recent   bool
	Home     bool
}

type ImageCreationHomeFeatured struct {
	ETag        string
	TemplateIDs []int64
	Templates   []ImageCreationTemplate
}

type ImageCreationTemplateApplication struct {
	TemplateID       int64
	PublishedVersion int
	Prompt           string
	Defaults         domain.ImageCreationTemplateDefaults
	InputMode        string
}

type ImageCreationRepository interface {
	StoreAsset(ctx context.Context, asset *ImageCreationAsset) (*ImageCreationAsset, bool, error)
	GetAsset(ctx context.Context, id string, withContent bool) (*ImageCreationAsset, error)
	CreateTemplate(ctx context.Context, template *ImageCreationTemplate) (*ImageCreationTemplate, error)
	GetTemplate(ctx context.Context, id int64, publishedOnly bool) (*ImageCreationTemplate, error)
	ListTemplates(ctx context.Context, userID int64, params pagination.PaginationParams, filters ImageCreationTemplateListFilters, admin bool) ([]ImageCreationTemplate, *pagination.PaginationResult, error)
	UpdateDraft(ctx context.Context, id int64, revision int, doc ImageCreationTemplateDocument, coverAssetID *string, actorID int64) (*ImageCreationTemplate, error)
	PublishTemplate(ctx context.Context, id int64, revision int, actorID int64) (*ImageCreationTemplate, error)
	ArchiveTemplate(ctx context.Context, id, actorID int64) (*ImageCreationTemplate, error)
	RestoreTemplate(ctx context.Context, id, actorID int64) (*ImageCreationTemplate, error)
	GetHomeFeatured(ctx context.Context) (*ImageCreationHomeFeatured, error)
	ReplaceHomeFeatured(ctx context.Context, etag string, templateIDs []int64, actorID int64) (*ImageCreationHomeFeatured, error)
	SetFavorite(ctx context.Context, userID, templateID int64, favorite bool) error
	ApplyTemplate(ctx context.Context, userID, templateID int64, publishedVersion int) (*ImageCreationTemplateApplication, error)
}

type ImageCreationService struct {
	repo ImageCreationRepository
}

const maxHomeFeaturedTemplates = 20

func NewImageCreationService(repo ImageCreationRepository) *ImageCreationService {
	return &ImageCreationService{repo: repo}
}

func (s *ImageCreationService) UploadAsset(ctx context.Context, content []byte, sourceType, sourceProvider, sourceModel string, actorID int64) (*ImageCreationAsset, bool, error) {
	asset, err := ValidateImageCreationAsset(content, sourceType, sourceProvider, sourceModel, actorID)
	if err != nil {
		return nil, false, err
	}
	return s.repo.StoreAsset(ctx, asset)
}

func (s *ImageCreationService) GetAsset(ctx context.Context, id string, withContent bool) (*ImageCreationAsset, error) {
	if !imageCreationAssetIDPattern.MatchString(id) {
		return nil, ErrImageCreationNotFound
	}
	return s.repo.GetAsset(ctx, id, withContent)
}

func (s *ImageCreationService) CreateTemplate(ctx context.Context, doc ImageCreationTemplateDocument, coverAssetID *string, actorID int64) (*ImageCreationTemplate, error) {
	if err := validateImageCreationWrite(doc, coverAssetID, actorID); err != nil {
		return nil, err
	}
	return s.repo.CreateTemplate(ctx, &ImageCreationTemplate{
		State: ImageCreationTemplateStateDraft, DraftData: doc, Revision: 1,
		DraftCoverAssetID: coverAssetID, CreatedBy: actorID, UpdatedBy: actorID,
	})
}

func (s *ImageCreationService) GetTemplate(ctx context.Context, id int64, publishedOnly bool) (*ImageCreationTemplate, error) {
	if id <= 0 {
		return nil, ErrImageCreationNotFound
	}
	return s.repo.GetTemplate(ctx, id, publishedOnly)
}

func (s *ImageCreationService) ListTemplates(ctx context.Context, userID int64, params pagination.PaginationParams, filters ImageCreationTemplateListFilters, admin bool) ([]ImageCreationTemplate, *pagination.PaginationResult, error) {
	filters.Query = strings.TrimSpace(filters.Query)
	filters.Category = strings.TrimSpace(filters.Category)
	filters.Tag = strings.TrimSpace(filters.Tag)
	filters.State = strings.TrimSpace(filters.State)
	if len(filters.Query) > 200 || len(filters.Category) > 64 || len(filters.Tag) > 40 {
		return nil, nil, invalidImageCreationInput("template filters are invalid")
	}
	if !admin && userID <= 0 {
		return nil, nil, ErrImageCreationSessionInvalid
	}
	return s.repo.ListTemplates(ctx, userID, params, filters, admin)
}

func (s *ImageCreationService) UpdateDraft(ctx context.Context, id int64, revision int, doc ImageCreationTemplateDocument, coverAssetID *string, actorID int64) (*ImageCreationTemplate, error) {
	if id <= 0 || revision <= 0 {
		return nil, invalidImageCreationInput("template id or revision is invalid")
	}
	if err := validateImageCreationWrite(doc, coverAssetID, actorID); err != nil {
		return nil, err
	}
	return s.repo.UpdateDraft(ctx, id, revision, doc, coverAssetID, actorID)
}

func (s *ImageCreationService) PublishTemplate(ctx context.Context, id int64, revision int, actorID int64) (*ImageCreationTemplate, error) {
	if id <= 0 || revision <= 0 || actorID <= 0 {
		return nil, invalidImageCreationInput("template id, revision, or actor is invalid")
	}
	return s.repo.PublishTemplate(ctx, id, revision, actorID)
}

func (s *ImageCreationService) ArchiveTemplate(ctx context.Context, id, actorID int64) (*ImageCreationTemplate, error) {
	if id <= 0 || actorID <= 0 {
		return nil, invalidImageCreationInput("template id or actor is invalid")
	}
	return s.repo.ArchiveTemplate(ctx, id, actorID)
}

func (s *ImageCreationService) RestoreTemplate(ctx context.Context, id, actorID int64) (*ImageCreationTemplate, error) {
	if id <= 0 || actorID <= 0 {
		return nil, invalidImageCreationInput("template id or actor is invalid")
	}
	return s.repo.RestoreTemplate(ctx, id, actorID)
}

func (s *ImageCreationService) GetHomeFeatured(ctx context.Context) (*ImageCreationHomeFeatured, error) {
	return s.repo.GetHomeFeatured(ctx)
}

func (s *ImageCreationService) ReplaceHomeFeatured(ctx context.Context, etag string, templateIDs []int64, actorID int64) (*ImageCreationHomeFeatured, error) {
	if strings.TrimSpace(etag) == "" || actorID <= 0 || len(templateIDs) > maxHomeFeaturedTemplates {
		return nil, invalidImageCreationInput("home featured input is invalid")
	}
	seen := make(map[int64]bool, len(templateIDs))
	for _, id := range templateIDs {
		if id <= 0 || seen[id] {
			return nil, invalidImageCreationInput("home featured template ids must be positive and unique")
		}
		seen[id] = true
	}
	return s.repo.ReplaceHomeFeatured(ctx, etag, templateIDs, actorID)
}

func (s *ImageCreationService) SetFavorite(ctx context.Context, userID, templateID int64, favorite bool) error {
	if userID <= 0 || templateID <= 0 {
		return invalidImageCreationInput("user or template id is invalid")
	}
	return s.repo.SetFavorite(ctx, userID, templateID, favorite)
}

func (s *ImageCreationService) ApplyTemplate(ctx context.Context, userID, templateID int64, publishedVersion int) (*ImageCreationTemplateApplication, error) {
	if userID <= 0 || templateID <= 0 || publishedVersion <= 0 {
		return nil, invalidImageCreationInput("template application input is invalid")
	}
	return s.repo.ApplyTemplate(ctx, userID, templateID, publishedVersion)
}

func validateImageCreationWrite(doc ImageCreationTemplateDocument, coverAssetID *string, actorID int64) error {
	if actorID <= 0 {
		return invalidImageCreationInput("actor is invalid")
	}
	if coverAssetID != nil && !imageCreationAssetIDPattern.MatchString(*coverAssetID) {
		return invalidImageCreationInput("cover_asset_id is invalid")
	}
	return ValidateImageCreationTemplateDocument(doc)
}
