//排名计算
package handler

import (
	"time"

	macaron "gopkg.in/macaron.v1"
	redis "gopkg.in/redis.v2"

	"strconv"

	"../model"
	"github.com/go-xorm/xorm"
)

/*
获取日排名
**/
func GetDayRanks(x *xorm.Engine) (string, []*model.DayRank) {

	//日排名
	dailyRanks := make([]*model.DayRank, 0)
	syncAt := time.Now().Format(model.DATE_TIME_FORMAT)
	if x.Where("1==1").Find(&dailyRanks); len(dailyRanks) > 0 {
		syncAt = dailyRanks[0].Created.Format(model.DATE_TIME_FORMAT)
	}
	return syncAt, dailyRanks
}

/**
获取周排名
**/
func GetWeekRanks(x *xorm.Engine) (string, []*model.WeekRank) {

	// //周排名
	newWeek := new(model.WeekRank)
	syncAt := time.Now().Format(model.DATE_TIME_FORMAT)
	if has, _ := x.Where("1==1").Desc("week").Limit(1, 0).Get(newWeek); has {
		weekRanks := make([]*model.WeekRank, 0)
		x.Where("week = ?", newWeek.Week).Find(&weekRanks)
		syncAt = newWeek.Created.Format(model.DATE_TIME_FORMAT)
		return syncAt, weekRanks
	} else {
		return syncAt, nil
	}
}

/**
获取月排名
**/
func GetMonthRanks(x *xorm.Engine) (string, []*model.MonthRank) {

	//月排名
	newMonth := new(model.MonthRank)
	syncAt := time.Now().Format(model.DATE_TIME_FORMAT)
	if has, _ := x.Where("1==1").Desc("month").Limit(1, 0).Get(newMonth); has {
		monthRanks := make([]*model.MonthRank, 0)
		x.Where("month=?", newMonth.Month).Find(&monthRanks)
		syncAt = newMonth.Created.Format(model.DATE_TIME_FORMAT)
		return syncAt, monthRanks
	} else {
		return syncAt, nil
	}
}

/*
模拟排行榜
按照收益率排名
**/
func RankListHandler(ctx *macaron.Context, x *xorm.Engine, redisCli *redis.Client) {
	var page = 0
	if ctx.Params(":page") != "" {
		page, _ = strconv.Atoi(ctx.Params("page"))
	}
	data := listTestRankData(x, page)
	ctx.Data["data"] = data
}

/**
获取模拟排行榜数据
按照最大收益
**/
func listTestRankData(x *xorm.Engine, page int) []*model.RankData {

	sql := "select ua.user_id,u.nick_name,ua.earning_rate," +
		" wr.earning_rate week_rate,mr.earning_rate month_rate," +
		" ua.total_follow from user_account ua  " +
		" left join user u on ua.user_id = u.id " +
		" left join week_rank wr on wr.user_id = ua.user_id" +
		" left join month_rank mr on mr.user_id = ua.user_id " +
		" order by ua.earning desc "

	ranks := make([]*model.RankData, 0)
	x.Sql(sql).Limit(model.PAGE_SIZE, model.PAGE_SIZE*page).Find(&ranks)
	return ranks
}
