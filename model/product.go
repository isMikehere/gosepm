package model

import "time"

/**
产品表
**/
type Product struct {
	Id          int64
	Name        string    `xorm:"varchar(32)"`   //名字
	Type        int8      `xorm:"int(8)"`        //0：周，1：月
	Price       float64   `xorm:"decimal(32,2)"` //价格
	Bonus       float64   `xorm:"decimal(32,2)"` //优惠
	description string    `xorm:"varchar(128)"`  //描述
	IsOnline    int       `xorm:int(4)`          //是否上线 0:下线，1:上线
	Created     time.Time `xorm:"created"`
	Updated     time.Time `xorm:"updated"`
	Version     int       `xorm:"version"`
}
