package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Browser    BrowserConfig    `yaml:"browser"`
	Stealth    StealthConfig    `yaml:"stealth"`
	RateLimits RateLimitsConfig `yaml:"rate_limits"`
	Search     SearchConfig     `yaml:"search"`
	Connection ConnectionConfig `yaml:"connection"`
	Messaging  MessagingConfig  `yaml:"messaging"`
	Scheduling SchedulingConfig `yaml:"scheduling"`
	Storage    StorageConfig    `yaml:"storage"`
	Logging    LoggingConfig    `yaml:"logging"`

	// From environment
	LinkedIn LinkedInCredentials
}

type BrowserConfig struct {
	Headless   bool           `yaml:"headless"`
	Viewport   ViewportConfig `yaml:"viewport"`
	UserAgents []string       `yaml:"user_agents"`
}

type ViewportConfig struct {
	Width  int `yaml:"width"`
	Height int `yaml:"height"`
}

type StealthConfig struct {
	EnableMouseMovement   bool            `yaml:"enable_mouse_movement"`
	EnableRandomScrolling bool            `yaml:"enable_random_scrolling"`
	EnableHumanTyping     bool            `yaml:"enable_human_typing"`
	EnableMouseHovering   bool            `yaml:"enable_mouse_hovering"`
	EnableIdleBreaks      bool            `yaml:"enable_idle_breaks"`
	ActionDelay           DelayConfig     `yaml:"action_delay"`
	ScrollDelay           DelayConfig     `yaml:"scroll_delay"`
	TypingDelay           DelayConfig     `yaml:"typing_delay"`
	ThinkTime             DelayConfig     `yaml:"think_time"`
	IdleBreak             IdleBreakConfig `yaml:"idle_break"`
}

type DelayConfig struct {
	Min int `yaml:"min"`
	Max int `yaml:"max"`
}

type IdleBreakConfig struct {
	MinDurationSeconds int `yaml:"min_duration_seconds"`
	MaxDurationSeconds int `yaml:"max_duration_seconds"`
	FrequencyActions   int `yaml:"frequency_actions"`
}

type RateLimitsConfig struct {
	Connections RateLimit `yaml:"connections"`
	Messages    RateLimit `yaml:"messages"`
	Searches    RateLimit `yaml:"searches"`
}

type RateLimit struct {
	PerHour int `yaml:"per_hour"`
	PerDay  int `yaml:"per_day"`
}

type SearchConfig struct {
	Targets             []SearchTarget `yaml:"targets"`
	MaxResultsPerSearch int            `yaml:"max_results_per_search"`
	PaginationLimit     int            `yaml:"pagination_limit"`
}

type SearchTarget struct {
	JobTitle string `yaml:"job_title"`
	Location string `yaml:"location"`
	Keywords string `yaml:"keywords"`
}

type ConnectionConfig struct {
	SendNote      bool     `yaml:"send_note"`
	NoteTemplates []string `yaml:"note_templates"`
	NoteMaxLength int      `yaml:"note_max_length"`
}

type MessagingConfig struct {
	Enabled                   bool     `yaml:"enabled"`
	DelayAfterConnectionHours int      `yaml:"delay_after_connection_hours"`
	Templates                 []string `yaml:"templates"`
	FollowUpEnabled           bool     `yaml:"follow_up_enabled"`
}

type SchedulingConfig struct {
	ActiveHours ActiveHoursConfig `yaml:"active_hours"`
	ActiveDays  []string          `yaml:"active_days"`
	Timezone    string            `yaml:"timezone"`
}

type ActiveHoursConfig struct {
	Start int `yaml:"start"`
	End   int `yaml:"end"`
}

type StorageConfig struct {
	DatabasePath string `yaml:"database_path"`
	CookiePath   string `yaml:"cookie_path"`
}

type LoggingConfig struct {
	Level   string `yaml:"level"`
	File    string `yaml:"file"`
	Console bool   `yaml:"console"`
}

type LinkedInCredentials struct {
	Email    string
	Password string
}

// Load reads configuration from config.yaml and environment variables
func Load() (*Config, error) {
	// Load .env file if exists
	_ = godotenv.Load()

	// Read YAML config
	data, err := os.ReadFile("config.yaml")
	if err != nil {
		return nil, fmt.Errorf("failed to read config.yaml: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config.yaml: %w", err)
	}

	// Override with environment variables
	cfg.LinkedIn.Email = getEnv("LINKEDIN_EMAIL", "")
	cfg.LinkedIn.Password = getEnv("LINKEDIN_PASSWORD", "")

	if cfg.LinkedIn.Email == "" || cfg.LinkedIn.Password == "" {
		return nil, fmt.Errorf("LINKEDIN_EMAIL and LINKEDIN_PASSWORD must be set")
	}

	// Override other settings from env if present
	if headless := os.Getenv("HEADLESS"); headless != "" {
		cfg.Browser.Headless = headless == "true"
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		cfg.Storage.DatabasePath = dbPath
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Logging.Level = logLevel
	}

	// Override rate limits from env
	if val := os.Getenv("MAX_CONNECTIONS_PER_DAY"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.RateLimits.Connections.PerDay = n
		}
	}

	if val := os.Getenv("MAX_MESSAGES_PER_DAY"); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			cfg.RateLimits.Messages.PerDay = n
		}
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Browser.Viewport.Width <= 0 || c.Browser.Viewport.Height <= 0 {
		return fmt.Errorf("invalid viewport dimensions")
	}

	if len(c.Browser.UserAgents) == 0 {
		return fmt.Errorf("at least one user agent must be specified")
	}

	if c.RateLimits.Connections.PerDay <= 0 {
		return fmt.Errorf("connections per day must be positive")
	}

	if c.Storage.DatabasePath == "" {
		return fmt.Errorf("database path must be specified")
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
