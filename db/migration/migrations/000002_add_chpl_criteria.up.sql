BEGIN;

CREATE TABLE IF NOT EXISTS certification_criteria (
    id                        SERIAL PRIMARY KEY,
    certification_id          INTEGER,
	cerification_number       VARCHAR(500),
	title                     VARCHAR(500),
	certification_edition_id  INTEGER,
	certification_edition     VARCHAR(500),
	description               VARCHAR(500),
	removed                   BOOLEAN,
    created_at                TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS product_criteria (
    healthit_product_id      INT REFERENCES healthit_products(id) ON DELETE CASCADE,
    certification_id         INTEGER,
    certification_number     VARCHAR(500),
    created_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT product_crit  PRIMARY KEY (healthit_product_id, certification_id)
);

CREATE TRIGGER set_timestamp_certification_criteria
BEFORE UPDATE ON certification_criteria
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

CREATE TRIGGER set_timestamp_product_criteria
BEFORE UPDATE ON product_criteria
FOR EACH ROW
EXECUTE PROCEDURE trigger_set_timestamp();

COMMIT;