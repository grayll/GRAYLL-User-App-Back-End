package api

import (
	"fmt"
	"log"
	"strconv"
	"strings"
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

// Update number open positions depends on the action OPEN/CLOSE
func (cache *RedisCache) UpdatePositionNumbers(uid, algorithm, action string, n int64) (int64, error) {
	str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	key := fmt.Sprintf("_total_%s_open_positions", str)
	var res *redis.IntCmd
	if action == "OPEN" {
		res = cache.client.IncrBy(uid+key, 1)
	} else if action == "CLOSE" && n > 0 {
		res = cache.client.DecrBy(uid+key, n)
	} else if n == 0 {
		err := cache.client.Set(uid+key, 0, 0).Err()
		return int64(0), err
	}

	return res.Result()
}
func (cache *RedisCache) UpdatePositionOpen(uid, algorithm, grayllTxId string, currentValue float64) (int64, error) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	//str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	_, hashCurrentValue := BuildHash(uid, algorithm)
	cache.client.HSet(hashCurrentValue, cacheTxId, currentValue)
	return cache.UpdatePositionNumbers(uid, algorithm, "OPEN", 1)
}
func (cache *RedisCache) UpdateCurrentPositionValues(uid, algorithm, grayllTxId string, roi, currentValue float64) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	hashRoi, hashCurrentValue := BuildHash(uid, algorithm)

	cache.client.HSet(hashRoi, cacheTxId, roi)
	cache.client.HSet(hashCurrentValue, cacheTxId, currentValue)
}

func (cache *RedisCache) UpdatePositionClose(uid, algorithm, grayllTxId string) (int64, error) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	hashRoi := fmt.Sprintf("%s_%s_current_ROI", uid, str)
	hashCurrentValue := fmt.Sprintf("%s_%s_current_value", uid, str)
	cache.client.HDel(hashRoi, cacheTxId)
	cache.client.HDel(hashCurrentValue, cacheTxId)

	return cache.UpdatePositionNumbers(uid, algorithm, "CLOSE", 1)
}

func (cache *RedisCache) UpdatePositionCloseAll(uid, algorithm, grayllTxId string) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	hashRoi := fmt.Sprintf("%s_%s_current_ROI", uid, str)
	hashCurrentValue := fmt.Sprintf("%s_%s_current_value", uid, str)
	cache.client.HDel(hashRoi, cacheTxId)
	cache.client.HDel(hashCurrentValue, cacheTxId)

	//return cache.UpdatePositionNumbers(uid, algorithm, "CLOSE", 1)
}

func (cache *RedisCache) GetCurrentValues(uid, algorithm string, isGetRoi bool) (float64, float64) {

	hashRoi, hashCurrentValue := BuildHash(uid, algorithm)

	currentValues := cache.client.HGetAll(hashCurrentValue)

	totalCurrentRoi := float64(0)
	totalCurrentValue := float64(0)

	if isGetRoi {
		rois := cache.client.HGetAll(hashRoi)
		_rois, err := rois.Result()
		if err != nil {
			log.Println("[ERROR] Can not get rois for key:", hashRoi)
		} else {
			for _, v := range _rois {

				roi, _ := strconv.ParseFloat(v, 64)
				totalCurrentRoi += roi
			}
		}
	}
	_currentValues, err := currentValues.Result()
	if err != nil {
		log.Println("[ERROR] Can not get current value for key:", hashCurrentValue)
	} else {
		for _, v := range _currentValues {
			currentValue, _ := strconv.ParseFloat(v, 64)
			totalCurrentValue += currentValue
		}
	}

	return totalCurrentRoi, totalCurrentValue
}

func BuildHash(uid, algorithm string) (string, string) {
	str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	hashCurrentValue := fmt.Sprintf("%s_%s_current_value", uid, str)
	hashRoi := fmt.Sprintf("%s_%s_current_ROI", uid, str)
	return hashRoi, hashCurrentValue
}
func GetCacheTxId(algoType, grayll_tx_id string) (string, string) {
	cacheTxId := ""
	unreadPath := ""
	switch algoType {
	case "GRZ":
		unreadPath = "UrGRZ"
		cacheTxId = "z1" + grayll_tx_id
	case "GRY 1":
		unreadPath = "UrGRY1"
		cacheTxId = "y1" + grayll_tx_id
	case "GRY 2":
		unreadPath = "UrGRY2"
		cacheTxId = "y2" + grayll_tx_id
	case "GRY 3":
		unreadPath = "UrGRY3"
		cacheTxId = "y3" + grayll_tx_id
	}
	return cacheTxId, unreadPath
}

func (cache *RedisCache) SetRefererUid(uid, refererUid string) {
	cache.client.HSet("referer", uid, refererUid)
}
func (cache *RedisCache) GetRefererUid(uid string) string {
	return cache.client.HGet("referer", uid).String()
}
func (cache *RedisCache) DelRefererUid(uid string) {
	cache.client.HDel("referer", uid)
}
