package handler

import (
	"fmt"

	"../model"
	"github.com/go-xorm/xorm"
)

/**
首页新闻动态
**/
func IndexNews(x *xorm.Engine) []*model.News {
	news := make([]*model.News, 0)
	x.Where("is_online=?", true).Desc("created").Limit(10, 0).Find(&news)
	fmt.Printf("%d", len(news))
	return news
}
