package handler

import (
	"fmt"

	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

/**
高手搜素
**/

/**
首页请求
**/
func IndexHandler(ctx *macaron.Context, engine *xorm.Engine, redisCli *redis.Client) {

	//日排名
	d_syncAt, dailyRanks := GetDayRanks(engine)
	fmt.Printf("%s", dailyRanks)
	ctx.Data["dailyRanks"] = dailyRanks
	ctx.Data["d_syncAt"] = d_syncAt
	// //周排名
	w_syncAt, weekRanks := GetWeekRanks(engine)
	ctx.Data["weekRanks"] = weekRanks
	ctx.Data["w_syncAt"] = w_syncAt
	//月排名
	m_syncAt, monthRanks := GetMonthRanks(engine)
	ctx.Data["monthRanks"] = monthRanks
	ctx.Data["m_syncAt"] = m_syncAt

	// //新闻动态
	news := IndexNews(engine)
	ctx.Data["news"] = news
	ctx.HTML(200, "index")
}

/**

**/
func GetPhoneCodeHandler(ctx *macaron.Context, redisCli *redis.Client) {
	ctx.Params("phone")
	// redisCli.Set(,)
}
