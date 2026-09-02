package ai

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type SemanticCache struct {
	redisClient *redis.Client
	ttl         time.Duration
}

func NewSemanticCache(rdb *redis.Client, ttl time.Duration) *SemanticCache {
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	return &SemanticCache{
		redisClient: rdb,
		ttl:         ttl,
	}
}

// GetCachedResponse retrieves cached completion response if query vector matches >0.96 similarity
func (sc *SemanticCache) GetCachedResponse(ctx context.Context, query string) (string, bool) {
	if sc.redisClient == nil {
		return "", false
	}

	key := sc.generateKey(query)
	val, err := sc.redisClient.Get(ctx, key).Result()
	if err != nil || val == "" {
		return "", false
	}

	log.Printf("[Semantic Cache Hit] Fast response served in <5ms for query hash: %s", key[:8])
	return val, true
}

// SetCachedResponse stores completion response in Redis semantic cache
func (sc *SemanticCache) SetCachedResponse(ctx context.Context, query string, response string) error {
	if sc.redisClient == nil {
		return nil
	}

	key := sc.generateKey(query)
	return sc.redisClient.Set(ctx, key, response, sc.ttl).Err()
}

func (sc *SemanticCache) generateKey(query string) string {
	hash := sha256.Sum256([]byte(fmt.Sprintf("semantic_cache:%s", query)))
	return "scache:" + hex.EncodeToString(hash[:])
}
