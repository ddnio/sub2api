package handler

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type ImageCreationHandler struct {
	service *service.ImageCreationService
}

type imageCreationTemplateWriteRequest struct {
	Document     service.ImageCreationTemplateDocument `json:"document"`
	CoverAssetID *string                               `json:"cover_asset_id"`
}

type imageCreationApplyRequest struct {
	PublishedVersion int `json:"published_version"`
}

type imageCreationHomeRequest struct {
	TemplateIDs []int64 `json:"template_ids"`
}

type imageCreationTemplateListDTO struct {
	ID               int64                                `json:"id"`
	Title            string                               `json:"title"`
	Summary          string                               `json:"summary"`
	Category         string                               `json:"category"`
	Tags             []string                             `json:"tags"`
	CoverAssetID     *string                              `json:"cover_asset_id,omitempty"`
	PublishedVersion int                                  `json:"published_version"`
	Defaults         domain.ImageCreationTemplateDefaults `json:"defaults"`
	InputMode        string                               `json:"input_mode"`
	HomePosition     *int16                               `json:"home_position,omitempty"`
	Favorited        bool                                 `json:"favorited"`
}

type imageCreationTemplateDetailDTO struct {
	imageCreationTemplateListDTO
	Prompt   string                                         `json:"prompt"`
	CoverAlt string                                         `json:"cover_alt"`
	Source   *domain.ImageCreationTemplateSourceAttribution `json:"source,omitempty"`
}

type imageCreationAdminTemplateDTO struct {
	ID                    int64                                  `json:"id"`
	State                 string                                 `json:"state"`
	DraftData             service.ImageCreationTemplateDocument  `json:"draft_data"`
	PublishedData         *service.ImageCreationTemplateDocument `json:"published_data,omitempty"`
	Revision              int                                    `json:"revision"`
	PublishedVersion      int                                    `json:"published_version"`
	DraftCoverAssetID     *string                                `json:"draft_cover_asset_id,omitempty"`
	PublishedCoverAssetID *string                                `json:"published_cover_asset_id,omitempty"`
	HomePosition          *int16                                 `json:"home_position,omitempty"`
	CreatedBy             int64                                  `json:"created_by"`
	UpdatedBy             int64                                  `json:"updated_by"`
	CreatedAt             time.Time                              `json:"created_at"`
	UpdatedAt             time.Time                              `json:"updated_at"`
	PublishedAt           *time.Time                             `json:"published_at,omitempty"`
}

func NewImageCreationHandler(imageCreationService *service.ImageCreationService) *ImageCreationHandler {
	return &ImageCreationHandler{service: imageCreationService}
}

func (h *ImageCreationHandler) List(c *gin.Context) {
	userID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListTemplates(c.Request.Context(), userID, pagination.PaginationParams{
		Page: page, PageSize: pageSize,
	}, service.ImageCreationTemplateListFilters{
		Query: c.Query("q"), Category: c.Query("category"), Tag: c.Query("tag"),
		Favorite: parseBoolQuery(c.Query("favorite")), Recent: parseBoolQuery(c.Query("recent")), Home: parseBoolQuery(c.Query("home")),
	}, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]imageCreationTemplateListDTO, 0, len(items))
	for i := range items {
		out = append(out, imageCreationUserTemplateListDTO(&items[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

func (h *ImageCreationHandler) Get(c *gin.Context) {
	if _, ok := imageCreationUserID(c); !ok {
		return
	}
	id, ok := imageCreationTemplateID(c)
	if !ok {
		return
	}
	template, err := h.service.GetTemplate(c.Request.Context(), id, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreationUserTemplateDetailDTO(template))
}

func (h *ImageCreationHandler) Favorite(c *gin.Context) {
	userID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	id, ok := imageCreationTemplateID(c)
	if !ok {
		return
	}
	if err := h.service.SetFavorite(c.Request.Context(), userID, id, true); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"favorited": true})
}

func (h *ImageCreationHandler) Unfavorite(c *gin.Context) {
	userID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	id, ok := imageCreationTemplateID(c)
	if !ok {
		return
	}
	if err := h.service.SetFavorite(c.Request.Context(), userID, id, false); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"favorited": false})
}

func (h *ImageCreationHandler) Apply(c *gin.Context) {
	userID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	id, ok := imageCreationTemplateID(c)
	if !ok {
		return
	}
	var req imageCreationApplyRequest
	if err := decodeImageCreationJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid image creation request")
		return
	}
	application, err := h.service.ApplyTemplate(c.Request.Context(), userID, id, req.PublishedVersion)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, application)
}

func (h *ImageCreationHandler) AssetContent(c *gin.Context) {
	asset, err := h.service.GetAsset(c.Request.Context(), strings.TrimSpace(c.Param("id")), true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("ETag", `"`+asset.ID+`"`)
	c.Header("Cache-Control", "public, max-age=31536000, immutable")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Length", strconv.Itoa(asset.ByteSize))
	c.Data(http.StatusOK, asset.ContentType, asset.Content)
}

func (h *ImageCreationHandler) AdminList(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	items, result, err := h.service.ListTemplates(c.Request.Context(), 0, pagination.PaginationParams{Page: page, PageSize: pageSize}, service.ImageCreationTemplateListFilters{
		Query: c.Query("q"), State: c.Query("state"),
	}, true)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	out := make([]imageCreationAdminTemplateDTO, 0, len(items))
	for i := range items {
		out = append(out, imageCreationAdminTemplateDTOFromService(&items[i]))
	}
	response.Paginated(c, out, result.Total, page, pageSize)
}

func (h *ImageCreationHandler) AdminGet(c *gin.Context) {
	id, ok := imageCreationTemplateID(c)
	if !ok {
		return
	}
	template, err := h.service.GetTemplate(c.Request.Context(), id, false)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreationAdminTemplateDTOFromService(template))
}

func (h *ImageCreationHandler) AdminCreate(c *gin.Context) {
	actorID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	var req imageCreationTemplateWriteRequest
	if err := decodeImageCreationJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid image creation request")
		return
	}
	template, err := h.service.CreateTemplate(c.Request.Context(), req.Document, req.CoverAssetID, actorID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreationAdminTemplateDTOFromService(template))
}

func (h *ImageCreationHandler) AdminUpdateDraft(c *gin.Context) {
	actorID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	id, ok := imageCreationTemplateID(c)
	if !ok {
		return
	}
	revision, ok := imageCreationRevision(c)
	if !ok {
		return
	}
	var req imageCreationTemplateWriteRequest
	if err := decodeImageCreationJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid image creation request")
		return
	}
	template, err := h.service.UpdateDraft(c.Request.Context(), id, revision, req.Document, req.CoverAssetID, actorID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreationAdminTemplateDTOFromService(template))
}

func (h *ImageCreationHandler) AdminPublish(c *gin.Context) {
	h.adminStateChange(c, func(ctxID, actorID int64, revision int) (*service.ImageCreationTemplate, error) {
		return h.service.PublishTemplate(c.Request.Context(), ctxID, revision, actorID)
	}, true)
}

func (h *ImageCreationHandler) AdminArchive(c *gin.Context) {
	h.adminStateChange(c, func(ctxID, actorID int64, _ int) (*service.ImageCreationTemplate, error) {
		return h.service.ArchiveTemplate(c.Request.Context(), ctxID, actorID)
	}, false)
}

func (h *ImageCreationHandler) AdminRestore(c *gin.Context) {
	h.adminStateChange(c, func(ctxID, actorID int64, _ int) (*service.ImageCreationTemplate, error) {
		return h.service.RestoreTemplate(c.Request.Context(), ctxID, actorID)
	}, false)
}

func (h *ImageCreationHandler) adminStateChange(c *gin.Context, change func(int64, int64, int) (*service.ImageCreationTemplate, error), needsRevision bool) {
	actorID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	id, ok := imageCreationTemplateID(c)
	if !ok {
		return
	}
	revision := 0
	if needsRevision {
		revision, ok = imageCreationRevision(c)
		if !ok {
			return
		}
	}
	template, err := change(id, actorID, revision)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, imageCreationAdminTemplateDTOFromService(template))
}

func (h *ImageCreationHandler) AdminUploadAsset(c *gin.Context) {
	actorID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, service.ImageCreationAssetMaxBytes+1024*1024)
	file, err := c.FormFile("file")
	if err != nil {
		if _, ok := extractMaxBytesError(err); ok {
			response.ErrorFrom(c, service.ErrImageCreationAssetTooLarge)
			return
		}
		response.BadRequest(c, "Image file is required")
		return
	}
	opened, err := file.Open()
	if err != nil {
		response.BadRequest(c, "Image file is invalid")
		return
	}
	defer opened.Close()
	content, err := io.ReadAll(io.LimitReader(opened, service.ImageCreationAssetMaxBytes+1))
	if err != nil {
		response.BadRequest(c, "Image file is invalid")
		return
	}
	asset, created, err := h.service.UploadAsset(c.Request.Context(), content, c.DefaultPostForm("source_type", "uploaded"), c.PostForm("source_provider"), c.PostForm("source_model"), actorID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{
		"id": asset.ID, "content_type": asset.ContentType, "byte_size": asset.ByteSize,
		"width": asset.Width, "height": asset.Height, "created": created,
	})
}

func (h *ImageCreationHandler) AdminGetHome(c *gin.Context) {
	home, err := h.service.GetHomeFeatured(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("ETag", home.ETag)
	response.Success(c, gin.H{"template_ids": home.TemplateIDs, "templates": imageCreationAdminTemplateDTOs(home.Templates)})
}

func (h *ImageCreationHandler) AdminReplaceHome(c *gin.Context) {
	actorID, ok := imageCreationUserID(c)
	if !ok {
		return
	}
	var req imageCreationHomeRequest
	if err := decodeImageCreationJSON(c, &req); err != nil {
		response.BadRequest(c, "Invalid image creation request")
		return
	}
	home, err := h.service.ReplaceHomeFeatured(c.Request.Context(), c.GetHeader("If-Match"), req.TemplateIDs, actorID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	c.Header("ETag", home.ETag)
	response.Success(c, gin.H{"template_ids": home.TemplateIDs, "templates": imageCreationAdminTemplateDTOs(home.Templates)})
}

func decodeImageCreationJSON(c *gin.Context, target any) error {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1024*1024)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func imageCreationUserID(c *gin.Context) (int64, bool) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "Image creation session is required")
		return 0, false
	}
	return subject.UserID, true
}

func imageCreationTemplateID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid image creation template ID")
		return 0, false
	}
	return id, true
}

func imageCreationRevision(c *gin.Context) (int, bool) {
	revision, err := strconv.Atoi(strings.Trim(strings.TrimSpace(c.GetHeader("If-Match")), `"`))
	if err != nil || revision <= 0 {
		response.BadRequest(c, "Valid If-Match revision is required")
		return 0, false
	}
	return revision, true
}

func imageCreationUserTemplateListDTO(template *service.ImageCreationTemplate) imageCreationTemplateListDTO {
	doc := template.PublishedData
	if doc == nil {
		doc = &template.DraftData
	}
	return imageCreationTemplateListDTO{
		ID: template.ID, Title: doc.Title, Summary: doc.Summary, Category: doc.Category, Tags: doc.Tags,
		CoverAssetID: template.PublishedCoverAssetID, PublishedVersion: template.PublishedVersion,
		Defaults: doc.Defaults, InputMode: doc.InputMode, HomePosition: template.HomePosition, Favorited: template.FavoritedAt != nil,
	}
}

func imageCreationUserTemplateDetailDTO(template *service.ImageCreationTemplate) imageCreationTemplateDetailDTO {
	doc := template.PublishedData
	if doc == nil {
		doc = &template.DraftData
	}
	return imageCreationTemplateDetailDTO{
		imageCreationTemplateListDTO: imageCreationUserTemplateListDTO(template),
		Prompt:                       doc.Prompt, CoverAlt: doc.CoverAlt, Source: doc.Source,
	}
}

func imageCreationAdminTemplateDTOFromService(template *service.ImageCreationTemplate) imageCreationAdminTemplateDTO {
	return imageCreationAdminTemplateDTO{
		ID: template.ID, State: template.State, DraftData: template.DraftData, PublishedData: template.PublishedData,
		Revision: template.Revision, PublishedVersion: template.PublishedVersion,
		DraftCoverAssetID: template.DraftCoverAssetID, PublishedCoverAssetID: template.PublishedCoverAssetID,
		HomePosition: template.HomePosition, CreatedBy: template.CreatedBy, UpdatedBy: template.UpdatedBy,
		CreatedAt: template.CreatedAt, UpdatedAt: template.UpdatedAt, PublishedAt: template.PublishedAt,
	}
}

func imageCreationAdminTemplateDTOs(templates []service.ImageCreationTemplate) []imageCreationAdminTemplateDTO {
	out := make([]imageCreationAdminTemplateDTO, 0, len(templates))
	for i := range templates {
		out = append(out, imageCreationAdminTemplateDTOFromService(&templates[i]))
	}
	return out
}
