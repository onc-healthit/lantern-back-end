package config

import "github.com/spf13/viper"

// SetupConfig associates the application with all of the relevant configuration parameters
// for the application with the prefix 'lantern'.
func SetupConfig() error {
	var err error

	viper.SetEnvPrefix("lantern")
	viper.AutomaticEnv()

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

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qhost", "localhost")
	viper.SetDefault("qport", "5672")
	viper.SetDefault("capquery_qname", "capability-statements")
	viper.SetDefault("capquery_numworkers", 10)
	viper.SetDefault("endptinfo_capquery_qname", "endpoints-to-capability")

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

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qname", "test-queue")
	viper.SetDefault("endptinfo_capquery_qname", "test-endpoints-to-capability")

	if prevQName == viper.GetString("qname") {
		panic("Test queue and dev/prod queue must be different. Test queue: " + viper.GetString("qname") + ". Prod/Dev queue: " + prevQName)
	}

	return nil
}
