// Package events defines what list-service announces about a user's lists,
// and the seam through which it is written to the transactional outbox.
package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/weeb-vip/go-outbox-lib"
	"gorm.io/gorm"
)

// Subject is the NATS subject every activity event is published on.
const Subject = "user-activity"

// Event types. The feed shows what someone added and where they moved it,
// not how many episodes they are into it: progress-only saves emit nothing.
const (
	AnimeAdded         = "anime.added"
	AnimeStatusChanged = "anime.status_changed"
	AnimeScored        = "anime.scored"
	AnimeRemoved       = "anime.removed"
	WorkAdded          = "work.added"
	WorkStatusChanged  = "work.status_changed"
	WorkScored         = "work.scored"
	WorkRemoved        = "work.removed"
)

// Activity is the payload on Subject.
//
// ID is also the outbox row id and the JetStream message id, so a consumer
// can be idempotent on the body alone. Exactly one of AnimeID and WorkID is
// set. Status values are the stored ones, which are the GraphQL enum names
// (WATCHING, COMPLETED, READING, ...); consumers map them to their own enums.
type Activity struct {
	ID             string    `json:"id"`
	Type           string    `json:"type"`
	UserID         string    `json:"user_id"`
	AnimeID        *string   `json:"anime_id,omitempty"`
	WorkID         *string   `json:"work_id,omitempty"`
	Status         *string   `json:"status,omitempty"`
	PreviousStatus *string   `json:"previous_status,omitempty"`
	Score          *float64  `json:"score,omitempty"`
	OccurredAt     time.Time `json:"occurred_at"`
}

// Writer records an activity in the caller's transaction. It fills in ID and
// OccurredAt. Tests substitute it to assert on what would be published.
type Writer interface {
	Write(ctx context.Context, tx *gorm.DB, activity *Activity) error
}

// OutboxWriter is the Writer backed by go-outbox-lib.
type OutboxWriter struct{}

// Write implements Writer.
func (OutboxWriter) Write(ctx context.Context, tx *gorm.DB, activity *Activity) error {
	id := uuid.New()
	activity.ID = id.String()
	activity.OccurredAt = time.Now().UTC()

	body, err := json.Marshal(activity)
	if err != nil {
		return fmt.Errorf("events: marshal activity: %w", err)
	}

	return outbox.WriteEvent(ctx, tx, &outbox.Event{ID: id, Subject: Subject, Payload: outbox.JSON(body)})
}

// Decide works out which event, if any, a change to a list row deserves.
//
// previous is nil for a new row. A change to anything but status or score --
// episodes, chapters, tags, rewatch counters, list membership -- is "" so
// that saving progress from a detail page, which goes through the same
// upsert, stays out of the feed.
func Decide(created bool, previousStatus, newStatus *string, previousScore, newScore *float64, added, statusChanged, scored string) string {
	if created {
		return added
	}
	if !equalString(previousStatus, newStatus) {
		return statusChanged
	}
	if !equalFloat(previousScore, newScore) {
		return scored
	}

	return ""
}

func equalString(a, b *string) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}

func equalFloat(a, b *float64) bool {
	switch {
	case a == nil && b == nil:
		return true
	case a == nil || b == nil:
		return false
	default:
		return *a == *b
	}
}
