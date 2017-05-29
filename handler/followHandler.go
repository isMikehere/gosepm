package handler

import (
	"fmt"
	"log"
	"time"

	"../model"

	"strconv"

	"github.com/go-macaron/session"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

/**
订阅的第一步的处理器
**/
func FollowStep1Handler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {
	//获取要订阅的人的ID
	userId := ctx.Params(":userId")
	if ok, u := QueryUserByIdWithEngine(x, userId); ok {
		//获取有效的上线产品
		products := ListOnlineProducts(x, log)
		//获取当前用户信息
		i := sess.Get("user").(model.User)
		ctx.Data["user2Follow"] = u
		ctx.Data["products"] = products
		ctx.Data["i"] = i
		ctx.HTML(200, "follow_step1")

	} else {
		log.Printf("没有找到对应订阅用户%s", userId)
	}
}

/**
 订阅前检查
 @param i :订阅人
 @param user2Follow:待订阅人
**/
func preCheckToFollow(i *model.User, user2Follow *model.User) (bool, string) {
	return true, "ok"
}

/**
订阅第二步奏
进入收银台
**/
func FollowStep2Handler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {

	userId := ctx.Params(":userId")
	if ok, u := QueryUserByIdWithEngine(x, userId); ok {
		//获取有效的上线产品
		Type := ctx.Params(":type")
		typeInt, _ := strconv.Atoi(Type)

		if has, product := QueryProductByType(typeInt, x); has {
			//获取当前用户信息
			i := sess.Get("user").(model.User)
			//订阅前检查
			if canFollow, errMsg := preCheckToFollow(&i, u); canFollow {
				ctx.Data["product"] = product
			} else {
				ctx.Data["errMsg"] = errMsg
			}
		}
	}
}

/**
订阅成功后对提醒
**/
func NotifyAllAfterPay(x *xorm.Engine, log *log.Logger, orderId int64) {

	order := new(model.StockOrder)

	if has, _ := x.Id(orderId).Get(order); has {

		user := new(model.User)         //订阅人
		followedUser := new(model.User) //被订阅人
		_, e1 := x.Id(order.UserId).Get(user)
		Chk(e1)
		_, e2 := x.Id(order.UserId).Get(followedUser)
		Chk(e2)
		uf := new(model.UserFollow)
		uf.UserId = order.UserId
		uf.FollowedId = order.FollowedUserId
		uf.FollowType = order.ProductType
		uf.FollowStart = time.Now()
		uf.FollowEnd = time.Now().AddDate(0, 0, 7) //订阅结束

		_, err := x.Insert(uf)
		Chk(err)

		var weeks = 1
		if order.ProductType == 1 {
			weeks = 4
		}
		// "尊敬的客户%s：您好，您已经成功订阅高手%s的为期%d的股票提醒，有效期为%s-%s,如有问题，请联系我们电话：%s"
		content := fmt.Sprintf(model.FOLLOW_OK_MSG, user.UserName, followedUser.UserName, weeks, uf.FollowStart.Format(model.DATE_TIME_FORMAT), uf.FollowEnd.Format(model.DATE_TIME_FORMAT), model.HOT_LINE)
		SendMessage(x, log, user.Mobile, content)
		// 尊敬的客户%s：您好，%s已经成功订阅您的为期%d周股票提醒，有效期为%s-%s,如有问题，请联系我们电话：%s
		SendMessage(x, log, followedUser.Mobile, fmt.Sprintf(model.TOBEFOLLOWED_OK_MSG, followedUser.UserName, user.UserName, weeks, uf.FollowStart.Format(model.DATE_TIME_FORMAT), uf.FollowEnd.Format(model.DATE_TIME_FORMAT), model.HOT_LINE))

	}

}
