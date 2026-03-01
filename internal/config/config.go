// Package config implements routines for initializing the distributed rate limiting
// application settings.
package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
)

// A Config encapsulates the configuration options for the application.
type Config struct {
	Hostname            string // Hostname for API server
	RedisHostname       string // Hostname for Redis server
	RedisPassword       string // Password for Redis server
	Port                int    // Port for API server
	RateLimit           int    // Maximum number of requests within the sliding window
	RedisPort           int    // Port for Redis server
	ReadHeaderTimeoutMS int    // HTTP request read header timeout in millisecond
	ReadTimeoutMS       int    // HTTP request timeout in milliseconds
	TimeoutMS           int    // HTTP request timeout in milliseconds
	WindowSizeSec       int    // Duration of the sliding window in seconds
}

// New returns the configuration for the distributed rate limiting application.  The
// configuration is set using the following environment variables:
//   - DRL_HOSTNAME: Hostname for the API server
//   - DRL_PORT: Port for the API server
//   - DRL_RATE_LIMIT: Maximum number of requests within the sliding window
//   - DRL_REDIS_HOSTNAME: Hostname for Redis server
//   - DRL_REDIS_PASSWORD: Password for Redis server
//   - DRL_REDIS_PORT: Port for Redis server
//   - DRL_WINDOW_SIZE_SEC: Duration of the sliding window in seconds
//   - DRL_READ_HEADER_TIMEOUT_MS: HTTP read request header timeout in milliseconds
//   - DRL_READ_TIMEOUT_MS: HTTP read request timeout in milliseconds
//   - DRL_TIMEOUT_MS: HTTP response timeout in milliseconds
func New() (*Config, error) {
	var err error
	var errs []error
	c := &Config{}

	c.Hostname = getEnv("DRL_HOSTNAME", "")
	c.Port, err = getInt("DRL_PORT", 8080)
	if err != nil {
		errs = append(errs, err)
	}
	c.RateLimit, err = getInt("DRL_RATE_LIMIT", 10)
	if err != nil {
		errs = append(errs, err)
	}
	c.RedisHostname = getEnv("DRL_REDIS_HOSTNAME", "drl-redis")
	c.RedisPassword = getEnv("DRL_REDIS_PASSWORD", "")
	c.RedisPort, err = getInt("DRL_REDIS_PORT", 6379)
	if err != nil {
		errs = append(errs, err)
	}
	c.WindowSizeSec, err = getInt("DRL_WINDOW_SIZE_SEC", 10)
	if err != nil {
		errs = append(errs, err)
	}
	c.ReadHeaderTimeoutMS, err = getInt("DRL_READ_HEADER_TIMEOUT_MS", 500)
	if err != nil {
		errs = append(errs, err)
	}
	c.ReadTimeoutMS, err = getInt("DRL_READ_TIMEOUT_MS", 500)
	if err != nil {
		errs = append(errs, err)
	}
	c.TimeoutMS, err = getInt("DRL_TIMEOUT_MS", 1000)
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	err = validate(c)
	if err != nil {
		return nil, err
	}

	return c, nil
}

func validate(c *Config) error {
	var errs []error
	if c.Port <= 0 || c.Port > 65535 {
		errs = append(errs, errors.New("DRL_PORT must be 1 to 65535"))
	}
	if c.RedisHostname == "" {
		errs = append(errs, errors.New("DRL_REDIS_HOSTNAME must be set"))
	}
	if c.RedisPort <= 0 || c.RedisPort > 65535 {
		errs = append(errs, errors.New("DRL_REDIS_PORT must be 1 to 65535"))
	}
	if c.RateLimit <= 0 {
		errs = append(errs, errors.New("DRL_RATE_LIMIT must be greater than zero"))
	}
	if c.WindowSizeSec <= 0 {
		errs = append(errs, errors.New("DRL_WINDOW_SIZE_SEC must be greater than zero"))
	}
	if c.ReadHeaderTimeoutMS <= 0 {
		errs = append(errs, errors.New("DRL_READ_HEADER_TIMEOUT_MS must be greater than zero"))
	}
	if c.ReadTimeoutMS <= 0 {
		errs = append(errs, errors.New("DRL_READ_TIMEOUT_MS must be greater than zero"))
	}
	if c.TimeoutMS <= 0 {
		errs = append(errs, errors.New("DRL_TIMEOUT_MS must be greater than zero"))
	}
	if c.ReadHeaderTimeoutMS > c.ReadTimeoutMS {
		errs = append(errs, errors.New("DRL_READ_TIMEOUT_MS must be equal to or greater than DRL_READ_HEADER_TIMEOUT_MS"))
	}

	return errors.Join(errs...)
}

func getEnv(key string, def string) string {
	v, ok := os.LookupEnv(key)
	if !ok {
		v = def
	}

	return v
}

func getInt(key string, def int) (int, error) {
	str, ok := os.LookupEnv(key)
	if !ok {
		return def, nil
	}

	v, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("invalid int %q for %s setting: %w", str, key, err)
	}

	return v, nil
}
