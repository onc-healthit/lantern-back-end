package config

import "github.com/spf13/viper"

// SetupConfig associates the application with all of the relevant configuration parameters
// for the application with the prefix 'lantern'.
func SetupConfig() error {
	var err error

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()

	err = viper.BindEnv("dbhost")
	if err != nil {
		return err
	}
	err = viper.BindEnv("dbport")
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
	err = viper.BindEnv("dbsslmode")
	if err != nil {
		return err
	}
	err = viper.BindEnv("chplapikey")
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

// SetupConfigForTests associates the application with all of the relevant configuration parameters
// for the application and replaces the prefix 'lantern' with 'lantern_test' for the following
// environment variables:
// - dbuser
// - dbpassword
// - dbname
func SetupConfigForTests() error {
	var err error

	err = SetupConfig()
	if err != nil {
		return err
	}

	prevDbName := viper.GetString("dbname")

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

	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpassword", "postgrespassword")
	viper.SetDefault("dbname", "lantern_test")

	if prevDbName == viper.GetString("dbname") {
		panic("Test database and dev/prod database must be different. Test database: " + viper.GetString("dbname") + ". Prod/Dev dataabse: " + prevDbName)
	}

	return nil
}
