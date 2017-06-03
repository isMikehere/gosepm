package handler

import (
	"errors"
	"log"

	macaron "gopkg.in/macaron.v1"

	"../model"
	"github.com/go-macaron/session"
	redis "github.com/go-redis/redis"

	"time"

	"fmt"

	"strconv"

	"github.com/go-xorm/xorm"
)

/**
生成订单
**/
func GenerateOrder(x *xorm.Engine, redis *redis.Client, types int, followUid, followedUId int64) (bool, *model.StockOrder) {

	s := x.NewSession()
	s.Begin()
	defer s.Close()

	product := new(model.Product)
	s.Where("types=?", types).Get(product)
	sn := OrderSnGenerator(redis)
	order := new(model.StockOrder)
	order.OrderSn = sn
	order.UserId = followUid
	order.FollowedId = followedUId
	order.OutTradeNo = "out-" + sn
	order.OrderStatus = 0
	order.OrderAmount = product.Price               //订单金额
	order.PayAmount = product.Price - product.Bonus //待支付金额
	order.ProductType = int8(types)

	if _, err := s.Insert(order); err != nil {
		s.Rollback()
		return false, nil
	}

	sm := new(model.SiteMessage)
	sm.ToUserId = followUid
	sm.Content = fmt.Sprintf("恭喜你下单成功，订单号：%s", sn)
	sm.IsRead = 0
	sm.FromUserId = -1

	if _, err := s.Insert(sm); err != nil {
		s.Rollback()
	}

	s.Commit()
	return true, order
}

/**
更新订单状态
*0：已下单，待支付，1：已支付，2：退款中，3：已退款
**/
func UpdateOrderPayStatus(x *xorm.Engine, orderStatus int, outTradeNo, payType string) error {

	log.Printf("修改订单状态%s,%d,%s", outTradeNo, orderStatus, payType)
	order := GetOrderByOutTradeNo(x, outTradeNo)
	if order != nil && order.OrderStatus == model.ORDER_STATUS_NOT_PAYED {
		order.OrderStatus = model.ORDER_STATUS_PAYED
		order.PayTime = time.Now()
		order.PayType = payType
		x.Id(order.Id).Update(order)

		//更新支付记录
		payLog := GetPayLogByOutTradeNo(x, outTradeNo)
		if payLog != nil {
			payLog.PayTime = time.Now()
			payLog.PayType = payType
			payLog.PayStatus = model.ORDER_STATUS_NOT_PAYED
			x.Id(payLog.Id).Update(payLog)
		}
		return nil
	} else {
		return errors.New(fmt.Sprintf("居然没有找到订单：%s,这不科学 ", outTradeNo))
	}
}

/**
根据外部订单号查询订单
**/
func GetOrderByOutTradeNo(x *xorm.Engine, outTradeNo string) *model.StockOrder {
	order := new(model.StockOrder)
	if has, _ := x.Where("out_trade_no=?", outTradeNo).Get(order); has {
		return order
	} else {
		return nil
	}
}

/**
根据外部订单号查询支付流水记录
**/
func GetPayLogByOutTradeNo(x *xorm.Engine, outTradeNo string) *model.PayLog {

	payLog := new(model.PayLog)
	if has, _ := x.Where("out_trade_no=?", outTradeNo).Get(payLog); has {
		return payLog
	} else {
		return nil
	}
}

/**
更新订单的外部订单号
**/
func UpdateOrderTradeNo(x *xorm.Engine, order *model.StockOrder, outTradeNo string, payType, prepayId string) error {

	payLog := new(model.PayLog)

	if has, _ := x.Where("out_trade_no=?", outTradeNo).Get(payLog); !has {
		payLog.OutTradeNo = outTradeNo
		payLog.PayType = payType
		payLog.PayTime = time.Now()
		payLog.PrepareId = prepayId
		payLog.PayStatus = model.ORDER_STATUS_NOT_PAYED
		payLog.Amount = order.PayAmount
		_, err := x.Insert(payLog)
		Chk(err)
	} else {
		payLog.OutTradeNo = outTradeNo
		_, err := x.Id(payLog.Id).Update(payLog)
		Chk(err)
	}
	order.OutTradeNo = outTradeNo
	_, err := x.Id(order.Id).Update(order)
	return err
}

/**
获取订单列表
**/
func OrderListHandler(ctx *macaron.Context, sess session.Store, x *xorm.Engine, r *redis.Client) {

	status, _ := strconv.Atoi(ctx.Params(":orderStatus"))

	_, user := GetSessionUser(sess)
	ofs := make([]*model.OrderFollow, 0)
	//小于0,所有的订单
	if status < 0 {
		x.Sql("select * from user_follow uf inner join stock_order so "+
			" on uf.order_id = so.id where uf.followed_id = ? ", user.Id).Find(&ofs)
	} else {
		x.Sql("select * from user_follow uf inner join stock_order so "+
			" on uf.order_id = so.id where uf.followed_id = ? and uf.follow_status = ?", user.Id, status).Find(&ofs)
	}
	ctx.Data["r"] = r
	ctx.Data["x"] = x
	ctx.Data["orders"] = ofs
	// ctx.JSON(200, ofs)
	ctx.HTML(200, "my_order")

}
