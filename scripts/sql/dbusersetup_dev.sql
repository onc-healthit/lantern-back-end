-- this file sets up a readonly user and a read/write user for a dev environment.
-- in a prod environment, a user should be setup for each service or user using
-- the database.

\set readonly_user `echo $LANTERN_DBUSER_READONLY`
\set readonly_pw `echo $LANTERN_DBPASSWORD_READONLY`
\set readwrite_user `echo $LANTERN_DBUSER_READWRITE`
\set readwrite_pw `echo $LANTERN_DBPASSWORD_READWRITE`

-- add a readonly user and grant readonly permissions
CREATE ROLE :readonly_user LOGIN PASSWORD :'readonly_pw';
GRANT readonly to :readonly_user;

-- add a readwrite user and grant readwrite permissions
CREATE ROLE :readwrite_user LOGIN PASSWORD :'readwrite_pw';
GRANT readwrite to :readwrite_user;