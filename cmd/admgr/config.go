package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

// Config defines the structure of the config object
// it defines information about server details and
// postgres db details including the credentials
type Config struct {
	InstanceId string            `json:"instance_id" mapstructure:"instance_id"`
	Host       string            `json:"host" mapstructure:"host"`
	Port       string            `json:"port" mapstructure:"port"`
	Redis      RedisConf         `json:"redis" mapstructure:"redis"`
	DB         DBConf            `json:"db" mapstructure:"db"`
	AcLogger   AsyncommLoggerCnf `json:"asyncomm_logger" mapstructure:"asyncomm_logger"`
	Logger     struct {
		Level          string `json:"level" mapstructure:"level"`
		FullTimestamp  bool   `json:"full_timestamp" mapstructure:"full_timestamp"`
		OutputFilePath string `json:"output_file_path" mapstructure:"output_file_path"`
	} `json:"logger" mapstructure:"logger"`
}

type RedisConf struct {
	Host            string        `json:"host" mapstructure:"host"`
	Port            string        `json:"port" mapstructure:"port"`
	Username        string        `json:"username" mapstructure:"username"`
	Password        string        `json:"password" mapstructure:"password"`
	DB              int           `json:"db"`
	MaxRetries      int           `json:"max_retries" mapstructure:"max_retries"`
	MinRetryBackoff time.Duration `json:"min_retry_backoff" mapstructure:"min_retry_backoff"`
	MaxRetryBackoff time.Duration `json:"max_retry_backoff" mapstructure:"max_retry_backoff"`
	PoolSize        int           `json:"pool_size" mapstructure:"pool_size"`
	MinIdleConns    int           `json:"min_idle_conns" mapstructure:"min_idle_conns"`
	IdleTimeout     time.Duration `json:"idle_timeout" mapstructure:"idle_timeout"`
	ReadTimeout     time.Duration `json:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout" mapstructure:"write_timeout"`
	TLSConfig       *tls.Config   `json:"tls_config" mapstructure:"tls_config"`
}

type DBConf struct {
	Host     string `json:"host" mapstructure:"host"`
	Port     string `json:"port" mapstructure:"port"`
	Name     string `json:"name" mapstructure:"name"`
	Username string `json:"username" mapstructure:"username"`
	Password string `json:"password" mapstructure:"password"`
}

type AsyncommLoggerCnf struct {
	Level          string `json:"level" mapstructure:"level"`
	OutputFilePath string `json:"output_file_path" mapstructure:"output_file_path"`
}

var config *Config

// InitializeConfig makes use of viper library to initialize
// config from multiple sources such as json, yaml, toml and
// even environment variables, it returns a pointer to Config
func InitializeConfig() *Config {
	if config != nil {
		return config
	}

	// set the file name of the configurations file
	viper.SetConfigName("config")

	// set the path to look for the configurations file
	//viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/admgr")

	// enable VIPER to read Environment Variables
	viper.AutomaticEnv()

	viper.SetConfigType("yml")

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("error reading config file - %s\n", err)
		os.Exit(1)
	}

	// Set undefined variables
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", "10001")
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.username", "")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.full_timestamp", true)

	err := viper.Unmarshal(&config)
	if err != nil {
		panic(fmt.Sprintf("unable to decode config file : %v", err))
	}
	log.Printf("Config: %v", config)
	return config
}
