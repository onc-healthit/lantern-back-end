package main

import "github.com/spf13/viper"

func SetupConfig() error {
	var err error

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()

	err = viper.BindEnv("chplapikey")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbhost")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbport")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbsslmode")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbuser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbname")
	if err != nil {
		return err
	}

	viper.SetEnvPrefix("lantern_test")
	viper.AutomaticEnv()

	err = viper.BindEnv("dbuser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbname")
	if err != nil {
		return err
	}

	viper.SetDefault("dbhost", "localhost")
	viper.SetDefault("dbport", 5432)
	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpassword", "postgrespassword")
	viper.SetDefault("dbname", "lantern")
	viper.SetDefault("dbsslmode", "disable")

	return nil
}

func main() {
	SetupConfig()
	println("host: " + viper.GetString("dbhost"))
	println("dbname: " + viper.GetString("dbname"))
}
