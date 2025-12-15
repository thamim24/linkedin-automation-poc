package config

import "time"

// AppConfig represents the complete application configuration
type AppConfig struct {
	Browser   BrowserConfig   `mapstructure:"browser"`
	Stealth   StealthConfig   `mapstructure:"stealth"`
	Limits    LimitsConfig    `mapstructure:"limits"`
	Schedule  ScheduleConfig  `mapstructure:"schedule"`
	Database  DatabaseConfig  `mapstructure:"database"`
	LinkedIn  LinkedInConfig  `mapstructure:"linkedin"`
}

// BrowserConfig contains browser launch settings
type BrowserConfig struct {
	Headless        bool     `mapstructure:"headless"`
	ChromiumPath    string   `mapstructure:"chromium_path"`
	UserDataDir     string   `mapstructure:"user_data_dir"`
	DisableGPU      bool     `mapstructure:"disable_gpu"`
	NoSandbox       bool     `mapstructure:"no_sandbox"`
	WindowWidth     int      `mapstructure:"window_width"`
	WindowHeight    int      `mapstructure:"window_height"`
	ExtraFlags      []string `mapstructure:"extra_flags"`
}

// StealthConfig contains stealth behavior settings
type StealthConfig struct {
	MouseMovement   MouseConfig     `mapstructure:"mouse_movement"`
	Typing          TypingConfig    `mapstructure:"typing"`
	Timing          TimingConfig    `mapstructure:"timing"`
	Scroll          ScrollConfig    `mapstructure:"scroll"`
	Fingerprint     FingerprintConfig `mapstructure:"fingerprint"`
}

// MouseConfig defines mouse movement behavior
type MouseConfig struct {
	BezierSteps       int     `mapstructure:"bezier_steps"`
	MinSteps          int     `mapstructure:"min_steps"`
	MaxSteps          int     `mapstructure:"max_steps"`
	Overshoot         bool    `mapstructure:"overshoot"`
	OvershootSpread   int     `mapstructure:"overshoot_spread"`
	JitterRadius      int     `mapstructure:"jitter_radius"`
}

// TypingConfig defines typing behavior
type TypingConfig struct {
	MinKeystrokeDelay time.Duration `mapstructure:"min_keystroke_delay"`
	MaxKeystrokeDelay time.Duration `mapstructure:"max_keystroke_delay"`
	ErrorRate         float64       `mapstructure:"error_rate"`
	WordPauseMin      time.Duration `mapstructure:"word_pause_min"`
	WordPauseMax      time.Duration `mapstructure:"word_pause_max"`
}

// TimingConfig defines general timing behavior
type TimingConfig struct {
	MinActionDelay    time.Duration `mapstructure:"min_action_delay"`
	MaxActionDelay    time.Duration `mapstructure:"max_action_delay"`
	PageLoadTimeout   time.Duration `mapstructure:"page_load_timeout"`
	FatigueMultiplier float64       `mapstructure:"fatigue_multiplier"`
	FatigueAfter      int           `mapstructure:"fatigue_after"`
}

// ScrollConfig defines scrolling behavior
type ScrollConfig struct {
	MinScrollAmount   int           `mapstructure:"min_scroll_amount"`
	MaxScrollAmount   int           `mapstructure:"max_scroll_amount"`
	ScrollSteps       int           `mapstructure:"scroll_steps"`
	ScrollBackChance  float64       `mapstructure:"scroll_back_chance"`
	PauseBetweenMin   time.Duration `mapstructure:"pause_between_min"`
	PauseBetweenMax   time.Duration `mapstructure:"pause_between_max"`
}

// FingerprintConfig defines browser fingerprint settings
type FingerprintConfig struct {
	UserAgents        []string `mapstructure:"user_agents"`
	Languages         []string `mapstructure:"languages"`
	Timezones         []string `mapstructure:"timezones"`
	ViewportWidthMin  int      `mapstructure:"viewport_width_min"`
	ViewportWidthMax  int      `mapstructure:"viewport_width_max"`
	ViewportHeightMin int      `mapstructure:"viewport_height_min"`
	ViewportHeightMax int      `mapstructure:"viewport_height_max"`
	CPUCores          []int    `mapstructure:"cpu_cores"`
}

// LimitsConfig defines rate limiting and daily caps
type LimitsConfig struct {
	DailyConnections    int           `mapstructure:"daily_connections"`
	DailyMessages       int           `mapstructure:"daily_messages"`
	DailySearches       int           `mapstructure:"daily_searches"`
	ConnectionCooldown  time.Duration `mapstructure:"connection_cooldown"`
	MessageCooldown     time.Duration `mapstructure:"message_cooldown"`
	MaxConsecutiveActions int         `mapstructure:"max_consecutive_actions"`
}

// ScheduleConfig defines when the bot should run
type ScheduleConfig struct {
	BusinessHoursOnly bool          `mapstructure:"business_hours_only"`
	StartHour         int           `mapstructure:"start_hour"`
	EndHour           int           `mapstructure:"end_hour"`
	BreakAfter        int           `mapstructure:"break_after"`
	BreakDuration     time.Duration `mapstructure:"break_duration"`
	Weekdays          []string      `mapstructure:"weekdays"`
}

// DatabaseConfig defines database settings
type DatabaseConfig struct {
	Path string `mapstructure:"path"`
}

// LinkedInConfig contains LinkedIn credentials and settings
type LinkedInConfig struct {
	Email    string `mapstructure:"email"`
	Password string `mapstructure:"password"`
	BaseURL  string `mapstructure:"base_url"`
}