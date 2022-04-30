package redis

import (
	"fmt"
	"os"
	"testing"

	vlog "github.com/svaloumas/valet/pkg/log"
)

var redisTest *RedisClient

func TestMain(m *testing.M) {

	redisURL := os.Getenv("REDIS_URL")
	redisTest = New(redisURL, 1, 5, "some-prefixed-key", nil)
	defer redisTest.Close()

	m.Run()
}

func TestCheckHealth(t *testing.T) {
	result := redisTest.CheckHealth()
	if result != true {
		t.Fatalf("expected true got %#v instead", result)
	}
}

func TestCheckHealthNotHealthy(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	notHealthyRedis := New(redisURL, 1, 5, "", nil)
	notHealthyRedis.Close()

	result := redisTest.CheckHealth()
	if result != true {
		t.Fatalf("expected false got %#v instead", result)
	}
}

func TestPanicWithInvalidURL(t *testing.T) {
	logger := vlog.NewLogger("redis", "text")
	redisURL := "invalid_url"
	defer func() {
		if p := recover(); p == nil {
			t.Errorf("New did not panic with invalid URL")
		} else {
			panicMsg := "redis: invalid URL scheme: "
			if err := fmt.Errorf("%s", p); err.Error() != panicMsg {
				t.Errorf("New paniced with unexpected panic message: got %v want %v", err.Error(), panicMsg)
			}
		}
	}()
	New(redisURL, 1, 5, "", logger)
}

func TestGetRedisPrefixedKey(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	rcWithoutPrefixedKey := New(redisURL, 1, 5, "", nil)
	defer rcWithoutPrefixedKey.Close()

	tests := []struct {
		name     string
		key      string
		expected string
		rc       *RedisClient
	}{
		{
			"with prefixed key",
			"test",
			"some-prefixed-key:test",
			redisTest,
		},
		{
			"no prefixed key",
			"test",
			"test",
			rcWithoutPrefixedKey,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redisKey := tt.rc.GetRedisPrefixedKey(tt.key)

			if redisKey != tt.expected {
				t.Errorf("get redis prefixed key returned wrong redis key: got %v want %v", redisKey, tt.expected)
			}
		})
	}
}
