\set lantern `echo $POSTGRES_DB`
\set readonly_user `echo $LANTERN_DBUSER_READONLY`
\set readonly_pw `echo $LANTERN_DBPASSWORD_READONLY`

-- remove any public access
REVOKE CONNECT ON DATABASE :lantern FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM PUBLIC;

-- add a readonly user and grant readonly permissions
CREATE ROLE :readonly_user LOGIN PASSWORD :'readonly_pw';
GRANT CONNECT ON DATABASE lantern TO :readonly_user;
GRANT USAGE ON SCHEMA public TO :readonly_user;
GRANT SELECT ON ALL TABLES IN SCHEMA public TO :readonly_user;