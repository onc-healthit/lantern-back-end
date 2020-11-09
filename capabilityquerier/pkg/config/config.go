package config

import "github.com/spf13/viper"

// SetupConfig associates the application with all of the relevant configuration parameters
// for the application with the prefix 'lantern'.
func SetupConfig() error {
	var err error

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()

	// Queue Setup

	err = viper.BindEnv("quser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qhost")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qport")
	if err != nil {
		return err
	}
	err = viper.BindEnv("capquery_qname")
	if err != nil {
		return err
	}
	err = viper.BindEnv("capquery_numworkers")
	if err != nil {
		return err
	}
	err = viper.BindEnv("endptinfo_capquery_qname")
	if err != nil {
		return err
	}

	// Database setup

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

	// Export JSON file

	err = viper.BindEnv("exportfile_wait")
	if err != nil {
		return err
	}

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qhost", "localhost")
	viper.SetDefault("qport", "5672")
	viper.SetDefault("capquery_qname", "capability-statements")
	viper.SetDefault("capquery_numworkers", 10)
	viper.SetDefault("endptinfo_capquery_qname", "endpoints-to-capability")

	viper.SetDefault("dbhost", "localhost")
	viper.SetDefault("dbport", 5432)
	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpassword", "postgrespassword")
	viper.SetDefault("dbname", "lantern")
	viper.SetDefault("dbsslmode", "disable")

	viper.SetDefault("exportfile_wait", 300)

	return nil
}

// SetupConfigForTests associates the application with all of the relevant configuration parameters
// for the application and replaces the prefix 'lantern' with 'lantern_test' for the following
// environment variables:
// - quser
// - qpassword
// - capquery_qname
func SetupConfigForTests() error {
	var err error

	err = SetupConfig()
	if err != nil {
		return err
	}
	prevDbName := viper.GetString("dbname")
	prevQName := viper.GetString("capquery_qname")

	viper.SetEnvPrefix("lantern_test")
	viper.AutomaticEnv()

	err = viper.BindEnv("quser")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qpassword")
	if err != nil {
		return err
	}
	err = viper.BindEnv("qname")
	if err != nil {
		return err
	}
	err = viper.BindEnv("endptinfo_capquery_qname")
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

	viper.SetDefault("dbuser", "lantern")
	viper.SetDefault("dbpassword", "postgrespassword")
	viper.SetDefault("dbname", "lantern_test")

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qname", "test-queue")
	viper.SetDefault("endptinfo_capquery_qname", "test-endpoints-to-capability")

	if prevQName == viper.GetString("qname") {
		panic("Test queue and dev/prod queue must be different. Test queue: " + viper.GetString("qname") + ". Prod/Dev queue: " + prevQName)
	}
	if prevDbName == viper.GetString("dbname") {
		panic("Test database and dev/prod database must be different. Test database: " + viper.GetString("dbname") + ". Prod/Dev dataabse: " + prevDbName)
	}

	return nil
}
