//持仓处理表

package handler

import (
	"log"
	"strconv"
	"time"

	"fmt"

	"../model"
	"github.com/go-macaron/session"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

/**
我的持仓
**/
func MyHoldingHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//访问的用户

	uid, err := strconv.Atoi(ctx.Params(":uid"))
	if err != nil {
		ctx.HTML(200, "error")
		return
	} else {

		//login user
		_, loginUser := GetSessionUser(sess)
		if loginUser != nil {
			if loginUser.Id == int64(uid) {
				ctx.Data["myself"] = true
			}
			ctx.Data["my"] = loginUser
		}

		//userAccount
		user := new(model.User)
		if has, _ := x.Id(uid).Get(user); has {
			ctx.Data["user"] = user
			userAccount := new(model.UserAccount)
			if has, _ := x.Where("user_id = ?", uid).Get(userAccount); has {
				ctx.Data["ua"] = userAccount
			}
			//stock holding list
			holdinigs := make([]*model.StockHolding, 0)
			if err := x.Where("user_id= ? and holding_status=?", user.Id, 1).OrderBy("id").Limit(5, 0).Desc("id").Find(&holdinigs); err == nil {
				ctx.Data["stockHoldings"] = holdinigs
			} else {
				log.Printf("no entrusts stocks %d", user.Id)
			}
			ctx.HTML(200, "my_holding")

		} else {
			ctx.Data["msg"] = "没有此用户"
			ctx.HTML(200, "error")
			return
		}
	}
}

/**
分页持仓
**/
func MyHoldingListHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//login user
	_, user := GetSessionUser(sess)
	ctx.Data["user"] = user

	//stock holding list
	page := ctx.Params(":page")
	i, _ := strconv.Atoi(page)
	count, _ := x.Where("user_id= ? and holding_status= ?", user.Id, 1).Count(new(model.StockHolding))
	mappp := Paginator(i, count)
	ctx.Data["paginator"] = mappp
	fmt.Print("")
	index := mappp["startIndex"].(int)
	if count > 0 {
		entrusts := make([]*model.StockHolding, 0)
		x.Where("user_id= ? and holding_status=?", user.Id, 1).Desc("id").
			Limit(model.PAGE_SIZE, index).Find(&entrusts)
		ctx.Data["stockHoldings"] = entrusts
	}
	ctx.HTML(200, "holding_list")
}

/**
当日委托
**/
func TodayEntrustHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//login user
	_, user := GetSessionUser(sess)
	ctx.Data["user"] = user

	//userAccount
	userAccount := new(model.UserAccount)
	if has, _ := x.Where("user_id = ?", user.Id).Get(userAccount); has {
		ctx.Data["ua"] = userAccount
	}
	//stock entrust list
	today := time.Now().Format(model.DATE_FORMAT_1)
	entrusts := make([]*model.StockEntrust, 0)
	if err := x.Where("user_id= ? and entrust_time>=?", user.Id, today).Limit(5, 0).Desc("id").Find(&entrusts); err == nil {
		ctx.Data["entrusts"] = entrusts
	}
	//今日委托
	ctx.HTML(200, "my_entrust")
}

/**
所有委托
**/
func MyEntrustListHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//login user
	_, user := GetSessionUser(sess)
	ctx.Data["user"] = user

	//userAccount
	userAccount := new(model.UserAccount)
	if has, _ := x.Where("user_id = ?", user.Id).Get(userAccount); has {
		ctx.Data["ua"] = userAccount
	}
	//stock entrust list
	entrusts := make([]*model.StockEntrust, 0)
	page := ctx.Params(":page")
	i, _ := strconv.Atoi(page)
	count, _ := x.Where("user_id= ?", user.Id).
		Count(new(model.StockEntrust))
	mappp := Paginator(i, count)
	ctx.Data["paginator"] = mappp
	fmt.Print("")
	index := mappp["startIndex"].(int)
	if count > 0 {
		x.Where("user_id= ?", user.Id).Limit(model.PAGE_SIZE, index).Desc("id").Find(&entrusts)
	}
	ctx.Data["entrusts"] = entrusts
	//今日委托
	ctx.HTML(200, "entrust_list")
}

/**
今日成交
**/
func TodayStockDealHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//login user
	_, user := GetSessionUser(sess)
	ctx.Data["user"] = user

	//userAccount
	userAccount := new(model.UserAccount)
	if has, _ := x.Where("user_id = ?", user.Id).Get(userAccount); has {
		ctx.Data["ua"] = userAccount
	}
	//stock deal today
	today := time.Now().Format(model.DATE_FORMAT_1)
	trxs := make([]*model.StockTrans, 0)
	if err := x.Where("user_id= ? and trans_status=? and trans_time>=?", user.Id, 1, today).Desc("id").
		Limit(5, 0).Find(&trxs); err != nil {
		ctx.Data["trxs"] = trxs
	} else {
		log.Printf("no trxs stocks %d", user.Id)
	}
	//今日委托
	ctx.HTML(200, "today_trx")
}

/**
历史成交
**/
func MyStockDealListHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//login user
	_, user := GetSessionUser(sess)
	ctx.Data["user"] = user

	//userAccount
	userAccount := new(model.UserAccount)
	if has, _ := x.Where("user_id = ?", user.Id).Get(userAccount); has {
		ctx.Data["ua"] = userAccount
	}
	//stock deal today

	//stock tranx list
	trxs := make([]*model.StockTrans, 0)
	page := ctx.Params(":page")
	i, _ := strconv.Atoi(page)
	count, _ := x.Where("user_id= ? and trans_status=?", user.Id, 1).
		Count(new(model.StockTrans))

	mappp := Paginator(i, count)
	ctx.Data["paginator"] = mappp
	index := mappp["startIndex"].(int)

	if count > 0 {
		x.Where("user_id= ? and trans_status=?", user.Id, 1).Desc("id").
			Limit(model.PAGE_SIZE, index).Find(&trxs)
		ctx.Data["trxs"] = trxs
	}
	ctx.HTML(200, "trx_list")
}
