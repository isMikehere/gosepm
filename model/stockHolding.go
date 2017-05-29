package model

import (
	"fmt"
	"time"
)

/**
持仓表 ，表中股票存在三种类型数据
1、今日前买进股票
2、今日前买进的加权平均后的股票
3、今日买进的股票
4、相同一个只股票，今日买进的股票不能进行加权平均
**/

type StockHolding struct {
	Id              int64
	UserId          int64
	StockCode       string
	StockNumber     int32     //股票数量 （单位：手）
	AvailableNumber int32     //当前可委托交易的量=stockNumber-(委托数量)
	HoldingStatus   int32     //持仓状态 1:持仓中，0：交易结束(所有的都已经交易)
	TransPrice      float64   //成本价
	TransTime       time.Time //买进时间
	Remark          string
	Created         time.Time `xorm:"created"`
	Updated         time.Time `xorm:"updated"`
	Version         int       `xorm:"version"`
}

/**
**/
func (s *StockHolding) ToString() string {
	return fmt.Sprintln("%d,%s", s.Id, s.StockCode)
}
