package service

import (
	"bytes"
	"image"
	"image/png"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/stretchr/testify/require"
)

func validImageCreationDocument() domain.ImageCreationTemplateDocument {
	return domain.ImageCreationTemplateDocument{
		SchemaVersion: 1,
		Title:         "电影感人像",
		Summary:       "柔和逆光和浅景深",
		Category:      "portrait",
		Tags:          []string{"人像", "电影感"},
		Prompt:        "一张电影感街头人像",
		InputMode:     "text",
		CoverAlt:      "逆光中的街头人像",
		Defaults: domain.ImageCreationTemplateDefaults{
			Size: "1024x1024", Quality: "high", OutputFormat: "png",
		},
	}
}

func TestValidateImageCreationTemplateDocument(t *testing.T) {
	require.NoError(t, ValidateImageCreationTemplateDocument(validImageCreationDocument()))
	doc := validImageCreationDocument()
	doc.CoverFit = "contain"
	require.NoError(t, ValidateImageCreationTemplateDocument(doc))

	tests := []struct {
		name   string
		mutate func(*domain.ImageCreationTemplateDocument)
	}{
		{"schema version", func(doc *domain.ImageCreationTemplateDocument) { doc.SchemaVersion = 2 }},
		{"title length", func(doc *domain.ImageCreationTemplateDocument) { doc.Title = strings.Repeat("字", 121) }},
		{"empty prompt", func(doc *domain.ImageCreationTemplateDocument) { doc.Prompt = "  " }},
		{"too many tags", func(doc *domain.ImageCreationTemplateDocument) { doc.Tags = make([]string, 9) }},
		{"duplicate tags", func(doc *domain.ImageCreationTemplateDocument) { doc.Tags = []string{"人像", "人像"} }},
		{"invalid category", func(doc *domain.ImageCreationTemplateDocument) { doc.Category = "Portrait Space" }},
		{"invalid input mode", func(doc *domain.ImageCreationTemplateDocument) { doc.InputMode = "image-only" }},
		{"invalid cover fit", func(doc *domain.ImageCreationTemplateDocument) { doc.CoverFit = "stretch" }},
		{"invalid size", func(doc *domain.ImageCreationTemplateDocument) { doc.Defaults.Size = "2048x2048" }},
		{"invalid quality", func(doc *domain.ImageCreationTemplateDocument) { doc.Defaults.Quality = "ultra" }},
		{"invalid format", func(doc *domain.ImageCreationTemplateDocument) { doc.Defaults.OutputFormat = "gif" }},
		{"non-https source", func(doc *domain.ImageCreationTemplateDocument) {
			doc.Source = &domain.ImageCreationTemplateSourceAttribution{URL: "http://example.com/template"}
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc := validImageCreationDocument()
			tt.mutate(&doc)
			require.Error(t, ValidateImageCreationTemplateDocument(doc))
		})
	}
}

func TestValidateImageCreationAsset(t *testing.T) {
	var data bytes.Buffer
	require.NoError(t, png.Encode(&data, image.NewRGBA(image.Rect(0, 0, 12, 8))))

	asset, err := ValidateImageCreationAsset(data.Bytes(), "uploaded", "", "", 9)
	require.NoError(t, err)
	require.Equal(t, "image/png", asset.ContentType)
	require.Equal(t, 12, asset.Width)
	require.Equal(t, 8, asset.Height)
	require.Len(t, asset.ID, 64)

	_, err = ValidateImageCreationAsset([]byte("not an image"), "uploaded", "", "", 9)
	require.Error(t, err)
	_, err = ValidateImageCreationAsset(make([]byte, ImageCreationAssetMaxBytes+1), "uploaded", "", "", 9)
	require.ErrorIs(t, err, ErrImageCreationAssetTooLarge)

	var oversized bytes.Buffer
	require.NoError(t, png.Encode(&oversized, image.NewRGBA(image.Rect(0, 0, 8193, 1))))
	_, err = ValidateImageCreationAsset(oversized.Bytes(), "generated", "openai", "gpt-image-2", 9)
	require.Error(t, err)
}
