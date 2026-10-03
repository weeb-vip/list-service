-- Transactional outbox (go-outbox-lib). List changes that the activity feed
-- cares about are written here in the same transaction as the user_anime or
-- user_work row, and `relay outbox` publishes them to NATS afterwards.
--
-- The shape is the library's outbox.Schema, copied so this chain is verified
-- by the migrations workflow without the library. The partial index is the
-- relay's only read: unpublished rows, oldest first.
CREATE TABLE IF NOT EXISTS outbox_events (
    id           uuid PRIMARY KEY,
    subject      text        NOT NULL,
    payload      jsonb       NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz
);
CREATE INDEX IF NOT EXISTS idx_outbox_events_unpublished ON outbox_events (created_at, id) WHERE published_at IS NULL;
