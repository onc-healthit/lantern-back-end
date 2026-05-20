BEGIN;

CREATE TABLE IF NOT EXISTS endpoint_query_errors (
    id              SERIAL PRIMARY KEY,
    list_source     VARCHAR(500),
    error_message   TEXT,
    queried_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
