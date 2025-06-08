package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"stateless-load-balancer/balancer"
	"syscall"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	// Basic configuration
	Port       string `mapstructure:"PORT"`
	Host       string `mapstructure:"HOST"`
	NodeFilter string `mapstructure:"NODE_FILTER"`
	ASGName    string `mapstructure:"ASG_NAME"`
	TagName    string `mapstructure:"TAG_NAME"`
	TagValue   string `mapstructure:"TAG_VALUE"`
	Region     string `mapstructure:"AWS_REGION"`

	// Health check configuration
	HealthCheckPeriod  int    `mapstructure:"HEALTH_CHECK_PERIOD"`
	HealthCheckPath    string `mapstructure:"HEALTH_CHECK_PATH"`
	HealthCheckTimeout int    `mapstructure:"HEALTH_CHECK_TIMEOUT"`

	// Node discovery configuration
	NodeCheckPeriod int `mapstructure:"NODE_CHECK_PERIOD"`
	TargetPort      int `mapstructure:"TARGET_PORT"`

	// Connection pool configuration
	PoolSize    int `mapstructure:"POOL_SIZE"`
	PoolTimeout int `mapstructure:"POOL_TIMEOUT"`

	// Circuit breaker configuration
	CircuitBreakerThreshold int `mapstructure:"CIRCUIT_BREAKER_THRESHOLD"`
	CircuitBreakerTimeout   int `mapstructure:"CIRCUIT_BREAKER_TIMEOUT"`

	// Graceful shutdown configuration
	ShutdownTimeout int `mapstructure:"SHUTDOWN_TIMEOUT"`

	// Logging configuration
	LogConnections bool `mapstructure:"LOG_CONNECTIONS"`
}

func initConfig() Config {
	viper.AutomaticEnv()

	// Basic configuration defaults
	viper.SetDefault("PORT", "8080")
	viper.SetDefault("HOST", "0.0.0.0")
	viper.SetDefault("NODE_FILTER", "asg")
	viper.SetDefault("ASG_NAME", "worker-nodes")
	viper.SetDefault("TAG_NAME", "Name")
	viper.SetDefault("TAG_VALUE", "worker-node")
	viper.SetDefault("AWS_REGION", "eu-west-1")

	// Health check configuration defaults
	viper.SetDefault("HEALTH_CHECK_PERIOD", 2)
	viper.SetDefault("HEALTH_CHECK_PATH", "/ping")
	viper.SetDefault("HEALTH_CHECK_TIMEOUT", 2)

	// Node discovery configuration defaults
	viper.SetDefault("NODE_CHECK_PERIOD", 10)
	viper.SetDefault("TARGET_PORT", 30080)

	// Connection pool configuration defaults
	viper.SetDefault("POOL_SIZE", 100)
	viper.SetDefault("POOL_TIMEOUT", 2)

	// Circuit breaker configuration defaults
	viper.SetDefault("CIRCUIT_BREAKER_THRESHOLD", 5)
	viper.SetDefault("CIRCUIT_BREAKER_TIMEOUT", 5)

	// Graceful shutdown configuration defaults
	viper.SetDefault("SHUTDOWN_TIMEOUT", 30)

	// Logging configuration defaults
	viper.SetDefault("LOG_CONNECTIONS", false)

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
		config.TargetPort,
		config.NodeCheckPeriod,
		config.HealthCheckPeriod,
		config.HealthCheckTimeout,
		config.PoolSize,
		config.PoolTimeout,
		config.CircuitBreakerThreshold,
		config.CircuitBreakerTimeout,
		config.LogConnections,
	)

	// Start the health check routine
	go lb.StartHealthChecks()

	// Start the TCP load balancer
	address := config.Host + ":" + config.Port
	go func() {
		if err := lb.StartTCPProxy(address); err != nil {
			log.Fatalf("Failed to start TCP proxy: %v", err)
		}
	}()

	// Set up graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down load balancer...")

	// Create a context with timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(config.ShutdownTimeout)*time.Second)
	defer cancel()

	if err := lb.Shutdown(ctx); err != nil {
		log.Fatalf("Error during shutdown: %v", err)
	}

	log.Println("Load balancer shutdown complete")
}
