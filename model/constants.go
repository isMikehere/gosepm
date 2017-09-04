package model

import "time"

const (
	DriverOfMysql = "mysql"
	DataSourceOfMysql = "root:zhsepm!@#$%@tcp(106.14.112.157:3306)/sepm?parseTime=true"
	// DataSourceOfMysql = "root:root@tcp(localhost:3306)/sepm?parseTime=true"
	RedisHost = "106.14.112.157:6379"
	// RedisHost = "localhost:6379"
	RedisPass = "xceof"
	HOST = "106.14.112.157"
)

const (
	REGISTER_MSG = "验证码：%s 您正在注册金修网络个人账号，请勿泄露校验码给任何人，以免造成账户或资金损失，5分钟后失效。" //注册
	// FOLLOW_OK_MSG       = "恭喜您订阅成功，订单有效期为%d周。请保持手机畅通并及时查看你的订阅信息，详情请参考订阅须知。祝您投资顺利！"     //订阅成功通知订阅人
	// TOBEFOLLOWED_OK_MSG = "验证码：10000 您正在注册金修网络个人账号，请勿泄露校验码给任何人，以免造成账户或资金损失，5分钟后失效。" //订阅成功通知被订阅人
	TOBEFOLLOWED_OK_MSG = "恭喜您：%s已经成功订阅您的为期%d周模拟交易提醒，有效期为%s-%s。请及时处理详情请参考订单须知"     //订阅成功通知被订阅人
	TRX_NOTIFY_DETAIL = "尊敬的用户:您在模拟交易中订阅的用户%s在%s买入%s，成交价格%s %s股。以上数据来自于模拟交易仅供参考" //交易提醒
	TRX_NOTIFY_MSG = "尊敬的用户:您在本平台中订阅的用户%s有了新交易动态，查看请点击链接 %s"                  //交易提醒
)
const (
	HOT_LINE = "020-000000000"
	DATE_TIME_FORMAT = "2006-01-02 15:04:05"
	DATE_FORMAT = "2006-01-02"
	DATE_MONTH = "2006-01"
	DATE_FORMAT_1 = "2006-01-02 00:00:00"
	DATE_FORMAT_2 = "2006-01-02 00:00"
	TIME_FORMAT = "15:04:05"
	DATE_ORDER_FORMAT = "20060102150405"
)
const (
	MSG_EXPIRE_DURATION = time.Minute * 5
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
	TRX_TIME1_END = "11:30:00" //早盘结束
	TRX_TIME2_START = "13:00:00" //午盘开始
	TRX_TIME2_END = "15:00:00" //午盘结束
)

//清算时间
const (
	SETTLEMENT_START = "15:20:00" //settle start time
	SETTLEMENT_END = "15:45:00" //settle end time
)

const (
	TRX_SHORT_OF_MONEY string = "资金不足"
	TRX_PARAM_NULL string = "参数为空"
	TRX_STOCK_NOT_SALE string = "股票停盘"
	TRX_STOCK_INC_TOP string = "股票涨停"
	TRX_STOCK_DEC_TOP string = "股票跌停"
	TRX_STOCK_ENT_OK string = "委托成功"
	TRX_STOCK_ENT_FAIL string = "委托失败"
	TRX_STOCK_QTY string = "购买数量为100的整数倍"
	TRX_STOCK_CAN_ENT_OK string = "撤销成功"
	TRX_STOCK_CAN_ENT_FAIL string = "撤销失败"
	TRX_STOCK_OK string = "撤销成功"
	TRX_STOCK_FAIL string = "撤销失败"
	TRX_STOCK_ENT_MORE string = "可售持仓数量小于委托数量"
	TRX_STOCK_X_PRICE string = "委托价格不能大于涨停价格"
	TRX_STOCK_O_PRICE string = "委托价格不能小于跌停价格"
)

//新浪股票api
const (
	STOCK_5_STAGES_API = "http://hq.sinajs.cn/"
	EXC_SH = "sh"
	EXC_SZ = "sz"
)
const (
	SYS_ERR_Q_STOCK string = "获取股票信息异常"
	NA string = "N/A"
)

const (
	MATCH_LIMIT int = 1000
)

const (
	RANK_SIZE int = 20
	PAGE_SIZE int = 10
	MESSAGE_SEND_BUF = 500
)

const ONE_SECOND = 1 * time.Second + 10 * time.Millisecond
const ONE_HOUR = 1 * time.Hour + 10 * time.Millisecond

const (
	TASK_SUCCESS int8 = iota
	TASK_FAIL
)

const (
	TASK_DAILY = "daily" //一周一次，下同理
	TASK_WEEKLY = "weekly"
	TASK_MONTHLY = "monthly"
	TASK_YEARLY = "yearly"
	TASK_POLLING = "polling" //timer polling
)

// REDIS  KEYS
const (
	R_KEY_USERS = "users"
	R_KEY_STOCKS = "stocks"
	R_KEY_STOCK_CODES = "stock_codes"
	R_KEY_STOCKS_DETAIL = "stock_detail"
	R_KEY_MSGS = "latest_msgs"
	R_MSG_SEND_CHAN = "send_message_chan" //pub -sub message queue
)

//网站配置
const (
	ME_HOST = "localhost"
	ME_SCHEMA = "http"
	ME_NOTIFY_API = ME_HOST + "/user/msg/" //通知api 接口
)

//short url api
const (
	SHORT_API = "http://suo.im/api.php?url=%s"
)

//短信API
const (
	MSG_SEND_API = "http://www.6610086.net/jk.aspx"
	MSG_LEFT_API = "http://www.6610086.net/jkyy"
	MSG_STATUS_API = "http://www.6610086.net/jk_new_report.aspx"
	MSG_BIZ_CHAN = "45"        // 商业渠道
	MSG_SALE_CHAN = "52"        // 营销渠道
	MSG_ACCOUNT = "tangguowu" // account
	MSG_PASS = "syg123456" // pass
	MSG_TITLE = "【金修网络】"    // account
)
