package api

import (
	//"errors"
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
