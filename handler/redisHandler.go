package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"strings"

	"../model"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
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

/**
 redis subscribe
**/
func SubscribeMsgChan(x *xorm.Engine, client *redis.Client) {

	//发布chan
	pubsub := client.Subscribe(model.R_MSG_SEND_CHAN)
	defer pubsub.Close()

	if msgi, err := pubsub.ReceiveTimeout(time.Second); err == nil {
		subscr := msgi.(*redis.Subscription)
		if subscr.Count == 1 {
			log.Printf("subscribe %s ok", model.R_MSG_SEND_CHAN)
		} else {
			log.Printf("subscribe %s fail", model.R_MSG_SEND_CHAN)
		}
	} else {
		log.Printf("err:%s", err.Error())
	}

	//监听
	for {
		if msg, _ := pubsub.ReceiveMessage(); msg != nil {
			messageLog := new(model.MessageLog)
			if err := json.Unmarshal([]byte(msg.Payload), messageLog); err == nil {

				if messageLog.SendStatus == 0 {
					if f, ret := sendMessage(messageLog.Mobile, messageLog.Content); f {
						messageLog.SendStatus = 1
						messageLog.RetBatchId = strings.Split(ret, "0:")[1]
					} else {
						messageLog.SendStatus = 2
					}
					fmt.Printf("save message to db: %s", messageLog.Mobile)
					//消息结果发送到结果队列
					go x.Insert(messageLog)

				}
			} else {
				log.Printf("解析失败%s", msg.Payload)
			}
		}
	}
}

/**
发布消息
**/
func PublishMessage(client *redis.Client, chanName string, messageLog *model.MessageLog) bool {

	if messageLog.SendStatus != 0 {
		log.Printf("消息状态有误%s", messageLog.Mobile)
		return false
	}

	log.Printf("发布消息：to chan:%s ,msg;%s", chanName, messageLog.Mobile)
	bs, _ := json.Marshal(messageLog)
	if _, e := client.Publish(chanName, string(bs)).Result(); e != nil {
		log.Printf("发布消息：to chan:%s ,msg;%s", chanName, "false")
		return false
	}
	log.Printf("发布消息：to chan:%s ,msg;%s", chanName, "true")
	return true
}
