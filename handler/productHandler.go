package handler

import (
	"log"
	"strconv"

	macaron "gopkg.in/macaron.v1"

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
	has, _ := x.Where("types=?", types).And("is_online=?", 1).Get(product)
	if has {
		return has, product
	} else {
		return false, nil
	}
}

/**
获取产品
**/
func GetProductHandler(ctx *macaron.Context, x *xorm.Engine) {

	r := new(model.JsonResult)
	typeId, _ := strconv.Atoi(ctx.Params(":type"))
	p := new(model.Product)
	if has, _ := x.Sql("select * from product where types=? and is_online=1", int8(typeId)).Get(p); has {
		r.Code = "200"
		r.Data = p
	} else {
		r.Code = "100"
		r.Msg = "没有此产品"
	}
	ctx.JSON(200, r)
}
