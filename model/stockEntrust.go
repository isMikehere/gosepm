package model

import "time"

/**
当日委托表
**/
type StockEntrust struct {
	Id             int64
	UserId         int64
	StockCode      string
	TransType      int8 //交易类型0:sale ,1:buy
	EntrustPrice   float64
	EntrustNumber  int32
	RentrustPrice  float64 //实际交易价格
	RentrustNumber int32   //实际交易数量
	EntrustTime    time.Time
	EntrustStatus  int8 //0:已交易，1:委托中，2:已取消，3:没有交易
	Remark         string
	Created        time.Time `xorm:"created"`
	Updated        time.Time `xorm:"updated"`
	Version        int       `xorm:"version"`
}

/**
委托表历史
**/
type StockEntrustHistory struct {
	Id            int64
	UserId        int64
	StockCode     string
	TransType     int8 //交易类型0:sale ,1:buy
	EntrustPrice  float64
	EntrustNumber int32
	EntrustTime   int32
	EntrustStatus int8 //1:已交易，2:已取消
	Remark        string
	Created       time.Time `xorm:"created"`
	Updated       time.Time `xorm:"updated"`
	Version       int       `xorm:"version"`
}
