package handler

import (
	"errors"
	"log"

	"../model"

	"time"

	"fmt"

	"github.com/go-xorm/xorm"
)

/**
更新订单状态
*0：已下单，待支付，1：已支付，2：退款中，3：已退款
**/
func UpdateOrderPayStatus(x *xorm.Engine, log *log.Logger, orderStatus int, outTradeNo, payType string) error {

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
