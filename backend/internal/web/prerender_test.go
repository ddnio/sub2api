//go:build embed

package web

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/fstest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPrerenderManifest = `{
  "placeholders": {
    "siteOrigin": "https://site.prerender.invalid",
    "apiBase": "https://api.prerender.invalid",
    "siteName": "__PRERENDER_SITE_NAME__",
    "contact": "__PRERENDER_CONTACT__"
  },
  "pages": { "/docs": "quick-start.html", "/docs/faq": "faq.html" }
}`

const testPrerenderPage = `<html><head><title>FAQ - __PRERENDER_SITE_NAME__</title>` +
	`<link rel="canonical" href="https://site.prerender.invalid/docs/faq"></head>` +
	`<body><div id="app"><pre>https://api.prerender.invalid/v1</pre><p>__PRERENDER_CONTACT__</p></div></body></html>`

func newPrerenderTestServer(t *testing.T, settings *mockSettingsProvider, files fstest.MapFS) *FrontendServer {
	t.Helper()
	base := fstest.MapFS{
		"index.html":                 {Data: []byte(`<html><head><title>SPA</title></head><body><div id="app"></div></body></html>`)},
		"prerender/manifest.json":    {Data: []byte(testPrerenderManifest)},
		"prerender/quick-start.html": {Data: []byte(testPrerenderPage)},
		"prerender/faq.html":         {Data: []byte(testPrerenderPage)},
	}
	for name, file := range files {
		if file == nil {
			delete(base, name)
		} else {
			base[name] = file
		}
	}
	cache := NewHTMLCache()
	cache.SetBaseHTML(base["index.html"].Data)
	s := &FrontendServer{
		distFS:   base,
		baseHTML: base["index.html"].Data,
		cache:    cache,
		settings: settings,
	}
	s.prerender = loadPrerenderedPages(base)
	s.SetSiteOrigin("https://router.example.com/")
	return s
}

func servePrerenderTest(s *FrontendServer, path string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(s.Middleware())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	return w
}

func liveSettings() *mockSettingsProvider {
	return &mockSettingsProvider{settings: map[string]any{
		"site_name":    "NanaFox <Router>",
		"api_base_url": "",
		"contact_info": "微信：nanafox-hi",
	}}
}

func TestPrerender_FillsPlaceholdersFromSettings(t *testing.T) {
	s := newPrerenderTestServer(t, liveSettings(), nil)

	w := servePrerenderTest(s, "/docs/faq")
	require.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "<title>FAQ - NanaFox &lt;Router&gt;</title>")
	assert.Contains(t, body, `href="https://router.example.com/docs/faq"`)
	assert.Contains(t, body, "<pre>https://router.example.com/v1</pre>")
	assert.Contains(t, body, "微信：nanafox-hi")
	assert.Contains(t, body, "window.__APP_CONFIG__=")
	assert.NotContains(t, body, "prerender.invalid")
	assert.NotContains(t, body, "__PRERENDER_")
	assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))

	// 尾斜杠同一页；设置只读一次
	assert.Contains(t, servePrerenderTest(s, "/docs/").Body.String(), "FAQ - NanaFox")
	assert.Equal(t, 1, s.settings.(*mockSettingsProvider).called)
}

func TestPrerender_UsesAPIBaseURLWithoutV1(t *testing.T) {
	settings := liveSettings()
	settings.settings.(map[string]any)["api_base_url"] = "https://api.example.com/v1/"
	s := newPrerenderTestServer(t, settings, nil)

	body := servePrerenderTest(s, "/docs/faq").Body.String()
	assert.Contains(t, body, "<pre>https://api.example.com/v1</pre>")
	assert.Contains(t, body, `href="https://router.example.com/docs/faq"`)
}

func TestPrerender_FallsBackToSPAWhenValuesAreMissing(t *testing.T) {
	cases := map[string]func(s *FrontendServer, p *mockSettingsProvider){
		"no contact":       func(_ *FrontendServer, p *mockSettingsProvider) { p.settings.(map[string]any)["contact_info"] = " " },
		"no site name":     func(_ *FrontendServer, p *mockSettingsProvider) { p.settings.(map[string]any)["site_name"] = "" },
		"no frontend_url":  func(s *FrontendServer, _ *mockSettingsProvider) { s.SetSiteOrigin("") },
		"settings failure": func(_ *FrontendServer, p *mockSettingsProvider) { p.err = errors.New("db down") },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			settings := liveSettings()
			s := newPrerenderTestServer(t, settings, nil)
			mutate(s, settings)

			w := servePrerenderTest(s, "/docs/faq")
			require.Equal(t, http.StatusOK, w.Code)
			assert.Contains(t, w.Body.String(), "<title>")
			assert.NotContains(t, w.Body.String(), "FAQ -")
			assert.NotContains(t, w.Body.String(), "prerender.invalid")
		})
	}
}

func TestPrerender_FallsBackWhenPlaceholderIsLeftOver(t *testing.T) {
	s := newPrerenderTestServer(t, liveSettings(), fstest.MapFS{
		"prerender/faq.html": {Data: []byte(`<title>FAQ</title><a href="https%3A%2F%2Fsite.prerender.invalid">x</a>`)},
	})

	body := servePrerenderTest(s, "/docs/faq").Body.String()
	assert.NotContains(t, body, "prerender.invalid")
	assert.NotContains(t, body, "FAQ")
	// 别的页面不受影响
	assert.Contains(t, servePrerenderTest(s, "/docs").Body.String(), "FAQ - NanaFox")
}

func TestPrerender_InvalidateRereadsSettings(t *testing.T) {
	settings := liveSettings()
	s := newPrerenderTestServer(t, settings, nil)
	assert.Contains(t, servePrerenderTest(s, "/docs/faq").Body.String(), "nanafox-hi")

	settings.settings.(map[string]any)["contact_info"] = "QQ 群 123"
	s.InvalidateCache()
	assert.Contains(t, servePrerenderTest(s, "/docs/faq").Body.String(), "QQ 群 123")
}

func TestPrerender_DoesNotExposeTemplatesOrOtherRoutes(t *testing.T) {
	s := newPrerenderTestServer(t, liveSettings(), nil)

	for _, path := range []string{"/prerender/faq.html", "/prerender/manifest.json", "/docs/unknown", "/home"} {
		body := servePrerenderTest(s, path).Body.String()
		assert.NotContains(t, body, "prerender.invalid", path)
		assert.NotContains(t, body, "__PRERENDER_", path)
		assert.NotContains(t, body, `"placeholders"`, path)
	}
}

func TestPrerender_DisabledWithoutManifest(t *testing.T) {
	s := newPrerenderTestServer(t, liveSettings(), fstest.MapFS{"prerender/manifest.json": nil})
	assert.Nil(t, s.prerender)
	assert.NotContains(t, servePrerenderTest(s, "/docs/faq").Body.String(), "FAQ -")
}

// 用真实构建产物走一遍：每个预渲染页都能填干净，标题和规范地址是各自小节的
func TestPrerender_EmbeddedDistPages(t *testing.T) {
	server, err := NewFrontendServer(liveSettings())
	require.NoError(t, err)
	if server.prerender == nil {
		t.Skip("embedded dist was built without docs prerender")
	}
	server.SetSiteOrigin("https://router.example.com")

	require.NotEmpty(t, server.prerender.base)
	for route := range server.prerender.base {
		w := servePrerenderTest(server, route)
		require.Equal(t, http.StatusOK, w.Code, route)
		body := w.Body.String()
		assert.Contains(t, body, `rel="canonical" data-docs-seo="" href="https://router.example.com`+route+`"`, route)
		assert.Contains(t, body, " - NanaFox &lt;Router&gt;</title>", route)
		assert.NotContains(t, body, "prerender.invalid", route)
		assert.NotContains(t, body, "__PRERENDER_", route)
	}
	assert.Contains(t, servePrerenderTest(server, "/docs/clients").Body.String(), "Cherry Studio")
}
