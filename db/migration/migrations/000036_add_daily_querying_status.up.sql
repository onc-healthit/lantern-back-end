BEGIN;

DROP TABLE IF EXISTS daily_querying_status;

CREATE TABLE daily_querying_status (status VARCHAR(500));

INSERT INTO daily_querying_status VALUES ('true');

COMMIT: