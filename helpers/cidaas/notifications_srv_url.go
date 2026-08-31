package cidaas

import "strings"

// NormalizeNotificationsContextPath returns the trimmed context segment (no leading/trailing slashes), or DefaultNotificationsContextPath if empty.
func NormalizeNotificationsContextPath(cfg ClientConfig) string {
	p := strings.Trim(cfg.NotificationsContextPath, "/")
	if p == "" {
		return DefaultNotificationsContextPath
	}
	return p
}

// SegmentNotificationsURL builds {baseURL}/{ctx}/part1/part2/... for notification-srv APIs.
func SegmentNotificationsURL(cfg ClientConfig, parts ...string) string {
	u := strings.TrimSuffix(cfg.BaseURL, "/") + "/" + NormalizeNotificationsContextPath(cfg)
	for _, p := range parts {
		u += "/" + strings.Trim(p, "/")
	}
	return u
}
