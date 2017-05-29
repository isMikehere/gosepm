package model

import "time"

//用户角色
const (
	ADMIN    string = "ADMIN"    //管理员
	CUSTOMER string = "CUSTOMER" //消费者
)

type Role struct {
	Id           int64
	AddTime      time.Time
	DeleteStatus int8
	Display      bool
	Info         string
	RoleCode     string
	RoleName     string
	Sequence     int8
	Types        string
}

/**
关联表
**/
type UserRole struct {
	UserId int64
	RoleId int64

	User `xorm:"extends"`
	Role `xorm:"extends"`
}

type Res struct {
	Id      int64
	ResName string
	Value   string
	Remark  string
}

/**
资源角色关联表
**/
type RoleRes struct {
	RoleId int64
	ResId  int64

	Role `xorm:"extends"`
	Res  `xorm:"extends"`
}
