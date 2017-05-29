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
		Addr:         "localhost:6379",
		DialTimeout:  10 * time.Second,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		PoolSize:     10,
		PoolTimeout:  30 * time.Second,
		Password:     "xceof",
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
func SetRedisStock(client *redis.Client, s model.Stock) {
	bs, _ := json.Marshal(s)
	client.HSet(model.R_KEY_STOCKS, s.StockCode, string(bs))
}

/**
根据用户ID获取用户
**/
func GetRedisUser(client *redis.Client, uid int64) *model.User {
	j, err := client.HGet(model.R_KEY_USERS, strconv.Itoa(int(uid))).Result()
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
func SetRedisUser(client *redis.Client, u model.User) {
	bs, _ := json.Marshal(u)
	client.HSet(model.R_KEY_USERS, strconv.Itoa(int(u.Id)), string(bs))
}

func SetTest(client *redis.Client) {
	client.Set("name", "mike", 0)
}
