package handler

import (
	"log"

	"../model"
	"github.com/go-xorm/xorm"
)

/**
获取上线产品列表
**/
func ListOnlineProducts(x *xorm.Engine, log *log.Logger) []*model.Product {
	products := make([]*model.Product, 0)
	x.Where("is_online=?", 1).Find(&products)
	return products
}

/**
查询一个上线的商品
**/
func QueryProductByType(types int, x *xorm.Engine) (bool, *model.Product) {

	product := new(model.Product)
	has, _ := x.Where("type=?", types).And("is_online=?", 1).Limit(1, 0).Get(product)
	if has {
		return has, product
	} else {
		return false, nil
	}
}
