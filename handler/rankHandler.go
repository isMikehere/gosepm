//排名计算
package handler

import (
	"fmt"
	"time"

	macaron "gopkg.in/macaron.v1"

	"strconv"

	"../model"
	"github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
)

/*
获取日排名
**/
func GetDayRanks(x *xorm.Engine) (string, []*model.RankData) {

	//日排名
	dailyRanks := make([]*model.DayRank, 0)
	syncAt := time.Now().Format(model.DATE_TIME_FORMAT)
	if x.Where("1=1").Find(&dailyRanks); len(dailyRanks) > 0 {
		rankData := func(dailyRanks []*model.DayRank) []*model.RankData {
			dataRanks := make([]*model.RankData, len(dailyRanks))
			for i, dayRank := range dailyRanks {
				dataRank := new(model.RankData)
				dataRank.Rank = int(dayRank.Id)
				dataRank.UserId = dayRank.UserId
				dataRank.NickName = dayRank.NickName
				dataRank.StockCode = dayRank.StockCode
				dataRank.DayRate = FormatRate(dayRank.EarningRate)
				dataRank.TotalFollow = dayRank.TotalFollow
				dataRanks[i] = dataRank
			}
			return dataRanks
		}(dailyRanks)
		syncAt = dailyRanks[0].Created.Format(model.DATE_TIME_FORMAT)
		if syncAt == "0001-01-01 00:00:00" {
			syncAt = time.Now().Format(model.DATE_TIME_FORMAT)
		}

		return syncAt, rankData

	} else {
		return syncAt, nil
	}

}

/**
获取周排名
**/
func GetWeekRanks(x *xorm.Engine) (string, []*model.RankData) {

	// //周排名
	newWeek := new(model.WeekRank)
	syncAt := time.Now().Format(model.DATE_TIME_FORMAT)
	if has, _ := x.Where("1=1").Desc("week").Limit(1, 0).Get(newWeek); has {
		ranks := make([]*model.WeekRank, 0)
		x.Where("week = ?", newWeek.Week).Find(&ranks)
		rankData := func(ranks []*model.WeekRank) []*model.RankData {
			dataRanks := make([]*model.RankData, len(ranks))
			for i, rank := range ranks {
				dataRank := new(model.RankData)
				dataRank.Rank = i + 1
				dataRank.UserId = rank.UserId
				dataRank.NickName = rank.NickName
				dataRank.StockCode = rank.StockCode
				dataRank.WeekRate = FormatRate(rank.EarningRate)
				dataRank.TotalFollow = rank.TotalFollow
				dataRanks[i] = dataRank
			}
			return dataRanks
		}(ranks)
		syncAt = newWeek.Created.Format(model.DATE_TIME_FORMAT)
		if syncAt == "0001-01-01 00:00:00" {
			syncAt = time.Now().Format(model.DATE_TIME_FORMAT)
		}
		return syncAt, rankData
	} else {
		return syncAt, nil
	}
}

/**
获取月排名
**/
func GetMonthRanks(x *xorm.Engine) (string, []*model.RankData) {

	//月排名
	newMonth := new(model.MonthRank)
	syncAt := time.Now().Format(model.DATE_TIME_FORMAT)
	if has, _ := x.Where("1=1").Desc("month").Limit(1, 0).Get(newMonth); has {
		ranks := make([]*model.MonthRank, 0)
		x.Where("month=?", newMonth.Month).Desc("earning_rate").Find(&ranks)
		rankData := func(ranks []*model.MonthRank) []*model.RankData {
			dataRanks := make([]*model.RankData, len(ranks))
			for i, rank := range ranks {
				dataRank := new(model.RankData)
				dataRank.Rank = i + 1
				dataRank.UserId = rank.UserId
				dataRank.NickName = rank.NickName
				dataRank.StockCode = rank.StockCode
				dataRank.MonthRate = FormatRate(rank.EarningRate)
				dataRank.TotalFollow = rank.TotalFollow
				dataRanks[i] = dataRank
			}
			return dataRanks
		}(ranks)
		syncAt = newMonth.Created.Format(model.DATE_TIME_FORMAT)
		if syncAt == "0001-01-01 00:00:00" {
			syncAt = time.Now().Format(model.DATE_TIME_FORMAT)
		}
		return syncAt, rankData

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
		if page >= 1 {
			page--
		}

	}
	paginator, data := listTestRankData(x, page)
	//周冠军次数
	data = func(d []*model.RankData) []*model.RankData {
		for _, rank := range d {
			rank.WeekXTimes = countWeekXTimes(x, rank.UserId)
		}
		return d
	}(data)
	ctx.Data["ranks"] = data
	ctx.Data["paginator"] = paginator
	ctx.HTML(200, "rank")
}

/**
获取模拟排行榜数据
按照最大收益
**/
func listTestRankData(x *xorm.Engine, page int) (map[string]interface{}, []*model.RankData) {

	c := new(struct {
		Count int64
	})

	countSQL := "select count(1) count from user_account ua  " +
		" left join user u on ua.user_id = u.id " +
		" left join week_rank wr on wr.user_id = ua.user_id" +
		" left join month_rank mr on mr.user_id = ua.user_id "
	if _, err := x.Sql(countSQL).Get(c); err == nil {
		paginator := Paginator(page, c.Count)
		sql := "select ua.user_id,u.nick_name,ua.earning_rate," +
			" wr.earning_rate week_rate,mr.earning_rate month_rate," +
			" ua.total_follow from user_account ua " +
			" left join user u on ua.user_id = u.id " +
			" left join week_rank wr on wr.user_id = ua.user_id" +
			" left join month_rank mr on mr.user_id = ua.user_id " +
			" order by ua.earning desc limit ?,?"

		ranks := make([]*model.RankData, 0)
		start := paginator["startIndex"]
		x.Sql(sql, start, model.PAGE_SIZE).Find(&ranks)
		return paginator, ranks
	} else {
		fmt.Errorf("查询排名列表异常：%s", err.Error())
	}
	return nil, nil
}

/**
周冠军次数
**/
func countWeekXTimes(x *xorm.Engine, uid int64) int {
	c, _ := x.Where("user_id = ?", uid).Count(new(model.WeekRank))
	return int(c)
}
