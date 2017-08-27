package model

import (
	"time"
)

//用户状态
const (
	USER_STATUS_OK int = iota
	USER_STATUS_LOCK
	USER_STATUS_DELETE
)

/**
用户表
**/
type User struct {
	Id                int64
	UserName          string `xorm:"varchar(32)"`
	TrueName          string `xorm:"varchar(128)"`
	Sex               int    `xorm:"bit(1)"` //0:F,1:M
	Mobile            string
	MobilePayPassword string
	Address           string `xorm:"varchar(128)"`
	Birthday          time.Time
	IdCard            string `xorm:"varchar(128)"` //身份证
	Email             string `xorm:"varchar(128)"`
	LastLoginDate     time.Time
	LastLoginIp       string `xorm:"varchar(128)"`
	LoginCount        int    `xorm:"varchar(128)"`
	NickName          string `xorm:"varchar(128)"`
	OpenId            string `xorm:"varchar(128)"`
	Password          string `xorm:"varchar(128)"`
	Salt              string `xorm:"varchar(128)"` //salt
	AlipayId          string `xorm:"varchar(128)"` //阿里pay
	QqOpenid          string
	Qq                string `xorm:"varchar(128)"`
	SinaOpenid        string
	AppLoginLoken     string `xorm:"varchar(128)"`
	UserStatus        int    `xorm:"int(4)"` //0:ok,1:lock,2:delete,
	Telephone         string `xorm:"varchar(128)"`
	UserMark          string `xorm:"varchar(128)"`
	UserRole          string `xorm:"varchar(20)"` //admin,customer
	AreaId            int64
	ReadContract      int8 `xorm:"int(4)"` //是否阅读协议 0:否，1:是
	PhotoId           int64
	Created           time.Time `xorm:"created"`
	Updated           time.Time `xorm:"updated"`
	Version           int       `xorm:"version"`
}

/**
用户的账户信息
**/
type UserAccount struct {
	Id              int64
	UserId          int64
	UserName        string    `xorm:"varchar(32)"`   //用户名
	BankCardNo      string    `xorm:"varchar(19)"`   //银行卡号
	BankAccountName string    `xorm:"varchar(20)"`   //银行名称
	InitAmount      float64   `xorm:"decimal(32,2)"` //初始化金额
	AvailableAmount float64   `xorm:"decimal(32,2)"` //可用金额
	LockAmount      float64   `xorm:"decimal(32,2)"` //锁定金额
	Gold            int       `xorm:"int(10) "`      //金币数量
	Integral        int       `xorm:"int(10) "`      //积分
	EarningRate     float64   `xorm:"decimal(16,2)"` //收益率,
	Earning         float64   `xorm:"decimal(32,2)"` //总收益（总市值+可用金额-初始化金额）
	StockAmount     float64   `xorm:"decimal(32,2)"` //股票总市值
	TotalAmount     float64   `xorm:"decimal(32,2)"` //总资产(总市值+可用金额)
	TransFrequency  float64   `xorm:"decimal(16,2)"` //交易频率,
	TotalTimes      int32     `xorm:"decimal(16,2)"` //失败次数,
	WeekTimes       int32     `xorm:"int(11)"`       //周第一名次数,
	MonthTimes      int32     `xorm:"int(11)"`       //月第一名次数,
	SuccessTimes    int32     `xorm:"decimal(16,2)"` //成功率,
	SuccessRate     float64   `xorm:"decimal(16,2)"` //成功率,
	Rank            int       `xorm:"int(11)"`       //总收益排名
	XCode           string    `xorm:"varchar(10)"`   //近30天收益最大的股票
	TransStart      time.Time //首次交易时间
	TotalFollow     int       `xorm:"int(10) "`    //订阅量
	UserLevel       int8      `xorm:"varchar(20)"` // 用户级别;1:小白；2：熟客；3：高手；4：骨灰
	Created         time.Time `xorm:"created"`
	Updated         time.Time `xorm:"updated"`
	Version         int       `xorm:"version"`
}

/**
用户当前订阅表
**/
type UserFollow struct {
	Id           int64
	UserId       int64     //订阅人
	FollowedId   int64     //被订阅人
	FollowType   int8      `xorm:"int(8)"` //0：周，1：月
	FollowStatus int8      `xorm:"int(8)"` //0:订阅中，1:待通知，2:订阅结束，3:已退订
	FollowStart  time.Time //订阅开始
	FollowEnd    time.Time //订阅结束
	OrderId      int64     //订单ID
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
	Version      int       `xorm:"version"`
}
