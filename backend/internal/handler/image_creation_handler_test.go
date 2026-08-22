package handler

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDecodeImageCreationJSONRejectsUnknownAndTrailingValues(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, body := range []string{
		`{"published_version":1,"unknown":true}`,
		`{"published_version":1} {}`,
	} {
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(body))

		var request imageCreationApplyRequest
		require.Error(t, decodeImageCreationJSON(ctx, &request))
	}
}

func TestImageCreationAssetUploadReportsSizeLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	file, err := writer.CreateFormFile("file", "too-large.png")
	require.NoError(t, err)
	_, err = file.Write(bytes.Repeat([]byte("x"), service.ImageCreationAssetMaxBytes+1024*1024+1))
	require.NoError(t, err)
	require.NoError(t, writer.Close())

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/", &body)
	ctx.Request.Header.Set("Content-Type", writer.FormDataContentType())
	ctx.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 9})

	(&ImageCreationHandler{}).AdminUploadAsset(ctx)

	require.Equal(t, http.StatusRequestEntityTooLarge, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"reason":"IMAGE_CREATION_ASSET_TOO_LARGE"`)
}

func TestImageCreationUserListDTODoesNotExposePrompt(t *testing.T) {
	template := &service.ImageCreationTemplate{
		ID:               7,
		PublishedVersion: 2,
		PublishedData: &service.ImageCreationTemplateDocument{
			SchemaVersion: 1,
			Title:         "光影人像",
			Prompt:        "private prompt",
			Defaults:      domain.ImageCreationTemplateDefaults{Size: "1024x1024", Quality: "high", OutputFormat: "png"},
			InputMode:     "text",
		},
	}

	data, err := json.Marshal(imageCreationUserTemplateListDTO(template))
	require.NoError(t, err)
	require.False(t, strings.Contains(string(data), "prompt"))
	require.False(t, strings.Contains(string(data), "draft"))
}

func TestImageCreationApplicationDTOUsesFrontendFieldNames(t *testing.T) {
	application := &service.ImageCreationTemplateApplication{
		TemplateID:       7,
		PublishedVersion: 2,
		Prompt:           "cinematic portrait",
		Defaults:         domain.ImageCreationTemplateDefaults{Size: "1024x1024", Quality: "high", OutputFormat: "png"},
		InputMode:        "text",
	}

	data, err := json.Marshal(imageCreationApplicationDTO(application))
	require.NoError(t, err)
	require.JSONEq(t, `{
		"template_id": 7,
		"published_version": 2,
		"prompt": "cinematic portrait",
		"defaults": {"size":"1024x1024","quality":"high","output_format":"png"},
		"input_mode": "text"
	}`, string(data))
}
