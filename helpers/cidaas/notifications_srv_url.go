package cidaas

import "strings"

const DefaultNotificationsContextPath = "notifications-srv"

func NormalizeNotificationsContextPath(cfg ClientConfig) string {
	p := strings.Trim(cfg.NotificationsContextPath, "/")
	if p == "" {
		return DefaultNotificationsContextPath
	}
	return p
}

func SegmentNotificationsURL(cfg ClientConfig, parts ...string) string {
	u := strings.TrimSuffix(cfg.BaseURL, "/") + "/" + NormalizeNotificationsContextPath(cfg)
	for _, p := range parts {
		u += "/" + strings.Trim(p, "/")
	}
	return u
}
