package handler

import (
	"log"

	"../model"

	"github.com/go-macaron/session"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
)

/**
发送短信等外置接口
mobiles :18201401937,18201401937
**/
func sendMessage(mobiles, content string) (bool, string) {
	log.Printf("---------短信发送：%s,%s", mobiles, content)
	p := make(map[string]string, 5)
	p["zh"] = model.MSG_ACCOUNT
	p["mm"] = model.MSG_PASS
	p["hm"] = mobiles
	p["nr"] = model.MSG_TITLE + content
	p["sms_type"] = model.MSG_BIZ_CHAN
	return HttpPost(model.MSG_SEND_API, p)
}

/**
最新通知列表
通过短地址获取msgKey ,如果redis，存在，则直接返回数据，
如果不存在，则判断用户是否登录，如果登录，则跳转消息页面，否则直接返回链接失效
**/
func LatestMsgHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	msgKey := ctx.Params(":msgKey")
	msg, _ := redis.Get(msgKey).Result()
	if msg == "" { //没有命中，则判断是否登录
		if login, _ := GetSessionUser(sess); login {
			//跳转消息列表
			ctx.HTML(200, "msg_list")
		} else {
			ctx.Redirect("/login.htm")
		}
	} else {
		ctx.Data["msg"] = msg
		ctx.HTML(200, "latest_msg")
	}
}
