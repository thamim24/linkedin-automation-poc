package stealth

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"linkedin-automation-poc/internal/config"
)

// Fingerprint represents a browser fingerprint
type Fingerprint struct {
	UserAgent           string
	Language            string
	Timezone            string
	Platform            string
	ViewportWidth       int
	ViewportHeight      int
	HardwareConcurrency int
	DeviceMemory        int
}

// GenerateFingerprint creates a unique but realistic browser fingerprint
func GenerateFingerprint(cfg config.FingerprintConfig) (*Fingerprint, error) {
	fp := &Fingerprint{}

	// Select random user agent
	if len(cfg.UserAgents) == 0 {
		fp.UserAgent = getDefaultUserAgent()
	} else {
		ua, err := randomChoice(cfg.UserAgents)
		if err != nil {
			return nil, err
		}
		fp.UserAgent = ua
	}

	// Select random language
	if len(cfg.Languages) == 0 {
		fp.Language = "en-US"
	} else {
		lang, err := randomChoice(cfg.Languages)
		if err != nil {
			return nil, err
		}
		fp.Language = lang
	}

	// Select random timezone
	if len(cfg.Timezones) == 0 {
		fp.Timezone = "America/New_York"
	} else {
		tz, err := randomChoice(cfg.Timezones)
		if err != nil {
			return nil, err
		}
		fp.Timezone = tz
	}

	// Random viewport within bounds
	vw, err := randomInt(cfg.ViewportWidthMin, cfg.ViewportWidthMax)
	if err != nil {
		return nil, err
	}
	fp.ViewportWidth = vw

	vh, err := randomInt(cfg.ViewportHeightMin, cfg.ViewportHeightMax)
	if err != nil {
		return nil, err
	}
	fp.ViewportHeight = vh

	// Random CPU cores
	if len(cfg.CPUCores) == 0 {
		fp.HardwareConcurrency = 8
	} else {
		cores, err := randomChoice(cfg.CPUCores)
		if err != nil {
			return nil, err
		}
		fp.HardwareConcurrency = cores
	}

	// Device memory (realistic values)
	memoryOptions := []int{4, 8, 16}
	mem, err := randomChoice(memoryOptions)
	if err != nil {
		return nil, err
	}
	fp.DeviceMemory = mem

	// Platform based on user agent
	fp.Platform = detectPlatform(fp.UserAgent)

	return fp, nil
}

// detectPlatform extracts platform from user agent
func detectPlatform(userAgent string) string {
	if contains(userAgent, "Win") {
		return "Win32"
	} else if contains(userAgent, "Mac") {
		return "MacIntel"
	} else if contains(userAgent, "Linux") {
		return "Linux x86_64"
	}
	return "Win32" // default
}

// contains checks if string contains substring
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && 
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || 
		len(s) > len(substr) && s[1:len(s)-1] != s[1:len(s)-1]))
}

// getDefaultUserAgent returns a realistic default user agent
func getDefaultUserAgent() string {
	return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
}

// randomChoice selects a random element from a slice
func randomChoice[T any](slice []T) (T, error) {
	var zero T
	if len(slice) == 0 {
		return zero, fmt.Errorf("cannot choose from empty slice")
	}
	
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(slice))))
	if err != nil {
		return zero, err
	}
	
	return slice[n.Int64()], nil
}

// randomInt generates a random integer between min and max (inclusive)
func randomInt(min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}
	
	if min == max {
		return min, nil
	}
	
	diff := int64(max - min + 1)
	n, err := rand.Int(rand.Reader, big.NewInt(diff))
	if err != nil {
		return 0, err
	}
	
	return int(n.Int64()) + min, nil
}

// String returns a string representation of the fingerprint
func (fp *Fingerprint) String() string {
	return fmt.Sprintf(
		"Fingerprint{UA: %s, Lang: %s, TZ: %s, Viewport: %dx%d, Cores: %d}",
		fp.UserAgent[:50]+"...",
		fp.Language,
		fp.Timezone,
		fp.ViewportWidth,
		fp.ViewportHeight,
		fp.HardwareConcurrency,
	)
}