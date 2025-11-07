package config

import (
	"log"
	"os"

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
	Url      string `yaml:"url"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Timeout  int    `yaml:"timeout"`
}

type ApiConfig struct {
	Host       string     `yaml:"host"`
	Port       string     `yaml:"port"`
	Production bool       `yaml:"production"`
	Admin      bool       `yaml:"admin"`
	Cors       CorsConfig `yaml:"cors"`
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
}

type IndexerConfig struct {
	BatchSize           int  `yaml:"batch_size"`
	PollInterval        int  `yaml:"poll_interval"`
	StartBlock          int64 `yaml:"start_block"`
	EnableReorgHandling bool `yaml:"enable_reorg_handling"`
	MaxReorgDepth       int  `yaml:"max_reorg_depth"`
}

type ModulesConfig struct {
	TxGraph  ModuleConfig `yaml:"tx_graph"`
	TzeGraph ModuleConfig `yaml:"tze_graph"`
	Starks   ModuleConfig `yaml:"starks"`
	Accounts ModuleConfig `yaml:"accounts"`
}

type ModuleConfig struct {
	Enabled bool                   `yaml:"enabled"`
	Options map[string]interface{} `yaml:",inline"`
}

func InitConfig(configPath string) {
	log.Printf("Loading configuration from: %s", configPath)

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	log.Println("Configuration loaded successfully")
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
