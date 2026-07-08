-- Add review-reminder tracking columns to the orders table, derived from the
-- GORM Order model (shipped_at, review_reminder_sent_at) via cmd/atlasloader.
-- Both are nullable timestamps; shipped_at is indexed for the review-reminder
-- query (idx_orders_shipped_at). Postgres-native (not document-typed) data.

ALTER TABLE "orders" ADD COLUMN "shipped_at" timestamptz;
ALTER TABLE "orders" ADD COLUMN "review_reminder_sent_at" timestamptz;
CREATE INDEX IF NOT EXISTS "idx_orders_shipped_at" ON "orders" ("shipped_at");
