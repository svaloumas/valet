package redis

import (
	"os"
	"testing"
)

func TestGetRedisPrefixedKey(t *testing.T) {
	redisURL := os.Getenv("REDIS_URL")
	rcWithPrefixedKey := New(redisURL, 1, 5, "some-prefixed-key")
	rcWithoutPrefixedKey := New(redisURL, 1, 5, "")

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
			rcWithPrefixedKey,
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
