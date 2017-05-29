package model

import "time"

/**
节假日，不能进行交易的日期列表
**/
type TaskExcludeDate struct {
	Id          int64
	Name        string    `xorm:"varchar(20)"`
	StartDate   time.Time //停盘开始日期
	EndDate     time.Time //复盘日期
	StockCode   string    `xorm:"varchar(10)"` //股票列表
	ExcludeType int8      `xorm:"int(4)"`      //0:全量，1:单股
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
	Version     int       `xorm:"version"`
}

/**
股票停复盘
**/
type StopStocks struct {
	StockCode    string
	StockName    string
	StopStart    string
	StopEnd      string
	StopDuration string
	StopRemark   string
	Url          string
	Checksum     string
	HandleStatus int `xorm:"int(4)"` //0:待处理，1:已处理
}
