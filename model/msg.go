package model

import "time"

/**
通知对应表
**/
type NotifyFollow struct {
	Id           int64
	FollowUserId int64     `xorm:"bigint(20) notnull index(idx_user_id)"` //订阅用户ID
	TrxInfoId    int64     `xorm:"bigint(20) notnull index(idx_info_id)"` //交易对ID
	LastMsgId    int64     `xorm:"bigint(20) index(idx_msg_id)"`          //上次通知ID
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
	Version      int       `xorm:"version"`
}

/**
短信通知记录
**/
type MessageLog struct {
	Id         int64
	InBatchId  string    `xorm:"varchar(20) notnull index(idx_inbatch_id)"` //内部批次号
	RetBatchId string    `xorm:"varchar(10) index(idx_outbatch_id)`         //外部批次号
	Mobile     string    `xorm:"varchar(11) notnull index(idx_mobile)"`     //手机号
	Detail     string    `xorm:"varchar(512)"`                              //点击链接看到的内容详情
	Content    string    `xorm:"varchar(512) notnull"`                      //发送内容
	ErrMsg     string    `xorm:"varchar(512) notnull"`                      //失败原因
	SendStatus int8      `xorm:"int(8) notnull index(idx_status)" `         //发送状态  0：待发送，1:提交发送成功，2:提交发送失败
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
	IsRead     int8      `xorm:"int(4) not null"`       //是否已读，0：未读，1：已读
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
	IsOnline int       `xorm:"int(4) not null"`       //是否显示到页面 0:不显示 ，1：显示
	Created  time.Time `xorm:"created"`
	Updated  time.Time `xorm:"updated"`
	Version  int       `xorm:"version"`
}

/**
json result
**/
type JsonResult struct {
	Code string //100:fail，200:success
	Data interface{}
	Msg  string
}
