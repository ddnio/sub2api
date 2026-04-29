package service

import "strings"

// firstNonEmptyString 返回第一个非空白字符串值。
// TODO: slice-4-openai-image 会带入 openai_images.go 包含同名 helper，届时本文件可移除。
func firstNonEmptyString(values ...any) string {
	for _, value := range values {
		if s, ok := value.(string); ok && strings.TrimSpace(s) != "" {
			return strings.TrimSpace(s)
		}
	}
	return ""
}
