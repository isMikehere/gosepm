package model

import "time"

const (
	DriverOfMysql     = "mysql"
	DataSourceOfMysql = "root:root@tcp(127.0.0.1:3306)/sepm?parseTime=true"
)

const (
	FOLLOW_OK_MSG       = "尊敬的客户%s：您好，您已经成功订阅高手%s的为期%d的股票提醒，有效期为%s-%s,如有问题，请联系我们电话：%s" //订阅成功通知订阅人
	TOBEFOLLOWED_OK_MSG = "尊敬的客户%s：您好，%s已经成功订阅您的为期%d周股票提醒，有效期为%s-%s,如有问题，请联系我们电话：%s"   //订阅成功通知被订阅人
	TRX_NOTIFY_MSG      = "尊敬的客户%s：您好，您订阅的用户%s于%s %s了股票%s %d 手 ,如有问题，请联系我们电话：%s"       //订阅成功通知被订阅人

)
const (
	HOT_LINE         = "020-000000000"
	DATE_TIME_FORMAT = "2006-01-02 15:04:05"
	DATE_FORMAT      = "2006-01-02"
	DATE_MONTH       = "2006-01"
	DATE_FORMAT_1    = "2006-01-02 00:00:00"
	DATE_FORMAT_2    = "2006-01-02 00:00"
	TIME_FORMAT      = "15:04:05"
)

const (
	MAX_DATE = "2099-01-01 00:00:01"
)
const (
	INIT_AMOUNT float64 = 200000 //初始资金
)

//交易时间
const (
	TRX_TIME1_START = "09:30:00" //早盘开市
	TRX_TIME1_END   = "11:30:00" //早盘结束
	TRX_TIME2_START = "13:00:00" //午盘开始
	TRX_TIME2_END   = "15:00:00" //午盘结束
)

//清算时间
const (
	SETTLEMENT_START = "15:20:00" //settle start time
	SETTLEMENT_END   = "15:45:00" //settle end time
)

const (
	TRX_SHORT_OF_MONEY     string = "资金不足"
	TRX_STOCK_NOT_SALE     string = "股票停盘"
	TRX_STOCK_INC_TOP      string = "股票涨停"
	TRX_STOCK_DEC_TOP      string = "股票跌停"
	TRX_STOCK_ENT_OK       string = "委托成功"
	TRX_STOCK_ENT_FAIL     string = "委托失败"
	TRX_STOCK_CAN_ENT_OK   string = "撤销成功"
	TRX_STOCK_CAN_ENT_FAIL string = "撤销失败"
	TRX_STOCK_OK           string = "撤销成功"
	TRX_STOCK_FAIL         string = "撤销失败"
	TRX_STOCK_ENT_MORE     string = "可售持仓数量小于委托数量"
	TRX_STOCK_X_PRICE      string = "委托价格不能大于涨停价格"
	TRX_STOCK_O_PRICE      string = "委托价格不能小于跌停价格"
)

//新浪股票api
const (
	STOCK_5_STAGES_API = "http://hq.sinajs.cn/"
	EXC_SH             = "sh"
	EXC_SZ             = "sz"
)
const (
	SYS_ERR_Q_STOCK string = "获取股票信息异常"
)

const (
	MATCH_LIMIT int = 1000
)

const (
	RANK_SIZE int = 20
)

const ONE_SECOND = 1*time.Second + 10*time.Millisecond
const ONE_HOUR = 1*time.Hour + 10*time.Millisecond

const (
	TASK_SUCCESS int8 = iota
	TASK_FAIL
)

const (
	TASK_DAILY   = "daily" //一周一次，下同理
	TASK_WEEKLY  = "weekly"
	TASK_MONTHLY = "monthly"
	TASK_YEARLY  = "yearly"
	TASK_POLLING = "polling" //timer polling
)

const (
	R_KEY_USERS  = "users"
	R_KEY_STOCKS = "stocks"
)
