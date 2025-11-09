package config

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var Conf Config

type Config struct {
	Rpc      RpcConfig      `yaml:"rpc"`
	Api      ApiConfig      `yaml:"api"`
	Database DatabaseConfig `yaml:"database"`
	Indexer  IndexerConfig  `yaml:"indexer"`
	Modules  ModulesConfig  `yaml:"modules"`
}

type RpcConfig struct {
	Url           string `yaml:"url"`
	Timeout       int    `yaml:"timeout"`
	RetryAttempts int    `yaml:"retry_attempts"`
	RetryDelay    int    `yaml:"retry_delay"`
}

type ApiConfig struct {
	Host           string           `yaml:"host"`
	Port           string           `yaml:"port"`
	Production     bool             `yaml:"production"`
	Admin          bool             `yaml:"admin"`
	Cors           CorsConfig       `yaml:"cors"`
	ReadTimeout    int              `yaml:"read_timeout"`
	WriteTimeout   int              `yaml:"write_timeout"`
	IdleTimeout    int              `yaml:"idle_timeout"`
	MaxHeaderBytes int              `yaml:"max_header_bytes"`
	Pagination     PaginationConfig `yaml:"pagination"`
}

type PaginationConfig struct {
	DefaultLimit int `yaml:"default_limit"`
	MaxLimit     int `yaml:"max_limit"`
	MaxOffset    int `yaml:"max_offset"`
}

type CorsConfig struct {
	AllowedOrigins []string `yaml:"allowed_origins"`
	AllowedMethods []string `yaml:"allowed_methods"`
	AllowedHeaders []string `yaml:"allowed_headers"`
}

type DatabaseConfig struct {
	Host               string `yaml:"host"`
	Port               string `yaml:"port"`
	User               string `yaml:"user"`
	Password           string `yaml:"password"`
	DBName             string `yaml:"dbname"`
	SSLMode            string `yaml:"sslmode"`
	MaxConnections     int    `yaml:"max_connections"`
	MaxIdleConnections int    `yaml:"max_idle_connections"`
	ConnectionLifetime int    `yaml:"connection_lifetime"`
	ConnectTimeout     int    `yaml:"connect_timeout"`
	StatementTimeout   int    `yaml:"statement_timeout"`
}

type IndexerConfig struct {
	BatchSize           int   `yaml:"batch_size"`
	PollInterval        int   `yaml:"poll_interval"`
	StartBlock          int64 `yaml:"start_block"`
	EnableReorgHandling bool  `yaml:"enable_reorg_handling"`
	MaxReorgDepth       int   `yaml:"max_reorg_depth"`
}

type ModulesConfig struct {
	TxGraph  TxGraphConfig  `yaml:"tx_graph"`
	TzeGraph TzeGraphConfig `yaml:"tze_graph"`
	Starks   StarksConfig   `yaml:"starks"`
	Accounts AccountsConfig `yaml:"accounts"`
}

type TxGraphConfig struct {
	Enabled       bool `yaml:"enabled"`
	MaxGraphDepth int  `yaml:"max_graph_depth"`
}

type TzeGraphConfig struct {
	Enabled             bool `yaml:"enabled"`
	MaxPreconditionSize int  `yaml:"max_precondition_size"`
}

type StarksConfig struct {
	Enabled       bool `yaml:"enabled"`
	IndexZtarknet bool `yaml:"index_ztarknet"`
}

type AccountsConfig struct {
	Enabled bool `yaml:"enabled"`
}

func InitConfig(configPath string) {
	log.Printf("Loading configuration from: %s", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	// Substitute environment variables in the config file
	configContent := string(data)
	configContent = expandEnvVars(configContent)

	err = yaml.Unmarshal([]byte(configContent), &Conf)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// Validate configuration
	if err := validateConfig(); err != nil {
		log.Fatalf("Configuration validation failed: %v", err)
	}

	log.Println("Configuration loaded successfully")
}

// expandEnvVars replaces ${VAR_NAME} patterns with environment variable values
func expandEnvVars(content string) string {
	// Match ${VAR_NAME} pattern
	re := regexp.MustCompile(`\$\{([A-Z_][A-Z0-9_]*)\}`)

	return re.ReplaceAllStringFunc(content, func(match string) string {
		// Extract variable name from ${VAR_NAME}
		varName := strings.TrimPrefix(match, "${")
		varName = strings.TrimSuffix(varName, "}")

		// Get environment variable value
		if value := os.Getenv(varName); value != "" {
			return value
		}

		// If environment variable is not set, log a warning and return empty string
		log.Printf("Warning: Environment variable %s is not set, using empty string", varName)
		return ""
	})
}

func ShouldConnectPostgres() bool {
	return Conf.Database.Host != "" && Conf.Database.Port != ""
}

func IsModuleEnabled(moduleName string) bool {
	switch moduleName {
	case "TX_GRAPH":
		return Conf.Modules.TxGraph.Enabled
	case "TZE_GRAPH":
		return Conf.Modules.TzeGraph.Enabled
	case "STARKS":
		return Conf.Modules.Starks.Enabled
	case "ACCOUNTS":
		return Conf.Modules.Accounts.Enabled
	default:
		return false
	}
}

// validateConfig validates the loaded configuration
func validateConfig() error {
	// Validate RPC configuration
	if Conf.Rpc.Url == "" {
		return fmt.Errorf("rpc.url is required")
	}
	if !strings.HasPrefix(Conf.Rpc.Url, "http://") && !strings.HasPrefix(Conf.Rpc.Url, "https://") {
		return fmt.Errorf("rpc.url must start with http:// or https://")
	}
	if Conf.Rpc.Timeout <= 0 {
		return fmt.Errorf("rpc.timeout must be greater than 0")
	}
	if Conf.Rpc.RetryAttempts < 0 {
		return fmt.Errorf("rpc.retry_attempts must be non-negative")
	}
	if Conf.Rpc.RetryDelay < 0 {
		return fmt.Errorf("rpc.retry_delay must be non-negative")
	}

	// Validate API configuration
	if Conf.Api.Host == "" {
		return fmt.Errorf("api.host is required")
	}
	if Conf.Api.Port == "" {
		return fmt.Errorf("api.port is required")
	}
	if Conf.Api.ReadTimeout <= 0 {
		return fmt.Errorf("api.read_timeout must be greater than 0")
	}
	if Conf.Api.WriteTimeout <= 0 {
		return fmt.Errorf("api.write_timeout must be greater than 0")
	}
	if Conf.Api.IdleTimeout <= 0 {
		return fmt.Errorf("api.idle_timeout must be greater than 0")
	}
	if Conf.Api.MaxHeaderBytes <= 0 {
		return fmt.Errorf("api.max_header_bytes must be greater than 0")
	}

	// Validate pagination configuration
	if Conf.Api.Pagination.DefaultLimit <= 0 {
		return fmt.Errorf("api.pagination.default_limit must be greater than 0")
	}
	if Conf.Api.Pagination.MaxLimit <= 0 {
		return fmt.Errorf("api.pagination.max_limit must be greater than 0")
	}
	if Conf.Api.Pagination.DefaultLimit > Conf.Api.Pagination.MaxLimit {
		return fmt.Errorf("api.pagination.default_limit must be less than or equal to max_limit")
	}
	if Conf.Api.Pagination.MaxOffset < 0 {
		return fmt.Errorf("api.pagination.max_offset must be non-negative")
	}

	// Validate CORS configuration (if provided)
	validMethods := map[string]bool{
		"GET": true, "POST": true, "PUT": true, "PATCH": true,
		"DELETE": true, "OPTIONS": true, "HEAD": true,
	}
	for _, method := range Conf.Api.Cors.AllowedMethods {
		if !validMethods[method] {
			return fmt.Errorf("api.cors.allowed_methods contains invalid method: %s", method)
		}
	}

	// Validate Database configuration (if database connection is enabled)
	if ShouldConnectPostgres() {
		if Conf.Database.User == "" {
			return fmt.Errorf("database.user is required when database connection is enabled")
		}
		if Conf.Database.Password == "" {
			return fmt.Errorf("database.password is required when database connection is enabled (use ${DB_PASSWORD} for environment variable substitution)")
		}
		if Conf.Database.DBName == "" {
			return fmt.Errorf("database.dbname is required when database connection is enabled")
		}

		// Validate SSL mode
		validSSLModes := map[string]bool{
			"disable":     true,
			"allow":       true,
			"prefer":      true,
			"require":     true,
			"verify-ca":   true,
			"verify-full": true,
		}
		if !validSSLModes[Conf.Database.SSLMode] {
			return fmt.Errorf("database.sslmode must be one of: disable, allow, prefer, require, verify-ca, verify-full")
		}

		// Validate connection pool settings
		if Conf.Database.MaxConnections <= 0 {
			return fmt.Errorf("database.max_connections must be greater than 0")
		}
		if Conf.Database.MaxIdleConnections < 0 {
			return fmt.Errorf("database.max_idle_connections must be non-negative")
		}
		if Conf.Database.MaxIdleConnections > Conf.Database.MaxConnections {
			return fmt.Errorf("database.max_idle_connections must be less than or equal to max_connections")
		}
		if Conf.Database.ConnectionLifetime <= 0 {
			return fmt.Errorf("database.connection_lifetime must be greater than 0")
		}

		// Validate timeout settings
		if Conf.Database.ConnectTimeout <= 0 {
			return fmt.Errorf("database.connect_timeout must be greater than 0")
		}
		if Conf.Database.StatementTimeout <= 0 {
			return fmt.Errorf("database.statement_timeout must be greater than 0")
		}
	}

	// Validate Indexer configuration
	if Conf.Indexer.BatchSize <= 0 {
		return fmt.Errorf("indexer.batch_size must be greater than 0")
	}
	if Conf.Indexer.PollInterval <= 0 {
		return fmt.Errorf("indexer.poll_interval must be greater than 0")
	}
	if Conf.Indexer.StartBlock < 0 {
		return fmt.Errorf("indexer.start_block must be non-negative")
	}
	if Conf.Indexer.MaxReorgDepth < 0 {
		return fmt.Errorf("indexer.max_reorg_depth must be non-negative")
	}

	// Validate Module configurations
	if Conf.Modules.TxGraph.Enabled {
		if Conf.Modules.TxGraph.MaxGraphDepth <= 0 {
			return fmt.Errorf("modules.tx_graph.max_graph_depth must be greater than 0")
		}
	}

	if Conf.Modules.TzeGraph.Enabled {
		if Conf.Modules.TzeGraph.MaxPreconditionSize <= 0 {
			return fmt.Errorf("modules.tze_graph.max_precondition_size must be greater than 0")
		}
	}

	return nil
}
