package config

import "github.com/spf13/viper"

// SetupConfig associates the application with all of the relevant configuration parameters
// for the application with the prefix 'lantern'.
func SetupConfig() error {
	var err error

	// Endpoint Query Setup

	viper.SetEnvPrefix("lantern_endptqry")
	viper.AutomaticEnv()

	err = viper.BindEnv("port")
	if err != nil {
		return err
	}

	// Queue Setup

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
	err = viper.BindEnv("enptinfo_netstats_qname")
	if err != nil {
		return err
	}

	viper.SetDefault("port", 3333)

	viper.SetDefault("quser", "capabilityquerier")
	viper.SetDefault("qpassword", "capabilityquerier")
	viper.SetDefault("qhost", "localhost")
	viper.SetDefault("qport", "5672")
	viper.SetDefault("enptinfo_netstats_qname", "endpoints-to-netstats")

	return nil
}
