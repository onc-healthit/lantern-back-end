CREATE DATABASE lantern_test;

\c lantern_test

\i /docker-entrypoint-initdb.d/dbsetup.sql

\c postgres