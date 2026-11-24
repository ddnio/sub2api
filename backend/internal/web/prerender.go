//go:build embed

package web

// NanaFox fork: /docs 各小节的预渲染 HTML（前端构建时生成在 dist/prerender/，见
// frontend/src/prerender/docs.prerender.ts），让不跑脚本的爬虫也能读到正文。
// 页面里取自线上设置的值是占位符，这里按当前设置和 server.frontend_url 填好再返回；
// 任一值缺失或没填干净就退回普通 SPA 首页，不返回半成品。

import (
	"bytes"
	"context"
	"encoding/json"
	htmlpkg "html"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/gin-gonic/gin"
)

const prerenderDir = "prerender"

type prerenderManifest struct {
	Placeholders struct {
		SiteOrigin string `json:"siteOrigin"`
		APIBase    string `json:"apiBase"`
		SiteName   string `json:"siteName"`
		Contact    string `json:"contact"`
	} `json:"placeholders"`
	Pages map[string]string `json:"pages"`
}

type prerenderedPages struct {
	manifest prerenderManifest
	base     map[string][]byte // 路由地址 -> 带占位符的 HTML

	mu         sync.RWMutex
	siteOrigin string
	generation uint64            // 设置或站点地址变更时加一，防止旧结果覆盖新结果
	ready      bool              // rendered 对当前设置有效
	rendered   map[string][]byte // 填好的页面；缺值时为空
}

// loadPrerenderedPages 读取 dist/prerender/manifest.json；没有或格式不对时返回 nil（功能关闭）。
func loadPrerenderedPages(distFS fs.FS) *prerenderedPages {
	raw, err := fs.ReadFile(distFS, path.Join(prerenderDir, "manifest.json"))
	if err != nil {
		return nil
	}
	var m prerenderManifest
	if err := json.Unmarshal(raw, &m); err != nil {
		log.Printf("Warning: invalid prerender manifest: %v", err)
		return nil
	}
	ph := m.Placeholders
	if ph.SiteOrigin == "" || ph.APIBase == "" || ph.SiteName == "" || ph.Contact == "" {
		log.Printf("Warning: prerender manifest is missing placeholders")
		return nil
	}
	base := make(map[string][]byte, len(m.Pages))
	for route, file := range m.Pages {
		content, err := fs.ReadFile(distFS, path.Join(prerenderDir, path.Clean("/"+file)))
		if err != nil {
			log.Printf("Warning: prerendered page %s missing: %v", route, err)
			return nil
		}
		base[route] = content
	}
	return &prerenderedPages{manifest: m, base: base}
}

func (p *prerenderedPages) setSiteOrigin(frontendURL string) {
	origin := ""
	if u, err := url.Parse(strings.TrimSpace(frontendURL)); err == nil && (u.Scheme == "http" || u.Scheme == "https") && u.Host != "" {
		origin = u.Scheme + "://" + u.Host
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	p.siteOrigin = origin
	p.resetLocked()
}

func (p *prerenderedPages) invalidate() {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.resetLocked()
}

func (p *prerenderedPages) resetLocked() {
	p.generation++
	p.ready = false
	p.rendered = nil
}

// page 返回填好的页面；该地址没有预渲染、设置读不到或缺值时返回 nil。
func (p *prerenderedPages) page(ctx context.Context, route string, settings PublicSettingsProvider) []byte {
	if _, ok := p.base[route]; !ok {
		return nil
	}
	p.mu.RLock()
	if p.ready {
		content := p.rendered[route]
		p.mu.RUnlock()
		return content
	}
	generation, siteOrigin := p.generation, p.siteOrigin
	p.mu.RUnlock()

	// 读设置不持锁，避免数据库慢时所有文档请求排队
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	data, err := settings.GetPublicSettingsForInjection(ctx)
	if err != nil {
		return nil // 不缓存，下次再试
	}
	settingsJSON, err := json.Marshal(data)
	if err != nil {
		return nil
	}
	rendered := p.renderAll(siteOrigin, settingsJSON)

	p.mu.Lock()
	if p.generation == generation {
		p.rendered = rendered
		p.ready = true
	}
	p.mu.Unlock()
	return rendered[route]
}

var trailingV1 = regexp.MustCompile(`/v1/?$`)

func (p *prerenderedPages) renderAll(siteOrigin string, settingsJSON []byte) map[string][]byte {
	var cfg struct {
		SiteName    string `json:"site_name"`
		APIBaseURL  string `json:"api_base_url"`
		ContactInfo string `json:"contact_info"`
	}
	rendered := map[string][]byte{}
	if err := json.Unmarshal(settingsJSON, &cfg); err != nil {
		return rendered
	}
	siteName := strings.TrimSpace(cfg.SiteName)
	contact := strings.TrimSpace(cfg.ContactInfo)
	if siteOrigin == "" || siteName == "" || contact == "" {
		log.Printf("Warning: docs prerender disabled: server.frontend_url, site_name or contact_info is empty")
		return rendered
	}
	// 和前端文档页一致：(api_base_url || 站点地址) 去掉结尾的 /v1 和斜杠
	apiBase := strings.TrimSpace(cfg.APIBaseURL)
	if apiBase == "" {
		apiBase = siteOrigin
	}
	apiBase = strings.TrimRight(trailingV1.ReplaceAllString(apiBase, ""), "/")
	if u, err := url.Parse(apiBase); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		log.Printf("Warning: docs prerender disabled: api_base_url is not an http(s) URL")
		return rendered
	}

	ph := p.manifest.Placeholders
	replacer := strings.NewReplacer(
		ph.SiteOrigin, htmlpkg.EscapeString(siteOrigin),
		ph.APIBase, htmlpkg.EscapeString(apiBase),
		ph.SiteName, htmlpkg.EscapeString(siteName),
		ph.Contact, htmlpkg.EscapeString(contact),
	)
	for route, base := range p.base {
		filled := []byte(replacer.Replace(string(base)))
		if bytes.Contains(filled, []byte("prerender.invalid")) || bytes.Contains(filled, []byte("__PRERENDER_")) {
			log.Printf("Warning: prerendered page %s still has placeholders, serving SPA instead", route)
			continue
		}
		rendered[route] = injectSiteFavicon(injectAppConfig(filled, settingsJSON), settingsJSON)
	}
	return rendered
}

// tryServePrerendered 命中预渲染地址时返回填好的页面；原始模板目录始终不直接对外。
func (s *FrontendServer) tryServePrerendered(c *gin.Context, requestPath string) bool {
	route := path.Clean("/" + requestPath)
	if route == "/"+prerenderDir || strings.HasPrefix(route, "/"+prerenderDir+"/") {
		s.serveIndexHTML(c)
		return true
	}
	if s.prerender == nil {
		return false
	}
	content := s.prerender.page(c.Request.Context(), route, s.settings)
	if content == nil {
		return false
	}
	applyHTMLNoCacheHeaders(c.Writer.Header())
	c.Data(http.StatusOK, "text/html; charset=utf-8", replaceNoncePlaceholder(content, middleware.GetNonceFromContext(c)))
	c.Abort()
	return true
}

// SetSiteOrigin 设置对外站点地址（server.frontend_url），用于预渲染页面里的链接和规范地址。
func (s *FrontendServer) SetSiteOrigin(frontendURL string) {
	if s != nil && s.prerender != nil {
		s.prerender.setSiteOrigin(frontendURL)
	}
}
