-- Create "product_rating_records" table
CREATE TABLE "product_rating_records" (
  "customer_id" bigint NOT NULL,
  "product_id" bigint NOT NULL,
  PRIMARY KEY ("customer_id", "product_id")
);
