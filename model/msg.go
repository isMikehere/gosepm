package model

import "time"

/**
交易提醒记录
**/
type NotifyLog struct {
	Id         int64
	TrxId      int64     `xorm:"int(11)"` //交易记录ID
	FollowedId int64     //被订阅人
	FollowerId int64     //订阅人
	Content    string    `xorm:"longtext"` //通知内容
	Created    time.Time `xorm:"created"`
	Updated    time.Time `xorm:"updated"`
	Version    int       `xorm:"version"`
}

/**
短信通知记录
**/
type MessageLog struct {
	Id         int64
	Mobile     string    `xorm:"varchar(11) not null"`  //手机号
	Content    string    `xorm:"varchar(512) not null"` //发送内容
	SendStatus int8      `xorm:"int(8) not null"`       //发送状态 1:提交发送成功，0：提交发送失败
	Created    time.Time `xorm:"created"`
	Updated    time.Time `xorm:"updated"`
	Version    int       `xorm:"version"`
}

/**
站内信息
**/
type SiteMessage struct {
	Id         int64
	FromUserId int64     // 发送人
	ToUserId   int64     // 目标人
	Content    string    `xorm:"varchar(512) not null"` //发送内容
	IsRead     bool      `xorm:"bit(1) not null"`       //是否已读，0：未读，1：已读
	Created    time.Time `xorm:"created"`
	Updated    time.Time `xorm:"updated"`
	Version    int       `xorm:"version"`
}

/**
 滚动新闻
**/
type News struct {
	Id       int64
	Type     int8      `xorm:"int(8) notnull index"`  //新闻类型 0:行业新闻，1：订阅动态
	Title    string    `xorm:"varchar(512) not null"` //新闻标题
	Content  string    `xorm:"varchar(512) "`         //新闻内容
	IsOnline bool      `xorm:"bit(1) not null"`       //是否显示到页面 0:不显示 ，1：显示
	Created  time.Time `xorm:"created"`
	Updated  time.Time `xorm:"updated"`
	Version  int       `xorm:"version"`
}

/**
json result
**/
type JsonResult struct {
	Code string //100:fail，200:success
	Data string
	Msg  string
}
