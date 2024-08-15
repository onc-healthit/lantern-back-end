BEGIN;

CREATE TABLE info_history_pruning_metadata (
    id                                  SERIAL PRIMARY KEY,
    started_on                          timestamp with time zone NOT NULL DEFAULT now(),
    ended_on                            timestamp with time zone,
    successful                          boolean NOT NULL DEFAULT false,
    num_rows_processed                  integer NOT NULL DEFAULT 0,
    num_rows_pruned                     integer NOT NULL DEFAULT 0,
    query_int_start_date                timestamp with time zone NOT NULL,
    query_int_end_date                  timestamp with time zone NOT NULL
);

COMMIT;