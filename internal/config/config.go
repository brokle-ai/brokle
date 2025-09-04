// Package config provides configuration management for the Brokle platform.
//
// Configuration is loaded from multiple sources in this order:
// 1. Configuration files (YAML)
// 2. Environment variables
// 3. Command line flags (if applicable)
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config represents the complete application configuration.
type Config struct {
	Environment string         `mapstructure:"environment"`
	App         AppConfig      `mapstructure:"app"`
	Server      ServerConfig   `mapstructure:"server"`
	CORS        CORSConfig     `mapstructure:"cors"`
	Database    DatabaseConfig `mapstructure:"database"`
	ClickHouse  ClickHouseConfig `mapstructure:"clickhouse"`
	Redis       RedisConfig    `mapstructure:"redis"`
	JWT         JWTConfig      `mapstructure:"jwt"`
	Logging     LoggingConfig  `mapstructure:"logging"`
	External    ExternalConfig `mapstructure:"external"`
	Features    FeatureConfig  `mapstructure:"features"`
	Monitoring  MonitoringConfig `mapstructure:"monitoring"`
	Enterprise  EnterpriseConfig `mapstructure:"enterprise"`
}

// AppConfig contains application-level configuration.
type AppConfig struct {
	Version string `mapstructure:"version"`
	Name    string `mapstructure:"name"`
}

// CORSConfig contains CORS configuration.
type CORSConfig struct {
	AllowedOrigins []string `mapstructure:"allowed_origins"`
	AllowedMethods []string `mapstructure:"allowed_methods"`
	AllowedHeaders []string `mapstructure:"allowed_headers"`
}

// ServerConfig contains HTTP and WebSocket server configuration.
type ServerConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	Environment     string        `mapstructure:"environment"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	MaxRequestSize  int64         `mapstructure:"max_request_size"`
	EnableCORS      bool          `mapstructure:"enable_cors"`
	CORSOrigins     []string      `mapstructure:"cors_origins"`
	TrustedProxies  []string      `mapstructure:"trusted_proxies"`
}

// DatabaseConfig contains PostgreSQL database configuration.
type DatabaseConfig struct {
	URL             string        `mapstructure:"url"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	SSLMode         string        `mapstructure:"ssl_mode"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime time.Duration `mapstructure:"conn_max_idle_time"`
}

// ClickHouseConfig contains ClickHouse database configuration.
type ClickHouseConfig struct {
	URL             string        `mapstructure:"url"`
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	Database        string        `mapstructure:"database"`
	MaxOpenConns    int           `mapstructure:"max_open_conns"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
}

// RedisConfig contains Redis configuration.
type RedisConfig struct {
	URL          string        `mapstructure:"url"`
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	Database     int           `mapstructure:"database"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout"`
	MaxRetries   int           `mapstructure:"max_retries"`
}

// JWTConfig contains JWT token configuration.
type JWTConfig struct {
	PrivateKey       string        `mapstructure:"private_key"`
	PublicKey        string        `mapstructure:"public_key"`
	Issuer           string        `mapstructure:"issuer"`
	AccessTokenTTL   time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL  time.Duration `mapstructure:"refresh_token_ttl"`
	APIKeyTokenTTL   time.Duration `mapstructure:"api_key_token_ttl"`
	Algorithm        string        `mapstructure:"algorithm"`
}

// LoggingConfig contains logging configuration.
type LoggingConfig struct {
	Level      string `mapstructure:"level"`      // debug, info, warn, error
	Format     string `mapstructure:"format"`     // json, text
	Output     string `mapstructure:"output"`     // stdout, stderr, file
	File       string `mapstructure:"file"`       // file path if output=file
	MaxSize    int    `mapstructure:"max_size"`   // megabytes
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`    // days
}

// ExternalConfig contains external service configurations.
type ExternalConfig struct {
	OpenAI    OpenAIConfig    `mapstructure:"openai"`
	Anthropic AnthropicConfig `mapstructure:"anthropic"`
	Cohere    CohereConfig    `mapstructure:"cohere"`
	Stripe    StripeConfig    `mapstructure:"stripe"`
	Email     EmailConfig     `mapstructure:"email"`
}

// OpenAIConfig contains OpenAI API configuration.
type OpenAIConfig struct {
	APIKey     string        `mapstructure:"api_key"`
	BaseURL    string        `mapstructure:"base_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

// AnthropicConfig contains Anthropic API configuration.
type AnthropicConfig struct {
	APIKey     string        `mapstructure:"api_key"`
	BaseURL    string        `mapstructure:"base_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

// CohereConfig contains Cohere API configuration.
type CohereConfig struct {
	APIKey     string        `mapstructure:"api_key"`
	BaseURL    string        `mapstructure:"base_url"`
	Timeout    time.Duration `mapstructure:"timeout"`
	MaxRetries int           `mapstructure:"max_retries"`
}

// StripeConfig contains Stripe API configuration.
type StripeConfig struct {
	SecretKey      string `mapstructure:"secret_key"`
	PublishableKey string `mapstructure:"publishable_key"`
	WebhookSecret  string `mapstructure:"webhook_secret"`
	Environment    string `mapstructure:"environment"` // test, live
}

// EmailConfig contains email service configuration.
type EmailConfig struct {
	Provider   string `mapstructure:"provider"`   // smtp, sendgrid, mailgun
	SMTPHost   string `mapstructure:"smtp_host"`
	SMTPPort   int    `mapstructure:"smtp_port"`
	Username   string `mapstructure:"username"`
	Password   string `mapstructure:"password"`
	FromEmail  string `mapstructure:"from_email"`
	FromName   string `mapstructure:"from_name"`
	ReplyEmail string `mapstructure:"reply_email"`
}

// FeatureConfig contains feature flag configuration.
type FeatureConfig struct {
	SemanticCaching   bool `mapstructure:"semantic_caching"`
	RealTimeMetrics   bool `mapstructure:"real_time_metrics"`
	MLRouting         bool `mapstructure:"ml_routing"`
	CustomModels      bool `mapstructure:"custom_models"`
	MultiModal        bool `mapstructure:"multi_modal"`
	BackgroundJobs    bool `mapstructure:"background_jobs"`
	RateLimiting      bool `mapstructure:"rate_limiting"`
	AuditLogging      bool `mapstructure:"audit_logging"`
}

// MonitoringConfig contains monitoring and observability configuration.
type MonitoringConfig struct {
	Enabled        bool          `mapstructure:"enabled"`
	PrometheusPort int           `mapstructure:"prometheus_port"`
	MetricsPath    string        `mapstructure:"metrics_path"`
	JaegerEndpoint string        `mapstructure:"jaeger_endpoint"`
	SampleRate     float64       `mapstructure:"sample_rate"`
	FlushInterval  time.Duration `mapstructure:"flush_interval"`
}


// Validate validates the main configuration and all sub-configurations.
func (c *Config) Validate() error {
	if err := c.Server.Validate(); err != nil {
		return fmt.Errorf("server config validation failed: %w", err)
	}
	
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}
	
	if err := c.ClickHouse.Validate(); err != nil {
		return fmt.Errorf("clickhouse config validation failed: %w", err)
	}
	
	if err := c.Redis.Validate(); err != nil {
		return fmt.Errorf("redis config validation failed: %w", err)
	}
	
	if err := c.JWT.Validate(); err != nil {
		return fmt.Errorf("jwt config validation failed: %w", err)
	}
	
	if err := c.Logging.Validate(); err != nil {
		return fmt.Errorf("logging config validation failed: %w", err)
	}
	
	if err := c.External.Validate(); err != nil {
		return fmt.Errorf("external config validation failed: %w", err)
	}
	
	if err := c.Features.Validate(); err != nil {
		return fmt.Errorf("features config validation failed: %w", err)
	}
	
	if err := c.Monitoring.Validate(); err != nil {
		return fmt.Errorf("monitoring config validation failed: %w", err)
	}
	
	if err := c.Enterprise.Validate(); err != nil {
		return fmt.Errorf("enterprise config validation failed: %w", err)
	}
	
	return nil
}

// Validate validates server configuration.
func (sc *ServerConfig) Validate() error {
	if sc.Port <= 0 || sc.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", sc.Port)
	}
	
	if sc.Host == "" {
		return fmt.Errorf("host cannot be empty")
	}
	
	if sc.ReadTimeout < 0 {
		return fmt.Errorf("read_timeout cannot be negative")
	}
	
	if sc.WriteTimeout < 0 {
		return fmt.Errorf("write_timeout cannot be negative")
	}
	
	if sc.MaxRequestSize <= 0 {
		return fmt.Errorf("max_request_size must be positive")
	}
	
	return nil
}

// Validate validates database configuration.
func (dc *DatabaseConfig) Validate() error {
	// If URL is provided, minimal validation
	if dc.URL != "" {
		// URL takes precedence, minimal validation
		if dc.MaxOpenConns < 0 {
			return fmt.Errorf("max_open_conns cannot be negative")
		}
		
		if dc.MaxIdleConns < 0 {
			return fmt.Errorf("max_idle_conns cannot be negative")
		}
		
		return nil
	}
	
	// If no URL, validate individual fields
	if dc.Host == "" {
		return fmt.Errorf("either url or host must be provided")
	}
	
	if dc.Port <= 0 || dc.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be 1-65535)", dc.Port)
	}
	
	if dc.User == "" {
		return fmt.Errorf("user cannot be empty when using individual fields")
	}
	
	if dc.Database == "" {
		return fmt.Errorf("database name cannot be empty when using individual fields")
	}
	
	if dc.MaxOpenConns < 0 {
		return fmt.Errorf("max_open_conns cannot be negative")
	}
	
	if dc.MaxIdleConns < 0 {
		return fmt.Errorf("max_idle_conns cannot be negative")
	}
	
	return nil
}

// Validate validates ClickHouse configuration.
func (cc *ClickHouseConfig) Validate() error {
	// If URL is provided, minimal validation
	if cc.URL != "" {
		return nil // URL takes precedence
	}
	
	// If no URL, validate individual fields
	if cc.Host == "" {
		return fmt.Errorf("either url or host must be provided for clickhouse")
	}
	
	if cc.Port <= 0 || cc.Port > 65535 {
		return fmt.Errorf("invalid clickhouse port: %d (must be 1-65535)", cc.Port)
	}
	
	if cc.Database == "" {
		return fmt.Errorf("clickhouse database name cannot be empty when using individual fields")
	}
	
	return nil
}

// Validate validates Redis configuration.
func (rc *RedisConfig) Validate() error {
	// If URL is provided, minimal validation
	if rc.URL != "" {
		// URL takes precedence, minimal validation
		if rc.PoolSize < 0 {
			return fmt.Errorf("pool_size cannot be negative")
		}
		
		return nil
	}
	
	// If no URL, validate individual fields
	if rc.Host == "" {
		return fmt.Errorf("either url or host must be provided for redis")
	}
	
	if rc.Port <= 0 || rc.Port > 65535 {
		return fmt.Errorf("invalid redis port: %d (must be 1-65535)", rc.Port)
	}
	
	if rc.Database < 0 || rc.Database > 15 {
		return fmt.Errorf("invalid redis database number: %d (must be 0-15)", rc.Database)
	}
	
	if rc.PoolSize < 0 {
		return fmt.Errorf("pool_size cannot be negative")
	}
	
	return nil
}

// Validate validates JWT configuration.
func (jc *JWTConfig) Validate() error {
	if jc.PrivateKey == "" {
		return fmt.Errorf("private_key is required")
	}
	
	if jc.PublicKey == "" {
		return fmt.Errorf("public_key is required")
	}
	
	if jc.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	
	if jc.AccessTokenTTL <= 0 {
		return fmt.Errorf("access_token_ttl must be positive")
	}
	
	if jc.RefreshTokenTTL <= 0 {
		return fmt.Errorf("refresh_token_ttl must be positive")
	}
	
	validAlgorithms := []string{"RS256", "RS384", "RS512", "HS256", "HS384", "HS512"}
	isValid := false
	for _, alg := range validAlgorithms {
		if jc.Algorithm == alg {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid algorithm: %s (must be one of %v)", jc.Algorithm, validAlgorithms)
	}
	
	return nil
}

// Validate validates logging configuration.
func (lc *LoggingConfig) Validate() error {
	validLevels := []string{"debug", "info", "warn", "error"}
	isValid := false
	for _, level := range validLevels {
		if lc.Level == level {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid log level: %s (must be one of %v)", lc.Level, validLevels)
	}
	
	validFormats := []string{"json", "text"}
	isValid = false
	for _, format := range validFormats {
		if lc.Format == format {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid log format: %s (must be one of %v)", lc.Format, validFormats)
	}
	
	validOutputs := []string{"stdout", "stderr", "file"}
	isValid = false
	for _, output := range validOutputs {
		if lc.Output == output {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid log output: %s (must be one of %v)", lc.Output, validOutputs)
	}
	
	if lc.Output == "file" && lc.File == "" {
		return fmt.Errorf("file path is required when output is 'file'")
	}
	
	return nil
}

// Validate validates external services configuration.
func (ec *ExternalConfig) Validate() error {
	if err := ec.OpenAI.Validate(); err != nil {
		return fmt.Errorf("openai config: %w", err)
	}
	
	if err := ec.Anthropic.Validate(); err != nil {
		return fmt.Errorf("anthropic config: %w", err)
	}
	
	if err := ec.Cohere.Validate(); err != nil {
		return fmt.Errorf("cohere config: %w", err)
	}
	
	if err := ec.Stripe.Validate(); err != nil {
		return fmt.Errorf("stripe config: %w", err)
	}
	
	if err := ec.Email.Validate(); err != nil {
		return fmt.Errorf("email config: %w", err)
	}
	
	return nil
}

// Validate validates OpenAI configuration.
func (oc *OpenAIConfig) Validate() error {
	if oc.APIKey != "" && oc.BaseURL == "" {
		return fmt.Errorf("base_url is required when api_key is provided")
	}
	
	if oc.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	
	return nil
}

// Validate validates Anthropic configuration.
func (ac *AnthropicConfig) Validate() error {
	if ac.APIKey != "" && ac.BaseURL == "" {
		return fmt.Errorf("base_url is required when api_key is provided")
	}
	
	if ac.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	
	return nil
}

// Validate validates Cohere configuration.
func (cc *CohereConfig) Validate() error {
	if cc.APIKey != "" && cc.BaseURL == "" {
		return fmt.Errorf("base_url is required when api_key is provided")
	}
	
	if cc.MaxRetries < 0 {
		return fmt.Errorf("max_retries cannot be negative")
	}
	
	return nil
}

// Validate validates Stripe configuration.
func (sc *StripeConfig) Validate() error {
	if sc.SecretKey != "" {
		validEnvs := []string{"test", "live", ""}
		isValid := false
		for _, env := range validEnvs {
			if sc.Environment == env {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid stripe environment: %s (must be 'test' or 'live')", sc.Environment)
		}
	}
	
	return nil
}

// Validate validates email configuration.
func (ec *EmailConfig) Validate() error {
	if ec.Provider != "" {
		validProviders := []string{"smtp", "sendgrid", "mailgun"}
		isValid := false
		for _, provider := range validProviders {
			if ec.Provider == provider {
				isValid = true
				break
			}
		}
		if !isValid {
			return fmt.Errorf("invalid email provider: %s (must be one of %v)", ec.Provider, validProviders)
		}
		
		if ec.Provider == "smtp" {
			if ec.SMTPHost == "" {
				return fmt.Errorf("smtp_host is required for SMTP provider")
			}
			if ec.SMTPPort <= 0 || ec.SMTPPort > 65535 {
				return fmt.Errorf("invalid smtp_port: %d", ec.SMTPPort)
			}
		}
		
		if ec.FromEmail == "" {
			return fmt.Errorf("from_email is required")
		}
	}
	
	return nil
}

// Validate validates feature configuration.
func (fc *FeatureConfig) Validate() error {
	// No specific validations needed for feature flags currently
	return nil
}

// Validate validates monitoring configuration.
func (mc *MonitoringConfig) Validate() error {
	if mc.Enabled {
		if mc.PrometheusPort <= 0 || mc.PrometheusPort > 65535 {
			return fmt.Errorf("invalid prometheus_port: %d", mc.PrometheusPort)
		}
		
		if mc.MetricsPath == "" {
			return fmt.Errorf("metrics_path is required when monitoring is enabled")
		}
		
		if mc.SampleRate < 0 || mc.SampleRate > 1 {
			return fmt.Errorf("sample_rate must be between 0 and 1, got %f", mc.SampleRate)
		}
	}
	
	return nil
}


// Load loads configuration from files and environment variables.
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	viper.AddConfigPath("/etc/brokle")

	// Set environment variable support
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	
	// Bind standard infrastructure variables (no BROKLE_ prefix)
	viper.BindEnv("database.url", "DATABASE_URL")
	viper.BindEnv("clickhouse.url", "CLICKHOUSE_URL") 
	viper.BindEnv("redis.url", "REDIS_URL")
	viper.BindEnv("server.port", "PORT")
	viper.BindEnv("environment", "ENV") 
	viper.BindEnv("logging.level", "LOG_LEVEL")
	
	// External API keys (standard names)
	viper.BindEnv("external.openai.api_key", "OPENAI_API_KEY")
	viper.BindEnv("external.anthropic.api_key", "ANTHROPIC_API_KEY")
	viper.BindEnv("external.cohere.api_key", "COHERE_API_KEY")
	viper.BindEnv("external.stripe.secret_key", "STRIPE_SECRET_KEY")
	viper.BindEnv("external.stripe.publishable_key", "STRIPE_PUBLISHABLE_KEY")
	viper.BindEnv("external.stripe.webhook_secret", "STRIPE_WEBHOOK_SECRET")
	
	// JWT keys (standard names)
	viper.BindEnv("jwt.private_key", "JWT_PRIVATE_KEY")
	viper.BindEnv("jwt.public_key", "JWT_PUBLIC_KEY")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	
	// Keep BROKLE_ prefix for Brokle-specific variables
	viper.BindEnv("enterprise.license.key", "BROKLE_ENTERPRISE_LICENSE_KEY")
	viper.BindEnv("enterprise.license.type", "BROKLE_ENTERPRISE_LICENSE_TYPE")
	viper.BindEnv("enterprise.license.offline_mode", "BROKLE_ENTERPRISE_LICENSE_OFFLINE_MODE")
	viper.BindEnv("enterprise.sso.enabled", "BROKLE_ENTERPRISE_SSO_ENABLED")
	viper.BindEnv("enterprise.sso.provider", "BROKLE_ENTERPRISE_SSO_PROVIDER")
	viper.BindEnv("enterprise.rbac.enabled", "BROKLE_ENTERPRISE_RBAC_ENABLED")
	viper.BindEnv("enterprise.compliance.enabled", "BROKLE_ENTERPRISE_COMPLIANCE_ENABLED")
	viper.BindEnv("enterprise.analytics.enabled", "BROKLE_ENTERPRISE_ANALYTICS_ENABLED")

	// Set default values
	setDefaults()

	// Read config file (optional)
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// Config file not found - continue with defaults and env vars
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create and validate license wrapper
	licenseWrapper := NewLicenseWrapper(&cfg)
	if err := licenseWrapper.ValidateLicense(); err != nil {
		return nil, fmt.Errorf("license validation failed: %w", err)
	}
	
	return &cfg, nil
}

// GetLicenseWrapper returns a license wrapper for enhanced license management
func (c *Config) GetLicenseWrapper() *LicenseWrapper {
	return NewLicenseWrapper(c)
}

// setDefaults sets default configuration values.
func setDefaults() {
	// App defaults
	viper.SetDefault("app.name", "Brokle Platform")
	viper.SetDefault("app.version", "1.0.0")
	
	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "30s")
	viper.SetDefault("server.write_timeout", "30s")
	viper.SetDefault("server.shutdown_timeout", "30s")
	viper.SetDefault("server.max_request_size", 32<<20) // 32MB
	viper.SetDefault("server.enable_cors", true)

	// Database defaults (URL-first, individual fields as fallback)
	viper.SetDefault("database.url", "")  // Preferred: Set via DATABASE_URL env var
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", 5432)
	viper.SetDefault("database.user", "brokle")
	viper.SetDefault("database.database", "brokle")
	viper.SetDefault("database.ssl_mode", "disable")
	viper.SetDefault("database.max_open_conns", 100)
	viper.SetDefault("database.max_idle_conns", 10)
	viper.SetDefault("database.conn_max_lifetime", "1h")
	viper.SetDefault("database.conn_max_idle_time", "15m")

	// ClickHouse defaults (URL-first, individual fields as fallback)
	viper.SetDefault("clickhouse.url", "")  // Preferred: Set via CLICKHOUSE_URL env var
	viper.SetDefault("clickhouse.host", "localhost")
	viper.SetDefault("clickhouse.port", 9000)
	viper.SetDefault("clickhouse.user", "default")
	viper.SetDefault("clickhouse.database", "brokle_analytics")
	viper.SetDefault("clickhouse.max_open_conns", 50)
	viper.SetDefault("clickhouse.max_idle_conns", 5)
	viper.SetDefault("clickhouse.conn_max_lifetime", "1h")
	viper.SetDefault("clickhouse.read_timeout", "30s")
	viper.SetDefault("clickhouse.write_timeout", "30s")

	// Redis defaults (URL-first, individual fields as fallback)
	viper.SetDefault("redis.url", "")  // Preferred: Set via REDIS_URL env var
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.database", 0)
	viper.SetDefault("redis.pool_size", 20)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.idle_timeout", "5m")
	viper.SetDefault("redis.max_retries", 3)

	// JWT defaults
	viper.SetDefault("jwt.issuer", "brokle-platform")
	viper.SetDefault("jwt.access_token_ttl", "15m")
	viper.SetDefault("jwt.refresh_token_ttl", "168h")  // 7 days
	viper.SetDefault("jwt.api_key_token_ttl", "8760h") // 1 year
	viper.SetDefault("jwt.algorithm", "RS256")
	viper.SetDefault("jwt.private_key", "")  // Must be set in environment
	viper.SetDefault("jwt.public_key", "")   // Must be set in environment

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
	viper.SetDefault("logging.output", "stdout")

	// External service defaults
	viper.SetDefault("external.openai.base_url", "https://api.openai.com/v1")
	viper.SetDefault("external.anthropic.base_url", "https://api.anthropic.com")
	viper.SetDefault("external.cohere.base_url", "https://api.cohere.ai/v1")

	// Feature flags defaults
	viper.SetDefault("features.semantic_caching", true)
	viper.SetDefault("features.real_time_metrics", true)
	viper.SetDefault("features.ml_routing", true)
	viper.SetDefault("features.background_jobs", true)
	viper.SetDefault("features.rate_limiting", true)
	viper.SetDefault("features.audit_logging", true)

	// Monitoring defaults
	viper.SetDefault("monitoring.enabled", true)
	viper.SetDefault("monitoring.prometheus_port", 9090)
	viper.SetDefault("monitoring.metrics_path", "/metrics")
	viper.SetDefault("monitoring.sample_rate", 0.1)

	// Enterprise defaults
	viper.SetDefault("enterprise.license.type", "free")
	viper.SetDefault("enterprise.license.max_requests", 10000) // Free tier: 10K requests
	viper.SetDefault("enterprise.license.max_users", 5)        // Free tier: 5 users
	viper.SetDefault("enterprise.license.max_projects", 2)     // Free tier: 2 projects
	viper.SetDefault("enterprise.license.offline_mode", false)

	// SSO defaults (disabled by default)
	viper.SetDefault("enterprise.sso.enabled", false)
	viper.SetDefault("enterprise.sso.provider", "")

	// RBAC defaults (disabled by default) 
	viper.SetDefault("enterprise.rbac.enabled", false)

	// Compliance defaults (disabled by default)
	viper.SetDefault("enterprise.compliance.enabled", false)
	viper.SetDefault("enterprise.compliance.audit_retention", "168h")  // Basic: 7 days
	viper.SetDefault("enterprise.compliance.data_retention", "720h")   // Basic: 30 days
	viper.SetDefault("enterprise.compliance.pii_anonymization", false)
	viper.SetDefault("enterprise.compliance.soc2_compliance", false)
	viper.SetDefault("enterprise.compliance.hipaa_compliance", false)
	viper.SetDefault("enterprise.compliance.gdpr_compliance", false)

	// Analytics defaults (basic enabled)
	viper.SetDefault("enterprise.analytics.enabled", true)
	viper.SetDefault("enterprise.analytics.predictive_insights", false)
	viper.SetDefault("enterprise.analytics.custom_dashboards", false)
	viper.SetDefault("enterprise.analytics.ml_models", false)

	// Support defaults
	viper.SetDefault("enterprise.support.level", "standard")
	viper.SetDefault("enterprise.support.sla", "99.9%")
	viper.SetDefault("enterprise.support.dedicated_manager", false)
	viper.SetDefault("enterprise.support.on_call_support", false)
}


// GetServerAddress returns the server address string.
func (c *Config) GetServerAddress() string {
	return fmt.Sprintf("%s:%d", c.Server.Host, c.Server.Port)
}

// GetDatabaseURL returns the PostgreSQL connection URL.
func (c *Config) GetDatabaseURL() string {
	// Priority 1: Use URL if provided
	if c.Database.URL != "" {
		return c.Database.URL
	}
	
	// Priority 2: Construct from individual fields
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.Database.User, c.Database.Password, c.Database.Host,
		c.Database.Port, c.Database.Database, c.Database.SSLMode)
}

// GetClickHouseURL returns the ClickHouse connection URL.
func (c *Config) GetClickHouseURL() string {
	// Priority 1: Use URL if provided
	if c.ClickHouse.URL != "" {
		return c.ClickHouse.URL
	}
	
	// Priority 2: Construct from individual fields
	return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
		c.ClickHouse.User, c.ClickHouse.Password, c.ClickHouse.Host,
		c.ClickHouse.Port, c.ClickHouse.Database)
}

// GetRedisURL returns the Redis connection URL.
func (c *Config) GetRedisURL() string {
	// Priority 1: Use URL if provided
	if c.Redis.URL != "" {
		return c.Redis.URL
	}
	
	// Priority 2: Construct from individual fields
	if c.Redis.Password != "" {
		return fmt.Sprintf("redis://:%s@%s:%d/%d",
			c.Redis.Password, c.Redis.Host, c.Redis.Port, c.Redis.Database)
	}
	return fmt.Sprintf("redis://%s:%d/%d",
		c.Redis.Host, c.Redis.Port, c.Redis.Database)
}

// IsDevelopment returns true if running in development environment.
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development" || c.Environment == "dev"
}

// IsProduction returns true if running in production environment.
func (c *Config) IsProduction() bool {
	return c.Environment == "production" || c.Environment == "prod"
}

// IsEnterpriseFeatureEnabled checks if an enterprise feature is enabled
func (c *Config) IsEnterpriseFeatureEnabled(feature string) bool {
	if !c.IsEnterpriseLicense() {
		return false
	}
	
	for _, f := range c.Enterprise.License.Features {
		if f == feature {
			return true
		}
	}
	return false
}

// IsEnterpriseLicense returns true if the license supports enterprise features
func (c *Config) IsEnterpriseLicense() bool {
	return c.Enterprise.License.Type == "pro" ||
		   c.Enterprise.License.Type == "business" || 
		   c.Enterprise.License.Type == "enterprise"
}

// GetLicenseTier returns the current license tier
func (c *Config) GetLicenseTier() string {
	if c.Enterprise.License.Type != "" {
		return c.Enterprise.License.Type
	}
	return "free" // Default to free tier
}

// CanUseFeature checks if a specific feature can be used based on license
func (c *Config) CanUseFeature(feature string) bool {
	// Allow all features in development mode
	if c.IsDevelopment() {
		return true
	}
	
	// Check if it's an enterprise feature
	enterpriseFeatures := []string{
		"advanced_rbac", "sso_integration", "custom_compliance",
		"predictive_insights", "custom_dashboards", "on_premise_deployment",
		"dedicated_support", "advanced_integrations", "cross_org_analytics",
	}
	
	for _, ef := range enterpriseFeatures {
		if ef == feature {
			return c.IsEnterpriseFeatureEnabled(feature)
		}
	}
	
	// Non-enterprise features are always available
	return true
}