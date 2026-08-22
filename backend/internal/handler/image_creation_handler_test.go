package handler

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/domain"
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
