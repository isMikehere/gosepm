package model

import "time"

/**
月排名
**/
type MonthRank struct {
	Id          int64
	UserId      int64
	Rank        int       `xorm:int(11) `
	NickName    string    `xorm:"varchar(32) not null"`
	EarningRate float64   `xorm:"float(12,2) not null"`
	Month       string    `xorm:"char(7) not null"`
	StockCode   string    `xorm:"varchar(10) not null"`
	StockName   string    `xorm:"varchar(32) not null"`
	TotalFollow int64     `xorm:"int(10)  "` // 累计订阅量
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
	Rank        int       `xorm:int(11) `
	NickName    string    `xorm:"varchar(32) not null"`
	EarningRate float64   `xorm:"float(12,2) not null"`
	Month       string    `xorm:"char(7) not null"` //2017-01
	Week        string    `xorm:"char(6)"`          //201721
	StockCode   string    `xorm:"varchar(10) not null"`
	StockName   string    `xorm:"varchar(32) not null"`
	TotalFollow int64     `xorm:"int(10)  "` // 累计订阅量
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
	Rank        int       `xorm:int(11) `
	NickName    string    `xorm:"varchar(32) not null"`
	EarningRate float64   `xorm:"float(12,2) not null"`
	Day         string    `xorm:"char(10) not null"` //yyyy-mm-dd
	StockCode   string    `xorm:"varchar(10) not null"`
	StockName   string    `xorm:"varchar(32) not null"`
	TotalFollow int64     `xorm:"int(10)  "` // 累计订阅量
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
	Version     int       `xorm:"version"`
}

//模拟排行榜实体
type RankData struct {
	Rank        int
	UserId      int64
	NickName    string //昵称
	StockCode   string //最大收益股票
	EarningRate string //总收益率
	DayRate     string //日收益率
	WeekRate    string //周收益率
	MonthRate   string //月收益率
	WeekXTimes  int    //周冠军次数
	TotalFollow int64
}
