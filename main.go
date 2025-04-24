package main

import (
	"log"
	"stateless-load-balancer/balancer"

	"github.com/spf13/viper"
)

type Config struct {
	Port              string `mapstructure:"PORT"`
	Host              string `mapstructure:"HOST"`
	NodeFilter        string `mapstructure:"NODE_FILTER"`
	ASGName           string `mapstructure:"ASG_NAME"`
	TagName           string `mapstructure:"TAG_NAME"`
	TagValue          string `mapstructure:"TAG_VALUE"`
	NodeCheckPeriod   int    `mapstructure:"NODE_CHECK_PERIOD"`
	HealthCheckPeriod int    `mapstructure:"HEALTH_CHECK_PERIOD"`
	HealthCheckPath   string `mapstructure:"HEALTH_CHECK_PATH"`
	HealthCheckPort   int    `mapstructure:"TARGET_PORT"`
	Region            string `mapstructure:"AWS_REGION"`
}

func initConfig() Config {
	viper.AutomaticEnv()

	viper.SetDefault("PORT", "8080")
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("NODE_CHECK_PERIOD", 10)
	viper.SetDefault("ASG_NAME", "worker-nodes")
	viper.SetDefault("TAG_NAME", "Name")
	viper.SetDefault("TAG_VALUE", "worker-node")
	viper.SetDefault("NODE_FILTER", "asg")
	viper.SetDefault("HEALTH_CHECK_PERIOD", 2)
	viper.SetDefault("HEALTH_CHECK_PATH", "/ping")
	viper.SetDefault("TARGET_PORT", 30080)
	viper.SetDefault("AWS_REGION", "eu-west-1")

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Unable to unmarshal configuration: %v", err)
	}

	return config
}

func main() {
	config := initConfig()

	// Initialize the load balancer
	lb := balancer.NewLoadBalancer(
		config.NodeFilter,
		config.TagName,
		config.TagValue,
		config.ASGName,
		config.HealthCheckPath,
		config.Region,
		config.HealthCheckPort,
		config.NodeCheckPeriod,
		config.HealthCheckPeriod,
	)

	// Start the health check routine
	go lb.StartHealthChecks()

	// Start the TCP load balancer
	address := config.Host + ":" + config.Port
	lb.StartTCPProxy(address)
}
