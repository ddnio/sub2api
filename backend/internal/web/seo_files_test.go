//go:build embed

package web

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// NanaFox fork: robots.txt / sitemap.xml 来自 frontend/public，必须按静态文件返回，
// 不能落到 SPA 回退把首页 HTML 当成 robots.txt。
func TestFrontendServer_ServesSEOFilesAsStatic(t *testing.T) {
	server, err := NewFrontendServer(&mockSettingsProvider{settings: map[string]string{}})
	require.NoError(t, err)
	if _, err := fs.Stat(server.distFS, "robots.txt"); err != nil {
		t.Skip("embedded dist was built without frontend/public SEO files")
	}

	cases := map[string]string{
		"/robots.txt":  "text/plain",
		"/sitemap.xml": "xml",
	}
	for path, wantType := range cases {
		t.Run(path, func(t *testing.T) {
			router := gin.New()
			router.Use(server.Middleware())

			w := httptest.NewRecorder()
			router.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))

			assert.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), wantType)
			assert.False(t, strings.Contains(w.Body.String(), "<html"), "served SPA index instead of %s", path)
		})
	}
}
