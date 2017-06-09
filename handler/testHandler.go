package handler

import (
	"fmt"

	"strconv"

	"../model"
	"github.com/go-xorm/xorm"
	"gopkg.in/macaron.v1"
)

/**
测试
**/
func TestPage(ctx *macaron.Context, x *xorm.Engine) {

	page := ctx.Params(":page")
	fmt.Printf("page:%s\n", page)
	i, _ := strconv.Atoi(page)
	//数据分页
	//统计条数
	count, _ := x.Where("1=1").Count(new(model.Stock))
	mappp := Paginator(i, count)
	stocks := make([]*model.Stock, 0)
	x.Where("1=1").Limit(model.PAGE_SIZE, mappp["startIndex"].(int)).Find(&stocks)
	ctx.Data["paginator"] = mappp
	// ctx.HTML(200, "pageable")
	ctx.JSON(200, stocks)
}
