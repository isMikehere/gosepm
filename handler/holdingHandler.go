//持仓处理表

package handler

import (
	"log"

	"../model"
	"github.com/go-macaron/session"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

/**
我的持仓
**/
func MyHoldingHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//login user
	_, user := GetSessionUser(sess)
	ctx.Data["user"] = user
	//userAccount
	userAccount := new(model.UserAccount)
	if has, _ := x.Where("user_id = ?", user.Id).Get(userAccount); has {
		ctx.Data["ua"] = userAccount
	}
	//stock holding list
	entrusts := make([]*model.StockHolding, 0)
	if err := x.Where("user_id= ? and holding_status=?", user.Id, 1).Limit(5, 0).Find(&entrusts); err != nil {
		ctx.Data["entrusts"] = entrusts
	} else {
		log.Printf("no entrusts stocks %d", user.Id)
	}
	ctx.HTML(200, "my_holding")
}

/**
分页数据
**/
func MyPageableHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client) {

	//login user
	_, user := GetSessionUser(sess)
	ctx.Data["user"] = user

	//stock holding list
	entrusts := make([]*model.StockHolding, 0)
	if err := x.Where("user_id= ? and holding_status=?", user.Id, 1).Limit(5, 0).Find(&entrusts); err != nil {
		ctx.Data["entrusts"] = entrusts
	} else {
		log.Printf("no entrusts stocks %d", user.Id)
	}

	// page := ctx.Params(":page")

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
	//stock holding list
	entrusts := make([]*model.StockEntrust, 0)
	if err := x.Sql("select * from stock_entrust "+
		" where user_id= ? and entrust_status=? "+
		"and date_format(entrust_time,'%Y-%m-%d')= date_format(curdate() ,'%Y-%m-%d')", user.Id, 1).
		And("").Limit(5, 0).Find(&entrusts); err != nil {

		ctx.Data["entrusts"] = entrusts

	} else {
		log.Printf("no entrusts stocks %d", user.Id)
	}
	//今日委托
	ctx.HTML(200, "today_entrust")
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
	entrusts := make([]*model.StockTrans, 0)
	if err := x.Sql("select * from stock_trans "+
		" where user_id= ? and trans_status=? "+
		"and date_format(entrust_time,'%Y-%m-%d')= date_format(curdate() ,'%Y-%m-%d')", user.Id, 1).
		And("").Limit(5, 0).Find(&entrusts); err != nil {

		ctx.Data["entrusts"] = entrusts

	} else {
		log.Printf("no entrusts stocks %d", user.Id)
	}
	//今日委托
	ctx.HTML(200, "today_entrust")
}
