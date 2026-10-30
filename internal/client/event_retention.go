package client

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
)

// EventRetentionService calls webhook-srv event description retention APIs
// (/webhook-srv/eventdescriptions/retention).
type EventRetentionService struct {
	cfg Config
}

// NewEventRetentionService returns an EventRetentionService bound to cfg.
func NewEventRetentionService(cfg Config) *EventRetentionService {
	return &EventRetentionService{cfg: cfg}
}

// ErrRetentionInactive is returned when webhook-srv answers 403 "retention is not active"
// (event descriptions are not seeded, i.e. the custom_reporting feature is off for the tenant).
var ErrRetentionInactive = errors.New("event retention is not active on this tenant (custom_reporting feature required)")

// RetentionBounds matches webhook-srv EventDescription.retention.
type RetentionBounds struct {
	DefaultDaysToLive int `json:"defaultDaysToLive"`
	MinDaysToLive     int `json:"minDaysToLive"`
	MaxDaysToLive     int `json:"maxDaysToLive"`
}

// RetentionEventDescription is the subset of webhook-srv EventDescription returned by GET /retention.
type RetentionEventDescription struct {
	ID                  string           `json:"_id"`
	ObjectType          string           `json:"objectType"`
	GoodForChangelog    bool             `json:"goodForChangelog"`
	GoodForUserActivity bool             `json:"goodForUserActivity"`
	DaysToLive          int              `json:"daysToLive"`
	Retention           *RetentionBounds `json:"retention,omitempty"`
}

// EventRetentionItem matches webhook-srv PATCH /retention response items.
type EventRetentionItem struct {
	EventID    string `json:"eventId"`
	DaysToLive int    `json:"daysToLive"`
}

type retentionCatalogResponse struct {
	Success bool                        `json:"success"`
	Status  int                         `json:"status"`
	Data    []RetentionEventDescription `json:"data"`
}

type retentionUpdateResponse struct {
	Success bool                 `json:"success"`
	Status  int                  `json:"status"`
	Data    []EventRetentionItem `json:"data"`
}

func (s *EventRetentionService) endpoint() (string, error) {
	return url.JoinPath(s.cfg.BaseURL, "webhook-srv/eventdescriptions/retention")
}

// List GETs the mutable retention catalog (changelog + user activity event types with bounds).
// webhook-srv answers 204 with no body when nothing matches; that yields an empty slice.
func (s *EventRetentionService) List(ctx context.Context) ([]RetentionEventDescription, error) {
	endpoint, err := s.endpoint()
	if err != nil {
		return nil, err
	}
	var out retentionCatalogResponse
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, &out, http.StatusOK, http.StatusNoContent); err != nil {
		return nil, mapRetentionErr(err)
	}
	return out.Data, nil
}

// Update PATCHes the retention map (event type ID -> daysToLive).
// webhook-srv applies entries one by one, so a failing entry can leave earlier ones applied.
func (s *EventRetentionService) Update(ctx context.Context, daysToLive map[string]int) ([]EventRetentionItem, error) {
	endpoint, err := s.endpoint()
	if err != nil {
		return nil, err
	}
	var out retentionUpdateResponse
	if err := requestJSON(ctx, s.cfg, http.MethodPatch, endpoint, daysToLive, &out, http.StatusOK); err != nil {
		return nil, mapRetentionErr(err)
	}
	return out.Data, nil
}

// EventTypeExists reports whether webhook-srv knows the event type at all
// (GET /webhook-srv/eventdescriptions/{id}); used to tell unknown IDs from immutable ones.
func (s *EventRetentionService) EventTypeExists(ctx context.Context, id string) (bool, error) {
	endpoint, err := url.JoinPath(s.cfg.BaseURL, "webhook-srv/eventdescriptions", id)
	if err != nil {
		return false, err
	}
	if err := requestJSON(ctx, s.cfg, http.MethodGet, endpoint, nil, nil, http.StatusOK); err != nil {
		if errors.Is(err, ErrNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// mapRetentionErr matches the message, not just the 403: a token missing the retention scopes also
// gets 403, and Read drops the resource from state on ErrRetentionInactive. An unmatched message
// therefore fails loudly instead of silently removing state.
func mapRetentionErr(err error) error {
	if err != nil && strings.Contains(strings.ToLower(err.Error()), "retention is not active") {
		return errors.Join(ErrRetentionInactive, err)
	}
	return err
}
