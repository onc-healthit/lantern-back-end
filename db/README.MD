# Database Setup Scripts
The `.sql` files located in db/sql are mounted to the postgres service `docker-entrypoint-initdb.d` directory in the `docker-compose.yml` and `docker-compose-override.yml` files. From the postgres Docker image documentation https://hub.docker.com/_/postgres "If you would like to do additional initialization in an image derived from this one, add one or more *.sql, *.sql.gz, or *.sh scripts under /docker-entrypoint-initdb.d (creating the directory if necessary). After the entrypoint calls initdb to create the default postgres user and database, it will run any *.sql files, run any executable *.sh scripts, and source any non-executable *.sh scripts found in that directory to do further initialization before starting the service."

Changes to the database (new tables, new columns on exiisting tables etc..) should exist in `db/dbsetup.sql` as this file will setup the database from scratch.

# Database Migration
If changes are made to the database setup file `db/dbsetup.sql` then there should also be a migration script defined that when run will migrate an existing database to be compatible with the changes to `db/dbsetup.sql`.

Migration scripts should be placed in `db/migration/migratrions`. For more information about how to write migration scripts see the tutorial here https://github.com/golang-migrate/migrate/blob/master/database/postgres/TUTORIAL.md

## Migrate Down
If you find yourself in the unlikely scenario of needing to undo migrations, by way of the `down.sql` scripts in the `db/migration/migrations` directory, you can change the line that reads `m.Steps(2)` in `db/migration/cmd/main.go` to `m.Steps(-2)` and run `make migrate_database` from the base directory.