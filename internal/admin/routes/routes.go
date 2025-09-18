package routes

import (
	"path"
	"strings"
)

// Normalize 将模块路由转换为相对路径（去除 /api/v1 及管理前缀）。
func Normalize(route string, prefix string) string {
	cleaned := path.Clean("/" + strings.TrimSpace(route))
	segments := splitSegments(cleaned)

	// 去除 API 版本前缀。
	segments = trimLeadingSegments(segments, []string{"api", "v1"})

	sanitizedPrefix := sanitizePrefix(prefix)
	if sanitizedPrefix != "" {
		segments = trimLeadingSegments(segments, []string{sanitizedPrefix})
	}

	// 历史版本默认使用 admin 作为前缀，需要额外剥离。
	segments = trimLeadingSegments(segments, []string{"admin"})

	return joinSegments(segments)
}

// APIPath 根据配置前缀拼接完整的 API 路径。
func APIPath(route string, prefix string) string {
	normalized := Normalize(route, prefix)
	base := "/api/v1"

	sanitizedPrefix := sanitizePrefix(prefix)
	if sanitizedPrefix != "" {
		base += "/" + sanitizedPrefix
	}

	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "" {
		return base
	}

	return base + "/" + normalized
}

func sanitizePrefix(prefix string) string {
	return strings.Trim(strings.TrimSpace(prefix), "/")
}

func splitSegments(cleaned string) []string {
	trimmed := strings.TrimPrefix(cleaned, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func trimLeadingSegments(segments []string, prefix []string) []string {
	if len(prefix) == 0 || len(segments) < len(prefix) {
		return segments
	}

	matches := true
	for i := range prefix {
		if !strings.EqualFold(segments[i], prefix[i]) {
			matches = false
			break
		}
	}

	if matches {
		return segments[len(prefix):]
	}

	return segments
}

func joinSegments(segments []string) string {
	if len(segments) == 0 {
		return "/"
	}
	return "/" + strings.Join(segments, "/")
}
