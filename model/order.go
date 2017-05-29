package model

import "time"

const (
	ORDER_STATUS_NOT_PAYED int = iota
	ORDER_STATUS_PAYED
	ORDER_STATUS_REFUNDING
	ORDER_STATUS_REFUNDED
)

const (
	ALIPAY string = "alipay"
	WXPAY  string = "weixin"
)

/**
订单表
**/
type StockOrder struct {
	Id             int64
	UserId         int64     //订阅人
	FollowedUserId int64     //被订阅人
	OrderSn        string    `xorm:"varchar(32) not null"`             //内部订单流水号
	OutTradeNo     string    `xorm:"varchar(32) not null"`             //外部订单号 ,发送支付平台
	OrderStatus    int       `xorm:"int(8) not null"`                  //0：已下单，待支付，1：已支付，2：退款中，3：已退款
	OrderAmount    float64   `xorm:"decimal(16,2) default 0 not null"` //订单金额
	BonusAmount    float64   `xorm:"decimal(16,2) default 0"`          //红包金额
	PayAmount      float64   `xorm:"decimal(16,2) not null"`           //支付金额 == 订单金额-红包金额
	PayType        string    `xorm:"varchar(20) not null"`             //alipay,weixin,
	ProductType    int8      `xorm:"int(8)"`                           //0：周，1：月
	PayTime        time.Time //支付时间
	Created        time.Time `xorm:"created"`
	Updated        time.Time `xorm:"updated"`
	Version        int       `xorm:"version"`
}

/**
支付流水
**/
type PayLog struct {
	Id         int64
	OrderId    int64
	OutTradeNo string    `xorm:"varchar(32) not null"`
	PayType    string    `xorm:"varchar(20) not null"`   //alipay,weixin
	Amount     float64   `xorm:"decimal(16,2) not null"` //金额
	PayStatus  int       `xorm:"int(4) notnull"`         //支付状态
	PrepareId  string    `xorm:"varchar(32) null"`       //微信预支付订单号
	PayTime    time.Time //支付时间
	Created    time.Time `xorm:"created"`
	Updated    time.Time `xorm:"updated"`
	Version    int       `xorm:"version"`
}

/**
退款流水
**/
type RefundLog struct {
	Id        int64
	UserId    int64
	PayId     int64     //支付ID
	PayStatus int8      `xorm:"int(8) not null"`        //0:退款中，1：已退款
	RefundSn  string    `xorm:"varchar(32) not null"`   //退款流水号
	Amount    float64   `xorm:"decimal(16,2) not null"` //金额
	Created   time.Time `xorm:"created"`
	Updated   time.Time `xorm:"updated"`
	Version   int       `xorm:"version"`
}
