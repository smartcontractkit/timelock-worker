package cli

import (
	"os"
	"strconv"

	"github.com/spf13/viper"
)

// Config holds the timelock.env configuration structure.
type Config struct {
	NodeURL          string `mapstructure:"NODE_URL"`
	TimelockAddress  string `mapstructure:"TIMELOCK_ADDRESS"`
	CallProxyAddress string `mapstructure:"CALL_PROXY_ADDRESS"`
	PrivateKey       string `mapstructure:"PRIVATE_KEY"`
	FromBlock        int64  `mapstructure:"FROM_BLOCK"`
	PollPeriod       int64  `mapstructure:"POLL_PERIOD"`
}

// NewTimelockCLI return a new Timelock instance configured.
func NewTimelockCLI() (*Config, error) {
	viper.AddConfigPath(".")
	viper.SetConfigName("timelock")
	viper.SetConfigType("env")

	var c Config

	// Read timelock.env and unmarshal the config into w.
	// timelock.env defines the default configuration
	// provided to the cli.
	// Do not return errors here, as the CLI can be completely functional
	// only by providing env vars.
	_ = viper.ReadInConfig()
	_ = viper.Unmarshal(&c)

	// Environment variables have precedence over timelock.env
	// Get them and apply it into worker whenever they exist.
	if os.Getenv("NODE_URL") != "" {
		c.NodeURL = os.Getenv("NODE_URL")
	}

	if os.Getenv("TIMELOCK_ADDRESS") != "" {
		c.TimelockAddress = os.Getenv("TIMELOCK_ADDRESS")
	}

	if os.Getenv("CALL_PROXY_ADDRESS") != "" {
		c.CallProxyAddress = os.Getenv("CALL_PROXY_ADDRESS")
	}

	if os.Getenv("PRIVATE_KEY") != "" {
		c.PrivateKey = os.Getenv("PRIVATE_KEY")
	}

	if os.Getenv("FROM_BLOCK") != "" {
		fb, err := strconv.Atoi(os.Getenv("FROM_BLOCK"))
		if err == nil {
			c.FromBlock = int64(fb)
		}
	}

	if os.Getenv("POLL_PERIOD") != "" {
		pp, err := strconv.Atoi(os.Getenv("POLL_PERIOD"))
		if err == nil {
			c.PollPeriod = int64(pp)
		}
	}

	return &c, nil
}
