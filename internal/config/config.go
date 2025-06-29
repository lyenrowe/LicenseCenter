package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Security SecurityConfig `mapstructure:"security"`
	Captcha  CaptchaConfig  `mapstructure:"captcha"`
	Logging  LoggingConfig  `mapstructure:"logging"`
	System   SystemConfig   `mapstructure:"system"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Driver          string `mapstructure:"driver"`
	DSN             string `mapstructure:"dsn"`
	MaxConns        int    `mapstructure:"max_conns"`
	MaxIdleConns    int    `mapstructure:"max_idle_conns"`
	ConnMaxLifetime int    `mapstructure:"conn_max_lifetime"`
}

type SecurityConfig struct {
	JWTSecret           string `mapstructure:"jwt_secret"`
	SessionTimeout      int    `mapstructure:"session_timeout"`
	AdminSessionTimeout int    `mapstructure:"admin_session_timeout"`
	RSAKeySize          int    `mapstructure:"rsa_key_size"`
	ForceTOTP           bool   `mapstructure:"force_totp"` // 强制启用双因子认证
}

type CaptchaConfig struct {
	Enabled   bool   `mapstructure:"enabled"`
	SiteKey   string `mapstructure:"site_key"`
	SecretKey string `mapstructure:"secret_key"`
}

type LoggingConfig struct {
	Level         string `mapstructure:"level"`
	File          string `mapstructure:"file"`
	MaxSize       int    `mapstructure:"max_size"`
	MaxBackups    int    `mapstructure:"max_backups"`
	MaxAge        int    `mapstructure:"max_age"`
	GinMode       string `mapstructure:"gin_mode"`        // gin框架模式: debug, release, test
	EnableHTTPLog bool   `mapstructure:"enable_http_log"` // 是否启用HTTP请求日志
}

type SystemConfig struct {
	MaxBindFilesPerRequest int    `mapstructure:"max_bind_files_per_request"`
	BackupRetentionDays    int    `mapstructure:"backup_retention_days"`
	DataDir                string `mapstructure:"data_dir"`
	UploadDir              string `mapstructure:"upload_dir"`
}

var AppConfig *Config

// LoadConfig 加载配置文件
func LoadConfig(configPath string) error {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")

	// 设置默认值
	setDefaults()

	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("读取配置文件失败: %w", err)
	}

	AppConfig = &Config{}
	if err := viper.Unmarshal(AppConfig); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}

	log.Printf("配置加载成功: %s", viper.ConfigFileUsed())
	return nil
}

func setDefaults() {
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.mode", "debug")

	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.dsn", "./data/license.db")
	viper.SetDefault("database.max_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", 3600)

	viper.SetDefault("security.session_timeout", 3600)
	viper.SetDefault("security.admin_session_timeout", 1800)
	viper.SetDefault("security.rsa_key_size", 2048)
	viper.SetDefault("security.force_totp", false)

	viper.SetDefault("captcha.enabled", true)

	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.file", "./logs/app.log")
	viper.SetDefault("logging.max_size", 100)
	viper.SetDefault("logging.max_backups", 3)
	viper.SetDefault("logging.max_age", 30)
	viper.SetDefault("logging.gin_mode", "release")
	viper.SetDefault("logging.enable_http_log", true)

	viper.SetDefault("system.max_bind_files_per_request", 10)
	viper.SetDefault("system.backup_retention_days", 30)
	viper.SetDefault("system.data_dir", "./data")
	viper.SetDefault("system.upload_dir", "./uploads")
}
