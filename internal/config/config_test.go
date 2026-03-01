package config

import (
	"os"
	"testing"
)

var envKeys = []string{
	"DRL_HOSTNAME",
	"DRL_PORT",
	"DRL_RATE_LIMIT",
	"DRL_READ_HEADER_TIMEOUT_MS",
	"DRL_READ_TIMEOUT_MS",
	"DRL_REDIS_HOSTNAME",
	"DRL_REDIS_PASSWORD",
	"DRL_REDIS_PORT",
	"DRL_TIMEOUT_MS",
	"DRL_WINDOW_SIZE_SEC",
}

type testCase struct {
	name                string
	expectedError       bool
	hostname            *string
	port                *string
	rateLimit           *string
	readHeaderTimeoutMS *string
	readTimeoutMS       *string
	redisHostname       *string
	redisPassword       *string
	redisPort           *string
	timeoutMS           *string
	windowSizeSec       *string
	expected            *Config
}

func TestConfiguration(t *testing.T) {
	for _, key := range envKeys {
		// Use Setenv to restore the key in the test environment to its original value.
		t.Setenv(key, "")
		// Unset the key in the test environment.  Will be set back to original value
		// during cleanup.
		_ = os.Unsetenv(key)
	}

	tests := []testCase{
		{
			name:                "All Settings Valid",
			hostname:            strPtr("test-host"),
			port:                strPtr("8080"),
			rateLimit:           strPtr("10"),
			readHeaderTimeoutMS: strPtr("250"),
			readTimeoutMS:       strPtr("300"),
			redisHostname:       strPtr("test-redis-host"),
			redisPassword:       strPtr("testpassword"),
			redisPort:           strPtr("6767"),
			timeoutMS:           strPtr("500"),
			windowSizeSec:       strPtr("3"),
			expected: &Config{
				Hostname:            "test-host",
				RedisHostname:       "test-redis-host",
				RedisPassword:       "testpassword",
				Port:                8080,
				RateLimit:           10,
				ReadHeaderTimeoutMS: 250,
				ReadTimeoutMS:       300,
				RedisPort:           6767,
				TimeoutMS:           500,
				WindowSizeSec:       3,
			},
		},
		{
			name: "Default Config",
			expected: &Config{
				Hostname:            "",
				RedisHostname:       "drl-redis",
				RedisPassword:       "",
				Port:                8080,
				RateLimit:           10,
				ReadHeaderTimeoutMS: 500,
				ReadTimeoutMS:       500,
				RedisPort:           6379,
				TimeoutMS:           1000,
				WindowSizeSec:       10,
			},
		},
		{
			name:          "Invalid Port",
			expectedError: true,
			port:          strPtr("not-a-number"),
		},
		{
			name:          "Negative Port",
			expectedError: true,
			port:          strPtr("-1"),
		},
		{
			name:          "Zero Port",
			expectedError: true,
			port:          strPtr("0"),
		},
		{
			name:          "Overbound Port",
			expectedError: true,
			port:          strPtr("65536"),
		},
		{
			name:          "Invalid Handler",
			expectedError: true,
			rateLimit:     strPtr("not-a-number"),
		},
		{
			name:          "Zero Handler",
			expectedError: true,
			rateLimit:     strPtr("0"),
		},
		{
			name:          "Negative Handler",
			expectedError: true,
			rateLimit:     strPtr("-1"),
		},
		{
			name:          "Empty RedisHostname",
			expectedError: true,
			redisHostname: strPtr(""),
		},
		{
			name:          "Invalid RedisPort",
			expectedError: true,
			redisPort:     strPtr("not-a-number"),
		},
		{
			name:          "Zero RedisPort",
			expectedError: true,
			redisPort:     strPtr("0"),
		},
		{
			name:          "Negative RedisPort",
			expectedError: true,
			redisPort:     strPtr("-1"),
		},
		{
			name:          "Overbound RedisPort",
			expectedError: true,
			redisPort:     strPtr("65536"),
		},
		{
			name:          "Invalid WindowSizeSec",
			expectedError: true,
			windowSizeSec: strPtr("not-a-number"),
		},
		{
			name:          "Zero WindowSizeSec",
			expectedError: true,
			windowSizeSec: strPtr("0"),
		},
		{
			name:          "Negative WindowSizeSec",
			expectedError: true,
			windowSizeSec: strPtr("-1"),
		},
		{
			name:                "Invalid ReadHeaderTimeoutMS",
			expectedError:       true,
			readHeaderTimeoutMS: strPtr("not-a-number"),
		},
		{
			name:          "Invalid ReadTimeoutMS",
			expectedError: true,
			readTimeoutMS: strPtr("not-a-number"),
		},
		{
			name:          "Invalid TimeoutMS",
			expectedError: true,
			timeoutMS:     strPtr("not-a-number"),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			setEnv(&test, t)

			result, err := New()

			if err != nil {
				if !test.expectedError {
					t.Fatalf("expected no error, got %v", err)
				}
				// An error was expected and returned.
				return
			}
			if test.expectedError {
				t.Fatal("expected an error, got none")
			}

			if test.expected.Hostname != result.Hostname {
				t.Errorf("expected Hostname %q got %q", test.expected.Hostname, result.Hostname)
			}
			if test.expected.RedisHostname != result.RedisHostname {
				t.Errorf("expected RedisHostname %q got %q", test.expected.RedisHostname, result.RedisHostname)
			}
			if test.expected.RedisPassword != result.RedisPassword {
				t.Errorf("expected RedisPassword %q got %q", test.expected.RedisPassword, result.RedisPassword)
			}
			if test.expected.Port != result.Port {
				t.Errorf("expected Port %d got %d", test.expected.Port, result.Port)
			}
			if test.expected.RateLimit != result.RateLimit {
				t.Errorf("expected Handler %d got %d", test.expected.RateLimit, result.RateLimit)
			}
			if test.expected.RedisPort != result.RedisPort {
				t.Errorf("expected RedisPort %d got %d", test.expected.RedisPort, result.RedisPort)
			}
			if test.expected.WindowSizeSec != result.WindowSizeSec {
				t.Errorf("expected WindowSizeSec %d got %d", test.expected.WindowSizeSec, result.WindowSizeSec)
			}
			if test.expected.ReadHeaderTimeoutMS != result.ReadHeaderTimeoutMS {
				t.Errorf("expected ReadHeaderTimeoutMS %d got %d", test.expected.ReadHeaderTimeoutMS, result.ReadHeaderTimeoutMS)
			}
			if test.expected.ReadTimeoutMS != result.ReadTimeoutMS {
				t.Errorf("expected ReadTimeoutMS %d got %d", test.expected.ReadTimeoutMS, result.ReadTimeoutMS)
			}
			if test.expected.TimeoutMS != result.TimeoutMS {
				t.Errorf("expected TimeoutMS %d got %d", test.expected.TimeoutMS, result.TimeoutMS)
			}
		})
	}
}

func setEnv(td *testCase, t *testing.T) {
	if td.hostname != nil {
		t.Setenv("DRL_HOSTNAME", *td.hostname)
	}
	if td.port != nil {
		t.Setenv("DRL_PORT", *td.port)
	}
	if td.rateLimit != nil {
		t.Setenv("DRL_RATE_LIMIT", *td.rateLimit)
	}
	if td.readHeaderTimeoutMS != nil {
		t.Setenv("DRL_READ_HEADER_TIMEOUT_MS", *td.readHeaderTimeoutMS)
	}
	if td.readTimeoutMS != nil {
		t.Setenv("DRL_READ_TIMEOUT_MS", *td.readTimeoutMS)
	}
	if td.redisHostname != nil {
		t.Setenv("DRL_REDIS_HOSTNAME", *td.redisHostname)
	}
	if td.redisPassword != nil {
		t.Setenv("DRL_REDIS_PASSWORD", *td.redisPassword)
	}
	if td.redisPort != nil {
		t.Setenv("DRL_REDIS_PORT", *td.redisPort)
	}
	if td.timeoutMS != nil {
		t.Setenv("DRL_TIMEOUT_MS", *td.timeoutMS)
	}
	if td.windowSizeSec != nil {
		t.Setenv("DRL_WINDOW_SIZE_SEC", *td.windowSizeSec)
	}
}

func strPtr(s string) *string {
	return &s
}
