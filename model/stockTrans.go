package model

import "time"

/**
交易流水
**/
type StockTrans struct {
	Id           int64
	UserId       int64
	StockCode    string  //股票代码
	TransType    int8    //交易类型 0:sale ,1:buy
	TransStatus  int8    //交易状态 1:成功，0:失败
	TransFrom    int8    //交易来源 0:web,1:ios,2:andriod
	TransNumber  int32   //交易笔数（手）
	TransPrice   float64 //交易价格
	NotifyStatus int8    //0:没有通知，1:待通知，2:通知成功，3:通知失败
	TransTime    time.Time
	Remark       string
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
	Version      int       `xorm:"version"`
}
