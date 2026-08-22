package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	_ "golang.org/x/image/webp"
)

const ImageCreationAssetMaxBytes = 8 * 1024 * 1024

var imageCreationCategoryPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,63}$`)

type ImageCreationAsset struct {
	ID             string
	Content        []byte
	ContentType    string
	ByteSize       int
	Width          int
	Height         int
	SourceType     string
	SourceProvider string
	SourceModel    string
	CreatedBy      int64
	CreatedAt      time.Time
}

func ValidateImageCreationTemplateDocument(doc domain.ImageCreationTemplateDocument) error {
	if doc.SchemaVersion != 1 {
		return invalidImageCreationInput("schema_version must be 1")
	}
	if strings.TrimSpace(doc.Title) == "" || utf8.RuneCountInString(doc.Title) > 120 {
		return invalidImageCreationInput("title must contain 1 to 120 characters")
	}
	if utf8.RuneCountInString(doc.Summary) > 300 {
		return invalidImageCreationInput("summary must not exceed 300 characters")
	}
	if !imageCreationCategoryPattern.MatchString(doc.Category) {
		return invalidImageCreationInput("category must be a lowercase slug")
	}
	if len(doc.Tags) > 8 {
		return invalidImageCreationInput("tags must not contain more than 8 values")
	}
	seenTags := make(map[string]bool, len(doc.Tags))
	for _, tag := range doc.Tags {
		tag = strings.TrimSpace(tag)
		if tag == "" || utf8.RuneCountInString(tag) > 40 || seenTags[tag] {
			return invalidImageCreationInput("tags must be unique values of 1 to 40 characters")
		}
		seenTags[tag] = true
	}
	if strings.TrimSpace(doc.Prompt) == "" || utf8.RuneCountInString(doc.Prompt) > 12000 {
		return invalidImageCreationInput("prompt must contain 1 to 12000 characters")
	}
	if doc.InputMode != "text" && doc.InputMode != "reference_optional" && doc.InputMode != "reference_required" {
		return invalidImageCreationInput("input_mode is invalid")
	}
	if utf8.RuneCountInString(doc.CoverAlt) > 200 {
		return invalidImageCreationInput("cover_alt must not exceed 200 characters")
	}
	if doc.Defaults.Size != "1024x1024" && doc.Defaults.Size != "1536x1024" && doc.Defaults.Size != "1024x1536" {
		return invalidImageCreationInput("defaults.size is invalid")
	}
	if doc.Defaults.Quality != "low" && doc.Defaults.Quality != "medium" && doc.Defaults.Quality != "high" {
		return invalidImageCreationInput("defaults.quality is invalid")
	}
	if doc.Defaults.OutputFormat != "png" && doc.Defaults.OutputFormat != "jpeg" && doc.Defaults.OutputFormat != "webp" {
		return invalidImageCreationInput("defaults.output_format is invalid")
	}
	if doc.Source == nil {
		return nil
	}
	if utf8.RuneCountInString(doc.Source.Name) > 120 || utf8.RuneCountInString(doc.Source.License) > 120 || utf8.RuneCountInString(doc.Source.Notes) > 1000 {
		return invalidImageCreationInput("source attribution is too long")
	}
	if doc.Source.URL == "" {
		return nil
	}
	sourceURL, err := url.ParseRequestURI(doc.Source.URL)
	if err != nil || sourceURL.Scheme != "https" || sourceURL.Host == "" || len(doc.Source.URL) > 2048 {
		return invalidImageCreationInput("source.url must be a valid HTTPS URL")
	}
	return nil
}

func ValidateImageCreationAsset(content []byte, sourceType, sourceProvider, sourceModel string, createdBy int64) (*ImageCreationAsset, error) {
	if len(content) == 0 {
		return nil, invalidImageCreationInput("image content is required")
	}
	if len(content) > ImageCreationAssetMaxBytes {
		return nil, infraerrors.New(http.StatusRequestEntityTooLarge, "IMAGE_CREATION_ASSET_TOO_LARGE", "image must not exceed 8 MiB")
	}
	if sourceType != "generated" && sourceType != "uploaded" && sourceType != "imported" {
		return nil, invalidImageCreationInput("source_type is invalid")
	}
	if createdBy <= 0 || len(sourceProvider) > 64 || len(sourceModel) > 120 {
		return nil, invalidImageCreationInput("image source metadata is invalid")
	}
	cfg, format, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil || cfg.Width < 1 || cfg.Height < 1 || cfg.Width > 8192 || cfg.Height > 8192 {
		return nil, invalidImageCreationInput("image format or dimensions are invalid")
	}
	contentTypes := map[string]string{"png": "image/png", "jpeg": "image/jpeg", "webp": "image/webp"}
	contentType := contentTypes[format]
	if contentType == "" {
		return nil, invalidImageCreationInput("image must be PNG, JPEG, or WebP")
	}
	hash := sha256.Sum256(content)
	return &ImageCreationAsset{
		ID:             hex.EncodeToString(hash[:]),
		Content:        content,
		ContentType:    contentType,
		ByteSize:       len(content),
		Width:          cfg.Width,
		Height:         cfg.Height,
		SourceType:     sourceType,
		SourceProvider: strings.TrimSpace(sourceProvider),
		SourceModel:    strings.TrimSpace(sourceModel),
		CreatedBy:      createdBy,
	}, nil
}

func invalidImageCreationInput(message string) error {
	return infraerrors.BadRequest("IMAGE_CREATION_INVALID_INPUT", message)
}
