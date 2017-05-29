package model

import "time"

/**
股票模型-->
**/
type Stock struct {
	Id        int64
	Location  string `xorm:"char(10) not null "`
	StockName string `xorm:"varchar(32) not null"`
	StockCode string `xorm:"varchar(10) not null"`
}

/**
 股票历史日交易统计表
**/
type PostStock struct {
	Id          int64
	StockName   string  `xorm:"varchar(32)"`
	StockCode   string  `xorm:"varchar(10) unique(idx_code_date)"`
	Sdate       string  `xorm:"varchar(10) unique(idx_code_date)"`
	OpenPrice   float64 `xorm:"decimal(16,3) "`
	EndPrice    float64 `xorm:"decimal(16,3) "`
	TopPrice    float64 `xorm:"decimal(16,3) "`
	BottomPrice float64 `xorm:"decimal(16,3) "`
	TransNumber int64   `xorm:"bigint(32) "`
	TransAmount float64 `xorm:"decimal(32,3) "`
}

/**
 股票历史日交易详细表
**/
type DayPostStock struct {
	Id          int64
	StockName   string  `xorm:"varchar(32)"`
	StockCode   string  `xorm:"varchar(10) unique(idx_code_date)"`
	DealDate    string  `xorm:"varchar(10) unique(idx_code_date)"`
	DealTime    string  `xorm:"varchar(8)"`
	DealPrice   float64 `xorm:"decimal(16,3) "`
	TransNumber int64   `xorm:"bigint(32) "`
	TransAmount float64 `xorm:"decimal(32,3) "`
	TransType   int8    `xorm:"smallint(4)  "`
}

/**
即时股票数据
**/
type CurrentStockDetail struct {
	Id           int64
	StockCode    string    `xorm:"varchar(10) unique(idx_stock_code)"`
	CurrentPrice float64   `xorm:"decimal(16,2)"` //当前价格
	Detail       string    //5档
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
	Version      int       `xorm:"version"`
}
