BEGIN;

DROP TABLE IF EXISTS daily_querying_status;

CREATE TABLE daily_querying_status (status BOOLEAN);

INSERT INTO daily_querying_status VALUES (TRUE);

COMMIT: