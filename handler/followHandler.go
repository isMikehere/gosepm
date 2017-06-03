package handler

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"../model"

	"github.com/go-macaron/session"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

/**
订阅的第一步的处理器
**/
func FollowStep1Handler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {
	//获取要订阅的人的ID
	userId := ctx.Params(":id")
	fmt.Print("-----userId," + userId)
	u := new(model.User)
	if has, _ := x.Id(userId).Get(u); has {
		//获取有效的上线产品
		products := ListOnlineProducts(x, log)
		//获取当前用户信息
		// i := sess.Get("user").(model.User)
		ctx.Data["user2Follow"] = u
		ctx.Data["products"] = products
		// ctx.Data["i"] = i
		ctx.HTML(200, "follow_step1")
	} else {
		ctx.Data["msg"] = "没有找到对应订阅用户"
		ctx.HTML(200, "error")
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
订阅第二步
提交订单
进入收银台
**/
func FollowStep2Handler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redisCli *redis.Client) {

	userId, _ := strconv.Atoi(ctx.Params(":id"))
	followedUser := new(model.User)
	if has, _ := x.Id(userId).Get(followedUser); has {
		//获取有效的上线产品
		Type := ctx.Params(":type")
		fmt.Printf("%s", Type)
		typeInt, _ := strconv.Atoi(Type)
		product := new(model.Product)
		if has, _ := x.Where("types=?", typeInt).And("is_online=?", 1).Get(product); has {
			fmt.Printf("OK")
			//获取当前用户信息
			i := sess.Get("user").(*model.User)
			//订阅前检查
			if canFollow, errMsg := preCheckToFollow(i, followedUser); canFollow {
				ctx.Data["product"] = product
				if ok, order := GenerateOrder(x, redisCli, typeInt, i.Id, followedUser.Id); ok {
					ctx.Data["ok"] = ok
					ctx.Data["order"] = order
					ctx.HTML(200, "follow_step2")
				} else {
					ctx.Data["msg"] = "下单失败，请重试"
					ctx.HTML(200, "error")
				}

			} else {
				ctx.Data["msg"] = errMsg
				ctx.HTML(200, "error")
			}
		} else {
			log.Printf("没有产品%d", typeInt)
			ctx.Data["msg"] = "没有找到对应的订阅产品"
			ctx.HTML(200, "error")
		}
	}
}

/**·
订阅成功后对提醒
**/
func NotifyAllAfterPay(x *xorm.Engine, orderId int64) {

	//订单
	s := x.NewSession()
	s.Begin()
	defer s.Close()

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
		uf.FollowedId = order.FollowedId
		uf.FollowType = order.ProductType
		uf.FollowStart = time.Now()
		uf.FollowEnd = time.Now().AddDate(0, 0, 7) //订阅结束

		_, err := x.Insert(uf)
		if err != nil {
			s.Rollback()
			return
		}
		//更新用户的订阅量
		userAccount := new(model.UserAccount)
		if has, err = x.Where("user_id=?", followedUser.Id).Get(&userAccount); err != nil {
			s.Rollback()
			return
		} else {
			if has {
				userAccount.TotalFollow++
				_, err = x.Where("user_id=?", followedUser.Id).Update(userAccount)
				if err != nil {
					s.Rollback()
					return
				}
			}
		}

		var weeks = 1
		if order.ProductType == 1 {
			weeks = 4
		}
		// 【金修网络】恭喜您：%s已经成功订阅您的为期%d周股票模拟交易提醒，有效期为%s-%s。请及时处理详情请参考订单须知。
		messageLog := new(model.MessageLog)
		messageLog.Mobile = followedUser.Mobile
		messageLog.Content =
			fmt.Sprintf(model.TOBEFOLLOWED_OK_MSG, user.NickName, weeks, uf.FollowStart.Format(model.DATE_TIME_FORMAT),
				uf.FollowEnd.Format(model.DATE_TIME_FORMAT))
		messageLog.SendStatus = 0

		if _, err := s.Insert(messageLog); err == nil {
			s.Commit()
		} else {
			log.Printf("出现异常%s", err.Error())
			s.Rollback()
		}

	}

}

/**
用户订阅列表
**/
func UserFollowListHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redisCli *redis.Client) {

	followStatus, _ := strconv.Atoi(ctx.Params(":followStatus"))
	_, user := GetSessionUser(sess)

	ufs := make([]*model.UserFollow, 0)
	s := x.Where("user_id=?", user.Id)
	if followStatus != -1 {
		s.And("follow_status=?", followStatus).Find(&ufs)
	} else {
		s.Find(&ufs)
	}
	//分页处理
	ctx.Data["follows"] = ufs
	ctx.JSON(200, ufs)
	// ctx.HTML(200, "my_follow")
}
