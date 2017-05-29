package model

import "time"

/**
月排名
**/
type MonthRank struct {
	Id          int64
	UserId      int64
	EarningRate float64 `xorm:"decimal(12,2) not null"`
	Month       string  `xorm:"char(7) not null"`
	Position    int16
	StockCode   string    `xorm:"varchar(10) not null"`
	StockName   string    `xorm:"varchar(32) not null"`
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
	Version     int       `xorm:"version"`
}

/**
周排名
所有已经交易过的用户，
当前用户的所有收益（可用资金+股票市值）向下排

**/
type WeekRank struct {
	Id          int64
	UserId      int64
	EarningRate float64   `xorm:"decimal(12,2) not null"`
	Month       string    `xorm:"char(7) not null"` //2017-01
	Week        string    `xorm:"char(6)"`          //201721
	StockCode   string    `xorm:"varchar(10) not null"`
	StockName   string    `xorm:"varchar(32) not null"`
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
	Version     int       `xorm:"version"`
}

/*
日排名
计算逻辑：单日买进的股票，截止收盘后 股票价格涨幅比例排名，
如果存在比例一致的，则依次下排，获取前20名
*/
type DayRank struct {
	Id          int64
	UserId      int64
	EarningRate float64   `xorm:"decimal(12,2) not null"`
	Day         string    `xorm:"char(10) not null"` //yyyy-mm-dd
	StockCode   string    `xorm:"varchar(10) not null"`
	StockName   string    `xorm:"varchar(32) not null"`
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
	Version     int       `xorm:"version"`
}
