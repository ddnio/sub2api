package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sort"
	"strings"
)

// ContactChannelType 渠道固定类型枚举
const (
	ContactChannelTypeWeChatGroup       = "wechat_group"
	ContactChannelTypeCustomerService   = "customer_service"
	ContactChannelTypeOfficialAccount   = "official_account"
)

// ValidContactChannelTypes 渠道允许的固定类型集合
var ValidContactChannelTypes = map[string]struct{}{
	ContactChannelTypeWeChatGroup:     {},
	ContactChannelTypeCustomerService: {},
	ContactChannelTypeOfficialAccount: {},
}

// 校验上限
const (
	contactChannelLabelMaxLen       = 50
	contactChannelDescriptionMaxLen = 500
	contactChannelExtraInfoMaxLen   = 200
	contactChannelImageMaxBytes     = 30 * 1024  // 单图解码后 ≤30KB
	contactChannelsJSONMaxBytes     = 150 * 1024 // 总 JSON ≤150KB
)

// 允许的图片 MIME（白名单）— 拒绝 SVG/HTML 等可执行格式
var allowedContactChannelMIMEs = map[string]struct{}{
	"image/png":  {},
	"image/jpeg": {},
}

// ContactChannel 单个联系渠道
type ContactChannel struct {
	Type        string `json:"type"`
	Label       string `json:"label"`
	QRImage     string `json:"qr_image"` // data URL: data:image/png;base64,xxx
	Description string `json:"description"`
	ExtraInfo   string `json:"extra_info"`
	Enabled     bool   `json:"enabled"`
	Priority    int    `json:"priority"`
}

// DefaultContactChannels v1 预置 3 类固定 type，全部 disabled
func DefaultContactChannels() []ContactChannel {
	return []ContactChannel{
		{Type: ContactChannelTypeWeChatGroup, Label: "用户交流群", Enabled: false, Priority: 0},
		{Type: ContactChannelTypeCustomerService, Label: "客服微信", Enabled: false, Priority: 1},
		{Type: ContactChannelTypeOfficialAccount, Label: "公众号", Enabled: false, Priority: 2},
	}
}

// parseContactChannelsRaw JSON 解析：失败容错返回空数组（不报错）
func parseContactChannelsRaw(raw string) []ContactChannel {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []ContactChannel{}
	}
	var items []ContactChannel
	if err := json.Unmarshal([]byte(raw), &items); err != nil {
		return []ContactChannel{}
	}
	if items == nil {
		return []ContactChannel{}
	}
	return items
}

// publicContactChannels 仅返回 enabled 的渠道、按 priority 升序、type 校验
func publicContactChannels(raw string) []ContactChannel {
	all := parseContactChannelsRaw(raw)
	out := make([]ContactChannel, 0, len(all))
	for _, c := range all {
		if !c.Enabled {
			continue
		}
		if _, ok := ValidContactChannelTypes[c.Type]; !ok {
			continue
		}
		out = append(out, c)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Priority < out[j].Priority
	})
	return out
}

// GetContactChannelsForAdmin admin 视图：返回完整数组（含 disabled，按 priority）
func (s *SettingService) GetContactChannelsForAdmin(ctx context.Context) ([]ContactChannel, error) {
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyContactChannels)
	if err != nil {
		if errors.Is(err, ErrSettingNotFound) {
			return DefaultContactChannels(), nil
		}
		return nil, fmt.Errorf("get contact_channels: %w", err)
	}
	channels := parseContactChannelsRaw(raw)
	// 数据库里损坏 / 缺省 → 返回默认骨架，避免 admin UI 空白
	if len(channels) == 0 {
		return DefaultContactChannels(), nil
	}
	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Priority < channels[j].Priority
	})
	return channels, nil
}

// UpdateContactChannels 持久化 admin 提交的渠道列表，含完整校验
func (s *SettingService) UpdateContactChannels(ctx context.Context, channels []ContactChannel) ([]ContactChannel, error) {
	if s == nil || s.settingRepo == nil {
		return nil, errors.New("setting repository not initialized")
	}
	if channels == nil {
		channels = []ContactChannel{}
	}

	// 单条校验 + 归一化
	seenType := map[string]struct{}{}
	for i := range channels {
		c := &channels[i]
		c.Type = strings.TrimSpace(c.Type)
		c.Label = strings.TrimSpace(c.Label)
		c.Description = strings.TrimSpace(c.Description)
		c.ExtraInfo = strings.TrimSpace(c.ExtraInfo)
		c.QRImage = strings.TrimSpace(c.QRImage)

		if _, ok := ValidContactChannelTypes[c.Type]; !ok {
			return nil, fmt.Errorf("invalid contact channel type: %q", c.Type)
		}
		if _, dup := seenType[c.Type]; dup {
			return nil, fmt.Errorf("duplicated contact channel type: %q", c.Type)
		}
		seenType[c.Type] = struct{}{}

		if len([]rune(c.Label)) > contactChannelLabelMaxLen {
			return nil, fmt.Errorf("label too long for type %s: max %d chars", c.Type, contactChannelLabelMaxLen)
		}
		if len([]rune(c.Description)) > contactChannelDescriptionMaxLen {
			return nil, fmt.Errorf("description too long for type %s: max %d chars", c.Type, contactChannelDescriptionMaxLen)
		}
		if len([]rune(c.ExtraInfo)) > contactChannelExtraInfoMaxLen {
			return nil, fmt.Errorf("extra_info too long for type %s: max %d chars", c.Type, contactChannelExtraInfoMaxLen)
		}

		// 启用渠道必须有图（避免 UI 空 tab）
		if c.Enabled && c.QRImage == "" {
			return nil, fmt.Errorf("enabled channel %s requires qr_image", c.Type)
		}

		// 图片严格校验（非空时即校验，禁用渠道也不能存非法图）
		if c.QRImage != "" {
			if err := validateContactChannelImage(c.QRImage); err != nil {
				return nil, fmt.Errorf("invalid qr_image for type %s: %w", c.Type, err)
			}
		}
	}

	// 整体 JSON 体积上限
	raw, err := json.Marshal(channels)
	if err != nil {
		return nil, fmt.Errorf("marshal contact_channels: %w", err)
	}
	if len(raw) > contactChannelsJSONMaxBytes {
		return nil, fmt.Errorf("contact_channels payload too large: %d bytes (max %d)", len(raw), contactChannelsJSONMaxBytes)
	}

	if err := s.settingRepo.Set(ctx, SettingKeyContactChannels, string(raw)); err != nil {
		return nil, fmt.Errorf("save contact_channels: %w", err)
	}

	// 触发缓存失效（如 HTML 注入缓存）
	if s.onUpdate != nil {
		s.onUpdate()
	}

	sort.SliceStable(channels, func(i, j int) bool {
		return channels[i].Priority < channels[j].Priority
	})
	return channels, nil
}

// validateContactChannelImage 校验 data URL 图片：MIME 白名单 + 解码 + 大小 + 真实图片格式
func validateContactChannelImage(dataURL string) error {
	const prefix = "data:"
	if !strings.HasPrefix(dataURL, prefix) {
		return errors.New("not a data URL")
	}
	// 拆 "data:<mime>;base64,<payload>"
	commaIdx := strings.Index(dataURL, ",")
	if commaIdx < 0 {
		return errors.New("malformed data URL")
	}
	header := dataURL[len(prefix):commaIdx]
	payload := dataURL[commaIdx+1:]

	// header: <mime>[;base64]
	parts := strings.Split(header, ";")
	if len(parts) < 2 || parts[len(parts)-1] != "base64" {
		return errors.New("data URL must be base64-encoded")
	}
	mime := strings.ToLower(strings.TrimSpace(parts[0]))
	if _, ok := allowedContactChannelMIMEs[mime]; !ok {
		return fmt.Errorf("unsupported image MIME: %s (only image/png and image/jpeg allowed)", mime)
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return fmt.Errorf("base64 decode failed: %w", err)
	}
	if len(decoded) > contactChannelImageMaxBytes {
		return fmt.Errorf("image too large: %d bytes (max %d)", len(decoded), contactChannelImageMaxBytes)
	}
	if len(decoded) == 0 {
		return errors.New("empty image payload")
	}

	// 真实格式校验：确实能被解码为 PNG/JPEG
	_, format, err := image.DecodeConfig(bytes.NewReader(decoded))
	if err != nil {
		return fmt.Errorf("image decode failed: %w", err)
	}
	switch format {
	case "png", "jpeg":
		// pass
	default:
		return fmt.Errorf("decoded image format %q not allowed", format)
	}

	// MIME 与真实格式必须一致，防止伪造
	expected := strings.TrimPrefix(mime, "image/")
	if expected == "jpg" {
		expected = "jpeg"
	}
	if expected != format {
		return fmt.Errorf("MIME %s mismatch with actual format %s", mime, format)
	}
	return nil
}
