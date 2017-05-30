package handler

import (
	"../model"
	"github.com/go-xorm/xorm"
)

/**
首页新闻动态
**/
func IndexNews(x *xorm.Engine) []*model.News {
	news := make([]*model.News, 0)
	x.Where("is_online=?", 1).Desc("created").Limit(10, 0).Find(&news)
	return news
}
