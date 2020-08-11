\set lantern `echo $POSTGRES_DB`

-- remove any public access
REVOKE CONNECT ON DATABASE :lantern FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA pg_catalog FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA pg_catalog FROM PUBLIC;

-- create read only and read/write group roles
CREATE ROLE readonly;
GRANT CONNECT ON DATABASE :lantern TO readonly;
GRANT USAGE ON SCHEMA public TO readonly;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO readonly;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT ON TABLES TO readonly; -- grants permissions on new tables

CREATE ROLE readwrite;
GRANT CONNECT ON DATABASE :lantern TO readwrite;
GRANT USAGE ON SCHEMA public TO readwrite;
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO readwrite;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO readwrite; -- grants permissions on new tables
GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO readwrite;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE ON SEQUENCES TO readwrite; -- grants permissions on new sequences
