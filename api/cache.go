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
	Client *redis.Client
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

func NewRedisCache(ttl time.Duration, config *Config) (*RedisCache, error) {
	// redisHost := os.Getenv("REDISHOST")
	// redisPort := os.Getenv("REDISPORT")
	cache := new(RedisCache)
	cache.Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.RedisHost, config.RedisPort),
		Password:     config.RedisPass,
		ReadTimeout:  time.Minute,
		MinIdleConns: 1,
	})

	pong, err := cache.Client.Ping().Result()
	fmt.Println("result:", pong, err)
	cache.ttl = ttl
	return cache, err
}
func NewRedisCacheHost(ttl time.Duration, host, pass string, port int) *RedisCache {
	// redisHost := os.Getenv("REDISHOST")
	// redisPort := os.Getenv("REDISPORT")
	cache := new(RedisCache)
	cache.Client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", host, port),
		Password:     pass,
		ReadTimeout:  time.Minute,
		MinIdleConns: 1,
	})

	pong, err := cache.Client.Ping().Result()
	fmt.Println("err:", pong, err)
	cache.ttl = ttl
	return cache
}

func (cache *RedisCache) SetUserSubs(uid, subs string) (bool, error) {
	return cache.Client.HSet(uid, UserSubs, subs).Result()
}
func (cache *RedisCache) GetUserSubs(uid string) (string, error) {
	return cache.Client.HGet(uid, UserSubs).Result()
}

// Public key cache
// {publicKey}:uid-{uid}
func (cache *RedisCache) SetPublicKey(uid, publicKey string) (bool, error) {
	cache.Client.HSet(publicKey, UIDC, uid)
	return cache.Client.HSet(uid, PublicKey, publicKey).Result()
}
func (cache *RedisCache) GetUidFromPublicKey(publicKey string) (string, error) {
	return cache.Client.HGet(publicKey, UIDC).Result()
}
func (cache *RedisCache) GetPublicKeyFromUid(Uid string) (string, error) {
	return cache.Client.HGet(Uid, PublicKey).Result()
}

// Notice cache
func (cache *RedisCache) SetNotice(uid, cacheType string, value bool) (bool, error) {
	return cache.Client.HSet(uid+UserInfo, cacheType, value).Result()
}
func (cache *RedisCache) SetNotices(uid string, pairs ...interface{}) (string, error) {
	return cache.Client.MSet(uid+UserInfo, pairs).Result()
}
func (cache *RedisCache) GetNotice(uid, cacheType string) (string, error) {
	return cache.Client.HGet(uid+UserInfo, cacheType).Result()
}

func (cache *RedisCache) SetXLMPrice(price float64) (bool, error) {
	return cache.Client.HSet("prices", "xlmP", price).Result()
}
func (cache *RedisCache) GetXLMPrice() (string, error) {
	return cache.Client.HGet("prices", "xlmP").Result()
}
func (cache *RedisCache) GetXLMUsd() (float64, error) {
	return cache.Client.HGet("prices", "xlmP").Float64()
}
func (cache *RedisCache) SetGRXPrice(price float64) (bool, error) {
	return cache.Client.HSet("prices", "grxP", price).Result()
}
func (cache *RedisCache) GetGRXPrice() (string, error) {
	return cache.Client.HGet("prices", "grxP").Result()
}
func (cache *RedisCache) SetGRXUsd(price float64) (bool, error) {
	return cache.Client.HSet("prices", "grxUsd", price).Result()
}
func (cache *RedisCache) GetGRXUsd() (float64, error) {
	return cache.Client.HGet("prices", "grxUsd").Float64()
}
func (cache *RedisCache) SetFunc2LAS(las float64) (string, error) {
	return cache.Client.Set("LAS", las, 0).Result()
}
func (cache *RedisCache) GetFunc2LAS() (float64, error) {
	las, err := cache.Client.Get("LAS").Float64()
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
		res = cache.Client.IncrBy(uid+key, 1)
	} else if action == "CLOSE" && n > 0 {
		res = cache.Client.DecrBy(uid+key, n)
	} else if n == 0 {
		err := cache.Client.Set(uid+key, 0, 0).Err()
		return int64(0), err
	}

	return res.Result()
}
func (cache *RedisCache) UpdatePositionOpen(uid, algorithm, grayllTxId string, currentValue float64) (int64, error) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	//str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	_, hashCurrentValue := BuildHash(uid, algorithm)
	cache.Client.HSet(hashCurrentValue, cacheTxId, currentValue)
	return cache.UpdatePositionNumbers(uid, algorithm, "OPEN", 1)
}
func (cache *RedisCache) UpdateCurrentPositionValues(uid, algorithm, grayllTxId string, roi, currentValue float64) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	hashRoi, hashCurrentValue := BuildHash(uid, algorithm)

	cache.Client.HSet(hashRoi, cacheTxId, roi)
	cache.Client.HSet(hashCurrentValue, cacheTxId, currentValue)
}

func (cache *RedisCache) UpdatePositionClose(uid, algorithm, grayllTxId string) (int64, error) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	hashRoi := fmt.Sprintf("%s_%s_current_ROI", uid, str)
	hashCurrentValue := fmt.Sprintf("%s_%s_current_value", uid, str)
	cache.Client.HDel(hashRoi, cacheTxId)
	cache.Client.HDel(hashCurrentValue, cacheTxId)

	return cache.UpdatePositionNumbers(uid, algorithm, "CLOSE", 1)
}

func (cache *RedisCache) UpdatePositionCloseAll(uid, algorithm, grayllTxId string) {
	cacheTxId, _ := GetCacheTxId(algorithm, grayllTxId)
	str := strings.ToLower(strings.ReplaceAll(algorithm, " ", ""))
	hashRoi := fmt.Sprintf("%s_%s_current_ROI", uid, str)
	hashCurrentValue := fmt.Sprintf("%s_%s_current_value", uid, str)
	cache.Client.HDel(hashRoi, cacheTxId)
	cache.Client.HDel(hashCurrentValue, cacheTxId)

	//return cache.UpdatePositionNumbers(uid, algorithm, "CLOSE", 1)
}

func (cache *RedisCache) GetCurrentValues(uid, algorithm string, isGetRoi bool) (float64, float64) {

	hashRoi, hashCurrentValue := BuildHash(uid, algorithm)

	currentValues := cache.Client.HGetAll(hashCurrentValue)

	totalCurrentRoi := float64(0)
	totalCurrentValue := float64(0)

	if isGetRoi {
		rois := cache.Client.HGetAll(hashRoi)
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
	cache.Client.HSet("referer", uid, refererUid)
}
func (cache *RedisCache) GetRefererUid(uid string) string {
	return cache.Client.HGet("referer", uid).String()
}
func (cache *RedisCache) DelRefererUid(uid string) {
	cache.Client.HDel("referer", uid)
}

func (cache *RedisCache) UpdateRoi(graylltx, algo string, value float64, roiType string) {
	roiKey := ""
	switch roiType {
	case "24h":
		roiKey = algo + "_roi24h"
		cache.Client.ZAdd(roiKey, redis.Z{value, graylltx})
	case "7d":
		roiKey = algo + "_roi7d"
		cache.Client.ZAdd(roiKey, redis.Z{value, graylltx})
	case "total":
		roiKey = algo + "_roitotal"
		cache.Client.ZAdd(roiKey, redis.Z{value, graylltx})
	case "all":
		roiKey = algo + "_roi24h"
		cache.Client.ZAdd(roiKey, redis.Z{value, graylltx})
		roiKey = algo + "_roi7d"
		cache.Client.ZAdd(roiKey, redis.Z{value, graylltx})
		roiKey = algo + "_roitotal"
		cache.Client.ZAdd(roiKey, redis.Z{value, graylltx})
	}

}
func (cache *RedisCache) RemoveRoi(graylltx, algo string, value float64, duration int64) {
	roiKey := algo + "_roi24h"
	cache.Client.ZRem(roiKey, graylltx)
	roiKey = algo + "_roi7d"
	cache.Client.ZRem(roiKey, graylltx)
	roiKey = algo + "_roitotal"
	cache.Client.ZRem(roiKey, graylltx)
}

func (cache *RedisCache) GetRois(algo string) []float64 {
	res := make([]float64, 0)
	val := float64(0)
	var err error
	roiType := algo + "_roi24h"
	//log.Println("roiType:", roiType)
	roi24h, err := cache.Client.ZRevRangeWithScores(roiType, 0, 0).Result()
	if err != nil || len(roi24h) == 0 {
		//log.Println("error ZRangeByScore roi24h", err, roi24h)
		val = 0
	} else {
		//log.Println("ZRangeByScore roi24h", roi24h)
		val = roi24h[0].Score
	}
	res = append(res, val)

	roiType = algo + "_roi7d"
	roi7d, err := cache.Client.ZRevRangeWithScores(roiType, 0, 0).Result()
	if err != nil || len(roi7d) == 0 {
		//log.Println("error ZRangeByScore roi7d", err)
		val = 0
	} else {
		val = roi7d[0].Score
	}
	res = append(res, val)

	roiType = algo + "_roitotal"
	roiTotal, err := cache.Client.ZRevRangeWithScores(roiType, 0, 0).Result()
	if err != nil || len(roiTotal) == 0 {
		log.Println("error ZRangeByScore roiTotal", err)
		val = 0
	} else {
		val = roiTotal[0].Score
	}
	res = append(res, val)

	return res

}

// emaillogin emailregister
func (cache *RedisCache) SetRecapchaToken(emailAction, recapchaToken string) {
	cache.Client.Set(emailAction+"-recapcha", recapchaToken, time.Minute*2)
}

func (cache *RedisCache) CheckRecapchaToken(emailAction string) bool {
	_, err := cache.Client.Get(emailAction + "-recapcha").Result()
	if err != nil || err == redis.Nil {
		return false
	}
	return true
}
