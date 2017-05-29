package model

import "time"

/*
系统参数配置表
*/
type SysParam struct {
	Id         int64
	ParamCode  string
	ParamValue string
	ParamInfo  string
	SysCode    string
	Sort       string
	DataStyle  string
	Creator    string
	CreateTime string
	Mender     string
	MendTime   string
	Remark     string
}

/*
 文件表
*/
type FileLog struct {
	Id           int64
	UserId       int64
	FileName     string
	FileType     string    //文件类型
	FileExt      string    //文件后缀
	RelativePath string    `xorm:"varchar(512) not null"` //相对路径 nginxServerIp:port/$this/ --->访问
	AbsolutePath string    `xorm:"varchar(512) not null"` //绝对路径
	Created      time.Time `xorm:"created"`
	Updated      time.Time `xorm:"updated"`
	Version      int       `xorm:"version"`
}
