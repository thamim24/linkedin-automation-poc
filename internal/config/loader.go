package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Load reads configuration from YAML file and environment variables
func Load(configPath string) (*AppConfig, error) {
	// Load .env file if it exists
	loadEnvFile(".env")

	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	// Set defaults
	setDefaults()

	// Read config file
	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Override from environment variables
	overrideFromEnv()

	var cfg AppConfig
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Browser defaults
	viper.SetDefault("browser.headless", false)
	viper.SetDefault("browser.disable_gpu", false)
	viper.SetDefault("browser.no_sandbox", false)
	viper.SetDefault("browser.window_width", 1920)
	viper.SetDefault("browser.window_height", 1080)

	// Mouse movement defaults
	viper.SetDefault("stealth.mouse_movement.bezier_steps", 100)
	viper.SetDefault("stealth.mouse_movement.min_steps", 80)
	viper.SetDefault("stealth.mouse_movement.max_steps", 150)
	viper.SetDefault("stealth.mouse_movement.overshoot", true)
	viper.SetDefault("stealth.mouse_movement.overshoot_spread", 10)
	viper.SetDefault("stealth.mouse_movement.jitter_radius", 3)

	// Typing defaults
	viper.SetDefault("stealth.typing.min_keystroke_delay", 50*time.Millisecond)
	viper.SetDefault("stealth.typing.max_keystroke_delay", 200*time.Millisecond)
	viper.SetDefault("stealth.typing.error_rate", 0.02)
	viper.SetDefault("stealth.typing.word_pause_min", 100*time.Millisecond)
	viper.SetDefault("stealth.typing.word_pause_max", 300*time.Millisecond)

	// Timing defaults
	viper.SetDefault("stealth.timing.min_action_delay", 500*time.Millisecond)
	viper.SetDefault("stealth.timing.max_action_delay", 2*time.Second)
	viper.SetDefault("stealth.timing.page_load_timeout", 30*time.Second)
	viper.SetDefault("stealth.timing.fatigue_multiplier", 1.5)
	viper.SetDefault("stealth.timing.fatigue_after", 20)

	// Scroll defaults
	viper.SetDefault("stealth.scroll.min_scroll_amount", 200)
	viper.SetDefault("stealth.scroll.max_scroll_amount", 600)
	viper.SetDefault("stealth.scroll.scroll_steps", 10)
	viper.SetDefault("stealth.scroll.scroll_back_chance", 0.1)
	viper.SetDefault("stealth.scroll.pause_between_min", 200*time.Millisecond)
	viper.SetDefault("stealth.scroll.pause_between_max", 800*time.Millisecond)

	// Fingerprint defaults
	viper.SetDefault("stealth.fingerprint.languages", []string{"en-US", "en"})
	viper.SetDefault("stealth.fingerprint.timezones", []string{"America/New_York", "America/Chicago", "America/Los_Angeles"})
	viper.SetDefault("stealth.fingerprint.viewport_width_min", 1366)
	viper.SetDefault("stealth.fingerprint.viewport_width_max", 1920)
	viper.SetDefault("stealth.fingerprint.viewport_height_min", 768)
	viper.SetDefault("stealth.fingerprint.viewport_height_max", 1080)
	viper.SetDefault("stealth.fingerprint.cpu_cores", []int{4, 8, 16})

	// Limits defaults
	viper.SetDefault("limits.daily_connections", 50)
	viper.SetDefault("limits.daily_messages", 20)
	viper.SetDefault("limits.daily_searches", 100)
	viper.SetDefault("limits.connection_cooldown", 5*time.Minute)
	viper.SetDefault("limits.message_cooldown", 10*time.Minute)
	viper.SetDefault("limits.max_consecutive_actions", 5)

	// Schedule defaults
	viper.SetDefault("schedule.business_hours_only", true)
	viper.SetDefault("schedule.start_hour", 9)
	viper.SetDefault("schedule.end_hour", 17)
	viper.SetDefault("schedule.break_after", 10)
	viper.SetDefault("schedule.break_duration", 15*time.Minute)
	viper.SetDefault("schedule.weekdays", []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday"})

	// Database defaults
	viper.SetDefault("database.path", "./data/automation.db")

	// LinkedIn defaults
	viper.SetDefault("linkedin.base_url", "https://www.linkedin.com")
}

// overrideFromEnv overrides sensitive config from environment variables
func overrideFromEnv() {
	if email := os.Getenv("LINKEDIN_EMAIL"); email != "" {
		viper.Set("linkedin.email", email)
	}
	if password := os.Getenv("LINKEDIN_PASSWORD"); password != "" {
		viper.Set("linkedin.password", password)
	}
	if dbPath := os.Getenv("DATABASE_PATH"); dbPath != "" {
		viper.Set("database.path", dbPath)
	}
}

// validate checks if configuration values are valid
func validate(cfg *AppConfig) error {
	// Validate limits
	if cfg.Limits.DailyConnections <= 0 {
		return fmt.Errorf("daily_connections must be greater than 0")
	}
	if cfg.Limits.DailyMessages < 0 {
		return fmt.Errorf("daily_messages must be non-negative")
	}
	if cfg.Limits.DailySearches <= 0 {
		return fmt.Errorf("daily_searches must be greater than 0")
	}

	// Validate viewport ranges
	if cfg.Stealth.Fingerprint.ViewportWidthMin > cfg.Stealth.Fingerprint.ViewportWidthMax {
		return fmt.Errorf("viewport_width_min must be <= viewport_width_max")
	}
	if cfg.Stealth.Fingerprint.ViewportHeightMin > cfg.Stealth.Fingerprint.ViewportHeightMax {
		return fmt.Errorf("viewport_height_min must be <= viewport_height_max")
	}
	if cfg.Stealth.Fingerprint.ViewportWidthMin < 800 {
		return fmt.Errorf("viewport_width_min must be at least 800")
	}
	if cfg.Stealth.Fingerprint.ViewportHeightMin < 600 {
		return fmt.Errorf("viewport_height_min must be at least 600")
	}

	// Validate business hours
	if cfg.Schedule.StartHour < 0 || cfg.Schedule.StartHour > 23 {
		return fmt.Errorf("start_hour must be between 0 and 23")
	}
	if cfg.Schedule.EndHour < 0 || cfg.Schedule.EndHour > 23 {
		return fmt.Errorf("end_hour must be between 0 and 23")
	}
	if cfg.Schedule.StartHour >= cfg.Schedule.EndHour {
		return fmt.Errorf("start_hour must be less than end_hour")
	}

	// Validate mouse movement
	if cfg.Stealth.MouseMovement.MinSteps > cfg.Stealth.MouseMovement.MaxSteps {
		return fmt.Errorf("mouse min_steps must be <= max_steps")
	}

	// Validate credentials
	if cfg.LinkedIn.Email == "" {
		return fmt.Errorf("linkedin email is required")
	}
	if cfg.LinkedIn.Password == "" {
		return fmt.Errorf("linkedin password is required")
	}

	// Validate database path
	if cfg.Database.Path == "" {
		return fmt.Errorf("database path is required")
	}

	return nil
}

// loadEnvFile loads environment variables from a .env file
func loadEnvFile(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		// .env file is optional, so don't return error
		return nil
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split on first =
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		value = strings.Trim(value, `"'`)

		// Set environment variable
		os.Setenv(key, value)
	}

	return scanner.Err()
}
