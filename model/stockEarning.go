package model

import "time"

/**
每日浮动收益
每日收盘计算、交易完毕后计算
**/
type StockEarning struct {
	Id          int64
	UserId      int64
	StockCode   string  `xorm:"varchar(10)"`
	EarningRate float64 `xorm:"decimal(32,2)"` //加权后 收益率
	StockAmount float64 `xorm:"decimal(32,2)"` //股票市值
	StockNumber int32   `xorm:"int(11)"`       //股票数量（单位：手）
	Remark      string
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
	Version     int       `xorm:"version"`
}
