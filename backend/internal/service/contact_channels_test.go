package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"errors"
	"image"
	"image/color"
	"image/png"
	"strings"
	"sync"
	"testing"
)

// ============== fakeSettingRepo ==============

type fakeSettingRepo struct {
	mu   sync.Mutex
	data map[string]string
}

func newFakeRepo(initial map[string]string) *fakeSettingRepo {
	cp := make(map[string]string, len(initial))
	for k, v := range initial {
		cp[k] = v
	}
	return &fakeSettingRepo{data: cp}
}

func (r *fakeSettingRepo) Get(ctx context.Context, key string) (*Setting, error) {
	return nil, errors.New("not implemented")
}
func (r *fakeSettingRepo) GetValue(ctx context.Context, key string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	v, ok := r.data[key]
	if !ok {
		return "", ErrSettingNotFound
	}
	return v, nil
}
func (r *fakeSettingRepo) Set(ctx context.Context, key, value string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[key] = value
	return nil
}
func (r *fakeSettingRepo) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := map[string]string{}
	for _, k := range keys {
		if v, ok := r.data[k]; ok {
			out[k] = v
		}
	}
	return out, nil
}
func (r *fakeSettingRepo) SetMultiple(ctx context.Context, settings map[string]string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for k, v := range settings {
		r.data[k] = v
	}
	return nil
}
func (r *fakeSettingRepo) GetAll(ctx context.Context) (map[string]string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]string, len(r.data))
	for k, v := range r.data {
		out[k] = v
	}
	return out, nil
}
func (r *fakeSettingRepo) Delete(ctx context.Context, key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.data, key)
	return nil
}

// ============== helpers ==============

func makePNGDataURL(t *testing.T, w, h int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.Set(x, y, color.Black)
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("png encode: %v", err)
	}
	return "data:image/png;base64," + base64.StdEncoding.EncodeToString(buf.Bytes())
}

// ============== publicContactChannels ==============

func TestPublicContactChannels_CorruptJSON_ReturnsEmpty(t *testing.T) {
	out := publicContactChannels("not-json{{{")
	if len(out) != 0 {
		t.Fatalf("expected empty, got %v", out)
	}
}

func TestPublicContactChannels_FiltersDisabledAndUnknownTypes_SortsByPriority(t *testing.T) {
	raw := `[
		{"type":"customer_service","label":"cs","enabled":true,"priority":2,"qr_image":"x"},
		{"type":"unknown_type","label":"x","enabled":true,"priority":0,"qr_image":"x"},
		{"type":"wechat_group","label":"wg","enabled":true,"priority":1,"qr_image":"x"},
		{"type":"official_account","label":"oa","enabled":false,"priority":0,"qr_image":"x"}
	]`
	out := publicContactChannels(raw)
	if len(out) != 2 {
		t.Fatalf("expected 2 channels, got %d (%v)", len(out), out)
	}
	if out[0].Type != ContactChannelTypeWeChatGroup || out[1].Type != ContactChannelTypeCustomerService {
		t.Fatalf("unexpected order: %v", out)
	}
}

// ============== validateContactChannelImage ==============

func TestValidateContactChannelImage_RejectsSVG(t *testing.T) {
	svg := "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(`<svg/>`))
	if err := validateContactChannelImage(svg); err == nil {
		t.Fatal("expected SVG to be rejected, got nil error")
	}
}

func TestValidateContactChannelImage_RejectsHTML(t *testing.T) {
	html := "data:text/html;base64," + base64.StdEncoding.EncodeToString([]byte(`<script>alert(1)</script>`))
	if err := validateContactChannelImage(html); err == nil {
		t.Fatal("expected HTML to be rejected, got nil error")
	}
}

func TestValidateContactChannelImage_RejectsMIMEMismatch(t *testing.T) {
	// 真实 PNG，但 MIME 写成 jpeg
	good := makePNGDataURL(t, 8, 8)
	mismatched := strings.Replace(good, "data:image/png;", "data:image/jpeg;", 1)
	if err := validateContactChannelImage(mismatched); err == nil {
		t.Fatal("expected MIME mismatch to be rejected")
	}
}

func TestValidateContactChannelImage_RejectsTooLarge(t *testing.T) {
	// 拼一个超 30KB 的 base64 PNG header + 填充
	payload := strings.Repeat("A", 60*1024) // base64 字符
	big := "data:image/png;base64," + payload
	if err := validateContactChannelImage(big); err == nil {
		t.Fatal("expected too-large image to be rejected")
	}
}

func TestValidateContactChannelImage_AcceptsSmallPNG(t *testing.T) {
	good := makePNGDataURL(t, 8, 8)
	if err := validateContactChannelImage(good); err != nil {
		t.Fatalf("expected small PNG to pass, got %v", err)
	}
}

func TestValidateContactChannelImage_RejectsMalformedDataURL(t *testing.T) {
	cases := []string{
		"not-a-data-url",
		"data:image/png,foo",                  // missing ;base64
		"data:image/png;base64",               // no comma
		"data:image/png;base64,!!!not-base64", // payload garbage
	}
	for _, in := range cases {
		if err := validateContactChannelImage(in); err == nil {
			t.Errorf("expected reject, got nil for %q", in)
		}
	}
}

// ============== UpdateContactChannels ==============

func mustService(repo *fakeSettingRepo) *SettingService {
	return &SettingService{settingRepo: repo}
}

func TestUpdateContactChannels_RejectsInvalidType(t *testing.T) {
	s := mustService(newFakeRepo(nil))
	_, err := s.UpdateContactChannels(context.Background(), []ContactChannel{
		{Type: "weibo", Label: "x", Enabled: false},
	})
	if err == nil {
		t.Fatal("expected invalid type to be rejected")
	}
}

func TestUpdateContactChannels_RejectsDuplicateType(t *testing.T) {
	s := mustService(newFakeRepo(nil))
	_, err := s.UpdateContactChannels(context.Background(), []ContactChannel{
		{Type: ContactChannelTypeWeChatGroup, Label: "a", Enabled: false},
		{Type: ContactChannelTypeWeChatGroup, Label: "b", Enabled: false},
	})
	if err == nil {
		t.Fatal("expected duplicate type to be rejected")
	}
}

func TestUpdateContactChannels_RejectsTooLongLabel(t *testing.T) {
	s := mustService(newFakeRepo(nil))
	_, err := s.UpdateContactChannels(context.Background(), []ContactChannel{
		{Type: ContactChannelTypeWeChatGroup, Label: strings.Repeat("a", 80), Enabled: false},
	})
	if err == nil {
		t.Fatal("expected too-long label to be rejected")
	}
}

func TestUpdateContactChannels_RejectsEnabledWithoutImage(t *testing.T) {
	s := mustService(newFakeRepo(nil))
	_, err := s.UpdateContactChannels(context.Background(), []ContactChannel{
		{Type: ContactChannelTypeWeChatGroup, Label: "ok", Enabled: true},
	})
	if err == nil {
		t.Fatal("expected enabled-without-image to be rejected")
	}
}

func TestUpdateContactChannels_AcceptsValidPayload_CallsOnUpdate(t *testing.T) {
	repo := newFakeRepo(nil)
	s := mustService(repo)
	called := false
	s.SetOnUpdateCallback(func() { called = true })

	out, err := s.UpdateContactChannels(context.Background(), []ContactChannel{
		{Type: ContactChannelTypeWeChatGroup, Label: "群", QRImage: makePNGDataURL(t, 8, 8), Description: "join", Enabled: true, Priority: 0},
		{Type: ContactChannelTypeCustomerService, Label: "客服", Enabled: false, Priority: 1},
	})
	if err != nil {
		t.Fatalf("expected ok, got %v", err)
	}
	if len(out) != 2 {
		t.Fatalf("expected 2 channels back, got %d", len(out))
	}
	if !called {
		t.Fatal("expected onUpdate callback to fire after save")
	}
	if _, ok := repo.data[SettingKeyContactChannels]; !ok {
		t.Fatal("expected setting to be persisted")
	}
}
