package handler

import (
	"encoding/json"
	"strconv"
	"time"

	"../model"
	redis "github.com/go-redis/redis"
)

func NewRedisClient() *redis.Client {

	redisCli := redis.NewClient(&redis.Options{
		Addr:         model.RedisHost,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		Password:     model.RedisPass,
	})
	return redisCli
}

/**
根据用户ID获取用户
**/
func GetRedisStock(client *redis.Client, stockCode string) *model.Stock {

	j, err := client.HGet(model.R_KEY_STOCKS, stockCode).Result()
	if err != nil || j == "" {
		return nil
	} else {
		s := new(model.Stock)
		json.Unmarshal([]byte(j), s)
		return s
	}
}

/**
设置股票
***/
func SetRedisStock(client *redis.Client, s *model.Stock) {
	bs, _ := json.Marshal(s)
	client.HSet(model.R_KEY_STOCKS, s.StockCode, string(bs))

}

/**
根据用户ID获取用户
**/
func GetRedisUser(client *redis.Client, uid string) *model.User {
	j, err := client.HGet(model.R_KEY_USERS, uid).Result()
	if err != nil || j == "" {
		return nil
	} else {
		user := new(model.User)
		json.Unmarshal([]byte(j), user)
		return user
	}
}

/**
设置用户
***/
func SetRedisUser(client *redis.Client, u *model.User) {
	bs, _ := json.Marshal(u)
	client.HSet(model.R_KEY_USERS, strconv.Itoa(int(u.Id)), string(bs))
}

/**
stockcodes
**/
func SetRedisStockCodes(client *redis.Client, codes []string) {
	bs, _ := json.Marshal(codes)
	client.Set(model.R_KEY_STOCK_CODES, string(bs), 0)
}

/**
stockcodes
**/
func GetRedisStockCodes(client *redis.Client) []string {

	j, err := client.Get(model.R_KEY_STOCK_CODES).Result()
	if err != nil || j == "" {
		return nil
	} else {
		codes := make([]string, 0)
		json.Unmarshal([]byte(j), &codes)
		return codes
	}
}

/**
stockcode detail
**/
func GetRedisStockDetail(client *redis.Client, stockCode string) string {

	j, _ := client.HGet(model.R_KEY_STOCKS_DETAIL, stockCode).Result()
	return j
}

/**
最新消息列表
**/
func GetRedisLatestMsg(client *redis.Client, key string) string {
	j, _ := client.Get(key).Result()
	return j
}

/**
设置消息
***/
func SetRedisLatestMsg(client *redis.Client, key, msg string) {
	client.Set(key, msg, time.Hour*24*2)
}
