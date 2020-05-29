package api

import (
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

const (
	UserSubs  = "subs"
	UserInfo  = "-info"
	PublicKey = "pk"
	UIDC      = "uid"

	// notice
	// AppGeneral = "app-general"
	// AppWallet  = "app-wallet"
	// AppAlgo    = "app-algo"
	//
	// MailGeneral = "mail-general"
	// MailWallet  = "mail-wallet"
	// MailAlgo    = "mail-algo"
)

func NewRedisCache(ttl time.Duration, config *Config) *RedisCache {
	// redisHost := os.Getenv("REDISHOST")
	// redisPort := os.Getenv("REDISPORT")
	cache := new(RedisCache)
	cache.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password: config.RedisPass,
	})

	pong, err := cache.client.Ping().Result()
	fmt.Println("err:", pong, err)
	cache.ttl = ttl
	return cache
}
func NewRedisCacheHost(ttl time.Duration, host, pass string, port int) *RedisCache {
	// redisHost := os.Getenv("REDISHOST")
	// redisPort := os.Getenv("REDISPORT")
	cache := new(RedisCache)
	cache.client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", host, port),
		Password: pass,
	})

	pong, err := cache.client.Ping().Result()
	fmt.Println("err:", pong, err)
	cache.ttl = ttl
	return cache
}

func (cache *RedisCache) SetUserSubs(uid, subs string) (bool, error) {
	return cache.client.HSet(uid, UserSubs, subs).Result()
}
func (cache *RedisCache) GetUserSubs(uid string) (string, error) {
	return cache.client.HGet(uid, UserSubs).Result()
}

// Public key cache
// {publicKey}:uid-{uid}
func (cache *RedisCache) SetPublicKey(uid, publicKey string) (bool, error) {
	cache.client.HSet(publicKey, UIDC, uid)
	return cache.client.HSet(uid, PublicKey, publicKey).Result()
}
func (cache *RedisCache) GetUidFromPublicKey(publicKey string) (string, error) {
	return cache.client.HGet(publicKey, UIDC).Result()
}
func (cache *RedisCache) GetPublicKeyFromUid(Uid string) (string, error) {
	return cache.client.HGet(Uid, PublicKey).Result()
}

// Notice cache
func (cache *RedisCache) SetNotice(uid, cacheType string, value bool) (bool, error) {
	return cache.client.HSet(uid+UserInfo, cacheType, value).Result()
}
func (cache *RedisCache) SetNotices(uid string, pairs ...interface{}) (string, error) {
	return cache.client.MSet(uid+UserInfo, pairs).Result()
}
func (cache *RedisCache) GetNotice(uid, cacheType string) (string, error) {
	return cache.client.HGet(uid+UserInfo, cacheType).Result()
}

func (cache *RedisCache) SetXLMPrice(price float64) (bool, error) {
	return cache.client.HSet("prices", "xlmP", price).Result()
}
func (cache *RedisCache) GetXLMPrice() (string, error) {
	return cache.client.HGet("prices", "xlmP").Result()
}
func (cache *RedisCache) GetXLMUsd() (float64, error) {
	return cache.client.HGet("prices", "xlmP").Float64()
}
func (cache *RedisCache) SetGRXPrice(price float64) (bool, error) {
	return cache.client.HSet("prices", "grxP", price).Result()
}
func (cache *RedisCache) GetGRXPrice() (string, error) {
	return cache.client.HGet("prices", "grxP").Result()
}
func (cache *RedisCache) SetGRXUsd(price float64) (bool, error) {
	return cache.client.HSet("prices", "grxUsd", price).Result()
}
func (cache *RedisCache) GetGRXUsd() (float64, error) {
	return cache.client.HGet("prices", "grxUsd").Float64()
}
func (cache *RedisCache) SetFunc2LAS(las float64) (string, error) {
	return cache.client.Set("LAS", las, 0).Result()
}
func (cache *RedisCache) GetFunc2LAS() (float64, error) {
	las, err := cache.client.Get("LAS").Float64()
	if err != nil {
		return 0, err
	}
	return las, nil
}
