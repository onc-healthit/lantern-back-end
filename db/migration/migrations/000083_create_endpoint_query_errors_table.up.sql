BEGIN;

DROP TABLE IF EXISTS endpoint_query_errors CASCADE;

CREATE TABLE IF NOT EXISTS endpoint_query_errors (
    id              SERIAL PRIMARY KEY,
    list_source     VARCHAR(500),
    error_message   TEXT,
    queried_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

DROP TABLE IF EXISTS endpoint_query_errors_history;

CREATE TABLE IF NOT EXISTS endpoint_query_errors_history (
    id              SERIAL PRIMARY KEY,
    list_source     VARCHAR(500),
    error_message   TEXT,
    queried_at      TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

COMMIT;
