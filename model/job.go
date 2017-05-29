package model

import "time"

type Task struct {
	Id            int64
	Name          string    `xorm:"varchar(32)"` //任务名
	Crontab       string    `xorm:"varchar(64)"` //定时任务执行cron str
	Func          string    `xomr:"varchar(32)"` //执行方法名,反射调用
	Enable        int8      `xorm:"int(4)"`      //0:disable，1:enable
	RunType       string    `xorm:"varchar(20)"` //daily(一周一次，下同理),weekly,monthly,yearly,polling(timer 轮询 )
	LastRuntime   time.Time //上次运行时间
	LastRunResult int8      `xorm:"int(4)"`   //0:成功，1:失败
	LastRunMsg    string    `xorm:"longtext"` //执行结果
	Created       time.Time `xorm:"created"`
	Updated       time.Time `xorm:"updated"`
	Version       int       `xorm:"version"`
}
