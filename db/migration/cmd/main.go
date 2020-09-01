package main

import (
	"database/sql"
	"fmt"
	"github.com/spf13/viper"
    _ "github.com/lib/pq"
    "github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	log "github.com/sirupsen/logrus"
    _ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()
	host := viper.GetString("dbhost")
	port := viper.GetInt("dbport")
	user := viper.GetString("dbuser")
	password := viper.GetString("dbpassword")
	dbname := viper.GetString("dbname")
	sslmode := viper.GetString("dbsslmode")

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
	"password=%s dbname=%s sslmode=%s",
	host, port, user, password, dbname, sslmode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Error("endpoint URL parsing error: ", err.Error())
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Error("endpoint URL parsing error: ", err.Error())
	}

    m, err := migrate.NewWithDatabaseInstance(
        "file://./migrations",
        "postgres", driver)
    m.Steps(2)
}