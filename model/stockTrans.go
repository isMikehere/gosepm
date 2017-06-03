package model

import "time"

/**
交易流水
**/
type StockTrans struct {
	Id           int64
	UserId       int64   `xorm:"bigint(20) notnull index(idx_user_code)"`
	StockCode    string  `xorm:"varchar(10) notnull index(idx_user_code)"` //股票代码
	TransType    int8    `xorm:"int(4) index(idx_trx_status)"`             //交易类型 0:sale ,1:buy
	TransStatus  int8    `xorm:"int(4) index(idx_trx_status)"`             //交易状态 1:成功，0:失败
	TransFrom    int8    `xorm:"int(4)"`                                   //交易来源 0:web,1:ios,2:andriod
	TransNumber  int32   `xorm:"int(10)"`                                  //交易笔数（手）
	TransPrice   float64 `xorm:"decimal(16,2)"`                            //交易价格
	NotifyStatus int8    `xorm:"int(4)"`                                   //0:没有通知，1:待通知，2:通知成功，3:通知失败
	TransTime    time.Time
	Remark       string
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
	Version      int       `xorm:"version"`
}

//股票交易对
type StockTrxInfo struct {
	Id           int64
	UserId       int64     `xorm:"index(idx_uid_code)"`                     //交易者
	StockCode    string    `xorm:"varchar(10) notnull index(idx_uid_code)"` //股票代码()
	BuyPrice     float64   `xorm:"decimal(16,2)"`                           //交易价格
	Earning      float64   `xorm:"decimal(16,2)"`                           //盈亏
	TransNumber  int32     //交易笔数（手） 总量
	LeftNumber   int32     //交易笔数（手） 剩余量
	TransStatus  int8      `xorm:"int(4)"` //交易状态 0交易中，1：已结束
	BuyTime      time.Time //买入时间
	SaleTime     time.Time //卖出
	NotifyStatus int8      `xorm:"int(4)"` //交易状态 0:买入待通知,1:买入已通知,2:卖出待通知，3:卖出已通知
	LastTrxId    int64     //最后一次交易ID
	Remark       string
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
	Version      int       `xorm:"version"`
}
