package handler

import (
	"strconv"

	"strings"

	"time"

	"../model"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
)

func TrxHandler(ctx *macaron.Context) {
	ctx.HTML(200, "deal")
}

/**
交易成功率比数据
**/
func TrxRateDataChartHander(ctx *macaron.Context, r *redis.Client, x *xorm.Engine) {

	jr := new(model.JsonResult)
	uid := ctx.Params(":uid")
	type Data struct {
		Value int32  `json:"value"`
		Name  string `json:"name"`
	}
	var datas = make([]*Data, 0)
	ua := new(model.UserAccount)
	data1 := new(Data)
	data2 := new(Data)

	if has, _ := x.Where("user_id=?", uid).Get(&ua); has {

		data1.Name = "盈利笔数"
		data1.Value = ua.SuccessTimes
		data2.Name = "亏损笔数"
		data2.Value = ua.TotalTimes - ua.SuccessTimes
	}
	datas = append(datas, data1, data2)
	jr.Code = "200"
	jr.Data = datas
	ctx.JSON(200, jr)
}

/**
排名图表
**/
func RankDataChartHandler(ctx *macaron.Context, x *xorm.Engine) {

	jr := new(model.JsonResult)
	jr.Code = "100"
	uid := ctx.Params(":uid")
	//
	weekRanks := make([]*model.WeekRank, 0)
	if err := x.Where("user_id=?", uid).OrderBy("week").Limit(5, 0).Find(&weekRanks); err == nil && len(weekRanks) > 0 {
		datas := make([][]int, 0)
		for _, v := range weekRanks {
			arr := make([]int, 0)
			weekInt, _ := strconv.Atoi(strings.SplitAfterN(v.Week, "", 5)[4])
			arr = append(arr, weekInt, v.Rank)
			datas = append(datas, arr)
		}
		jr.Code = "200"
		jr.Data = datas
	} else {
		_, week := time.Now().ISOWeek()
		mapp := Paginator(week, 52)
		weeks := mapp["pages"].([]int)
		datas := make([][]int, 0)
		for _, v := range weeks {
			arr := make([]int, 0)
			arr = append(arr, v, 0)
			datas = append(datas, arr)
		}
		jr.Code = "200"
		jr.Data = datas

	}

	ctx.JSON(200, jr)
}
