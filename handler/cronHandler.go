package handler

import (
	"log"
	"math"
	"strings"
	"time"

	"../model"

	"fmt"

	"strconv"

	"github.com/go-xorm/xorm"
	"github.com/robfig/cron"
	"github.com/shopspring/decimal"
)

/**
任务实体
**/
type MyJob struct {
	Task   *model.Task
	Engine *xorm.Engine
	Log    *log.Logger
}

type keyType []float64
type kvType map[float64]string

//任务执行
func (job *MyJob) Run() {

	if ok := job.isRunabled(); ok { //
		job.Log.Printf("开始任务%s，上次执行时间：%s, 执行结果：%d",
			job.Task.Name, job.Task.LastRuntime.Format(model.DATE_TIME_FORMAT), job.Task.LastRunResult)
		InvokeObjectMethod(job, job.Task.Func)
		//更新任务数据库
		job.updateJobResult()
		job.Log.Printf("开始任务%s，上次执行时间：%s, 执行结果：%d",
			job.Task.Name, job.Task.LastRuntime.Format(model.DATE_FORMAT), job.Task.LastRunResult)
	} else {
		fmt.Printf("任务已经被停止运行...")
	}
}

/**
1、判断任务是有可用
2、判断任务是否已经执行
**/
func (job *MyJob) isRunabled() bool {

	task := new(model.Task)
	now := time.Now()
	if has, _ := job.Engine.Where("id=?", job.Task.Id).And("enable=?", 1).Get(task); has {

		//定时任务类型判断，并判断是否已经执行
		switch task.RunType {
		case model.TASK_DAILY:
			{
				today := now.Format(model.DATE_FORMAT)
				if task.LastRuntime.Format(model.DATE_FORMAT) == today {
					return false
				}
			}
		case model.TASK_WEEKLY:
			{
				n_year, n_week := now.ISOWeek()
				t_year, t_week := task.LastRuntime.ISOWeek()
				if n_year == t_year && n_week == t_week {
					return false
				}
			}
		case model.TASK_YEARLY:
			{

				if task.LastRuntime.Year() == now.Year() {
					return false
				}
			}
		case model.TASK_MONTHLY:
			{
				if task.LastRuntime.Format(model.DATE_MONTH) == now.Format(model.DATE_MONTH) {
					return false
				}
			}
		case model.TASK_POLLING:
			{

			}

		default:
			{

			}
		}

		job.Task.LastRuntime = task.LastRuntime
		job.Task.LastRunResult = task.LastRunResult
		job.Task = task
		return true
	}
	return false
}

func (job *MyJob) updateJobResult() {
	job.Task.LastRuntime = time.Now()
	job.Engine.Id(job.Task.Id).Update(job.Task)
}

/**
 获取数据库中的有效的task
 并且进行转换成job
**/
func ExtractEnabledTask2MyJob(c *cron.Cron, s *xorm.Engine, log *log.Logger) {

	tasks := make([]*model.Task, 0)
	if has, _ := s.Where("enable=?", 1).Get(&tasks); has {
		for _, task := range tasks {
			job := new(MyJob)
			job.Task = task
			job.Log = log
			job.Engine = s
			//添加到队列
			AddToRun(c, job, "schedule", task.Crontab)
		}
	}
}

/**
init
**/
func InitScheduleJobs(s *xorm.Engine, log *log.Logger) *cron.Cron {
	c := new(cron.Cron)
	ExtractEnabledTask2MyJob(c, s, log)
	return c
}

//运行调度器
func StartSchedule(c *cron.Cron) {
	c.Start()
	select {
	case <-time.After(1<<63 - 1):
		{
			return //从不停止
		}
	}

}

/**
cronType "func"|"job"|"schedule
**/
func AddToRun(c *cron.Cron, job *MyJob, cronType string, jobCron interface{}) {

	switch cronType {
	case "func":
		{
			c.AddFunc(jobCron.(string), func() {
				// return
			})
		}
	case "job":
		{
			// c.AddJob("@every 1s", job)
			c.AddJob(jobCron.(string), job)
		}
	case "schedule":
		{
			c.Schedule(cron.Every(jobCron.(time.Duration)), job)
		}
	default:
		{

		}
	}
}

/**************jobs ***************************************************************************
/**
收盘后，将今日买进的股票和之前的股票进行加权平均，并组合成一条记录
**/
func (job *MyJob) J_MixAndAvgHoldingStocks() {

	today := time.Now().Format(model.DATE_FORMAT)

	//1、判断是否已经执行，或者执行失败
	if (job.Task.LastRuntime.Format(model.DATE_FORMAT) == today && job.Task.LastRunResult == 1) || (job.Task.LastRuntime.Format(model.DATE_FORMAT) != today) { //今日已经执行
		if ok, err := doMixHoldingStocks(job.Engine, job.Log); ok {
			job.Task.LastRunResult = 0
		} else {
			job.Task.LastRunResult = 1
			job.Task.LastRunMsg = err.Error()
		}

		//更新执行任务结果
		job.Log.Println("任务之行结束：%s", job.Task)
		job.Engine.Update(job.Task)
	}
}

/**
收盘后，定时将所有今日委托进行取消
**/
func (job *MyJob) J_CancelAllEntrust() {
	CanCellAllEntrust(job.Engine.NewSession(), job.Log)
}

/**
抓取每天的股票动态数据
**/
func (job *MyJob) J_CurrentStockDetail() {

	job.Log.Println("获取最新的股票数据:%s", time.Now().Format(model.DATE_FORMAT))

	stocks := make([]*model.Stock, 0)
	err := job.Engine.Select("stock_code").Find(&stocks)

	if err == nil {
		len := len(stocks)
		if len > 0 {

			//一次去400个股票数据
			step := len / 400
			if len%400 != 0 {
				step++
			}
			for i := 0; i < step; i++ {
				//数据分片查询
				start := i * 400
				end := (i + 1) * 400
				if end > len {
					end = len
				}
				temp := stocks[start:end] //取 start~(end-1)
				//	启动协程，通过通道进行并行处理
				stringList := ConcatStockList(temp)
				ch := make(chan map[string]string)
				go GetStock5StagesWithChan(ch, stringList) //获取数据
				s := job.Engine.NewSession()
				//进行存储
				go func(*xorm.Session) {
					v5 := <-ch
					i := 0
					for k, v := range v5 {
						detail := new(model.CurrentStockDetail)
						v55 := strings.Split(v, ",")
						detail.StockCode = k
						detail.Detail = v55[3]
						f, _ := strconv.ParseFloat(v55[3], 64)
						detail.CurrentPrice = f
						i++
						if has, _ := s.Where("stock_code=?", k).Get(detail); has {
							s.Insert(detail)
						} else {
							s.Where("stock_code=?", k).Update(detail)
						}
					}
				}(s)
			}

		}
		job.Task.LastRunResult = 0
	} else {
		job.Task.LastRunResult = 1
		job.Task.LastRunMsg = err.Error()
	}
	job.Log.Println("获取最新的股票数据结束:%s", time.Now().Format(model.DATE_FORMAT))
}

/**
计算所有用户的今日股票市值
总盈亏
成功率
每只股票的浮动盈亏可以通过sql直接查询计算
**/
func (job *MyJob) J_DayEarningCalc() {

	//0、获取所有持仓
	earningSql := "update user_account ua LEFT JOIN " +
		"(SELECT ors.user_id, sum(round(ors.stock_number * 100 * (cs.current_price - holdings_price), 2)) earning, " +
		"sum(round(ors.stock_number * 100 * cs.current_price, 2))  value FROM (SELECT user_id, stock_code,  sum(stock_number) stock_number,  holdings_price   FROM stock_holding  WHERE holding_status = 1   GROUP BY user_id, stock_code, holdings_price) ors" +
		"LEFT JOIN  current_stock_detail cs  ON ors.stock_code = cs.stock_code" +
		"GROUP BY ors.user_id) v on ua.user_id = v.user_id " +
		"set ua.earning_rate = (v.value+ua.available_amount) /ua.init_amount, " + //总收益率
		"ua.stock_amount = v.value, " + //股票总市值
		"ua.total_amount = (v.value+ua.available_amount), " + //总资产 = (市值+可用金额)
		"ua.earning = (v.value+ua.available_amount - ua.init_amount)," + //总收益 = 总资产-初始化金额
		"ua.updated = now()"

	_, err := job.Engine.Exec(earningSql)
	if err == nil {
		job.Task.LastRunResult = model.TASK_SUCCESS
	} else {
		job.Task.LastRunResult = model.TASK_FAIL
		job.Task.LastRunMsg = err.Error()
	}
}

/***************************排名计算***********************************
/**
日排名计算
**/
func (job *MyJob) J_DailyRankCalc() {

	log.Println("日排名计算")
	userAccounts := make([]*model.UserAccount, 0)
	dayRanks := make([]*model.DayRank, 20)
	session := job.Engine.NewSession()
	session.Begin()
	defer session.Close()

	//持仓中最大收益率股票
	earningSql := "SELECT ors.user_id, ors.stock_code,s.stock_name, (cs.current_price-ors.trans_price ) / ors.trans_price earning_rate" +
		",date_format(curdate(),'%Y-%m-%d') day " +
		"FROM  (SELECT user_id, stock_code, sum(stock_number) stock_number, trans_price FROM stock_holding " +
		"WHERE holding_status = 1 AND date_format(trans_time,'%Y-%m-%d')= date_format(curdate() ,'%Y-%m-%d') " +
		" GROUP BY user_id, stock_code, trans_price ) ors " +
		"LEFT JOIN  current_stock_detail cs  ON ors.stock_code = cs.stock_code" +
		"LEFT JOIN  stock s ON ors.stock_code = s.stock_code" +
		"GROUP BY  ors.user_id,ors.stock_code" +
		"ORDER BY  earning_rate DESC " +
		"LIMIT  ?"

	err := session.Desc("earning", "earning_rate").Limit(model.RANK_SIZE, 0).Find(&userAccounts)
	if err == nil {
		i := 0
		for _, userAccount := range userAccounts {
			dayRank := new(model.DayRank)
			has, err := session.Sql(earningSql, userAccount.UserId, model.RANK_SIZE).Get(&dayRank)
			if err == nil {
				if has {
					dayRank.TotalFollow = countFollowNumbers(session, dayRank.UserId)
					dayRanks[i] = dayRank
					i++
				}
			} else {
				job.Log.Println("日排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
				break
			}
		}
		if len(dayRanks) > 0 {
			_, err := session.InsertMulti(dayRanks)
			if err == nil {
				job.Task.LastRunResult = model.TASK_SUCCESS
				session.Commit()
			} else {
				job.Log.Println("日排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
			}
		}

	} else {
		job.Log.Printf("查询今日买入股票列表异常\n")
		return
	}
}

/**
将计算的结果排序后插入到数据库
**/
func dumpDailyRanks(x *xorm.Session, keys keyType, kv kvType) {

	today := time.Now().Format(model.DATE_FORMAT)
	log.Printf("开始计算今日排名：%s", today)
	x.Begin()
	p := 0
	ranks := make([]*model.DayRank, model.RANK_SIZE)
	for i, key := range keys {
		if i < 10 {
			infos := kv[key] //1-600001,1-60002,3-600001
			info := strings.Split(infos, ",")
			for _, in := range info {
				if p < model.RANK_SIZE {
					coodes := strings.Split(in, "-")
					dayRank := new(model.DayRank)
					if id, err := strconv.Atoi(coodes[0]); err != nil {
						dayRank.UserId = int64(id)
						dayRank.EarningRate = math.Abs(key)
						dayRank.Day = today
						dayRank.StockCode = coodes[1]
						ranks[p] = dayRank
						p++
					}

				} else {
					break
				}
			}
		} else {
			break
		}
	}
	_, err := x.InsertMulti(ranks)
	if err == nil {
		x.Commit()
		log.Printf("今日排名计算结束：%s", today)
	} else {
		x.Rollback()
		log.Printf("今日排名计算失败：%s", today)
	}
}

/**
周排名计算
**/
func (job *MyJob) J_WeeklyRankCalc() {

	log.Println("周排名计算")
	userAccounts := make([]*model.UserAccount, 0)
	dailyRanks := make([]*model.WeekRank, 20)
	session := job.Engine.NewSession()
	session.Begin()
	defer session.Close()

	//持仓中最大收益率股票
	earningSql := "SELECT  ors.user_id, ors.stock_code,s.stock_name, (cs.current_price-ors.trans_price ) / ors.trans_price earning_rate" +
		",yearweek(curdate()) week, date_format(curdate(),'%Y-%m') month " +
		"FROM  (SELECT user_id, stock_code, sum(stock_number) stock_number, trans_price FROM stock_holding " +
		"WHERE holding_status = 1 AND user_id = ? GROUP BY user_id, stock_code, trans_price  ) ors " +
		"LEFT JOIN  current_stock_detail cs  ON ors.stock_code = cs.stock_code" +
		"LEFT JOIN  stock s  ON ors.stock_code = s.stock_code" +
		"GROUP BY  ors.user_id,ors.stock_code" +
		"ORDER BY  earning_rate DESC " +
		"LIMIT  1"

	err := session.Desc("earning", "earning_rate").Limit(model.RANK_SIZE, 0).Find(&userAccounts)
	if err == nil {
		i := 0
		for _, userAccount := range userAccounts {
			weekRank := new(model.WeekRank)
			has, err := session.Sql(earningSql, userAccount.UserId).Get(&weekRank)
			if err == nil {
				if has {
					weekRank.TotalFollow = countFollowNumbers(session, weekRank.UserId)
					dailyRanks[i] = weekRank
					i++
				}
			} else {
				job.Log.Println("周排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
				break
			}
		}
		if len(dailyRanks) > 0 {
			_, err := session.InsertMulti(dailyRanks)
			if err == nil {
				job.Task.LastRunResult = model.TASK_SUCCESS
				session.Commit()
			} else {
				job.Log.Println("周排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
			}
		}

	} else {
		job.Log.Printf("查询今日买入股票列表异常\n")
		return
	}
}

/**
月排名计算
**/
func (job *MyJob) J_MonthlyRankCalc() {

	log.Println("月排名计算")
	userAccounts := make([]*model.UserAccount, 0)
	monthRanks := make([]*model.MonthRank, 20)
	session := job.Engine.NewSession()
	session.Begin()
	defer session.Close()

	//持仓中最大收益率股票
	earningSql := "SELECT  ors.user_id, ors.stock_code,s.stock_name, (cs.current_price-ors.trans_price ) / ors.trans_price earning_rate" +
		" ,date_format(curdate(),'%Y-%m') month " +
		"FROM  (SELECT user_id, stock_code, sum(stock_number) stock_number, trans_price  FROM stock_holding " +
		"WHERE holding_status = 1 AND user_id = ? GROUP BY user_id, stock_code, trans_price  ) ors " +
		"LEFT JOIN  current_stock_detail cs  ON ors.stock_code = cs.stock_code" +
		"LEFT JOIN  stock s  ON ors.stock_code = s.stock_code" +
		"GROUP BY  ors.user_id,ors.stock_code" +
		"ORDER BY  earning_rate DESC" +
		"LIMIT  1"

	err := session.Desc("earning", "earning_rate").Limit(model.RANK_SIZE, 0).Find(&userAccounts)
	if err == nil {
		i := 0
		for _, userAccount := range userAccounts {

			monthRank := new(model.MonthRank)
			has, err := session.Sql(earningSql, userAccount.UserId).Get(&monthRank)
			if err == nil {
				if has {
					monthRank.TotalFollow = countFollowNumbers(session, monthRank.UserId)
					monthRanks[i] = monthRank
					i++
				}
			} else {
				job.Log.Println("月排名计算异常：%s\n", err)

				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
				break
			}

		}
		if len(monthRanks) > 0 {
			_, err := session.InsertMulti(monthRanks)
			if err == nil {
				job.Task.LastRunResult = model.TASK_SUCCESS
				session.Commit()
			} else {
				job.Log.Println("月排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
			}
		}

	} else {
		job.Log.Printf("查询今日买入股票列表异常\n")
		return
	}
}

/**
如果自己有订阅用户，则自己交易完毕后，短信通知客户
1秒轮询机制

**/
func (job *MyJob) NotifyFollowersAfterTrx() {

	//0、开启事务
	session := job.Engine.NewSession()
	session.Begin()
	defer session.Close()
	//1、查询所有没有通知过的交易
	trxs := make([]*model.StockTrans, 0)
	err := session.Where("notify_status=?", 1).Find(&trxs)
	if err == nil {
		//遍历所有交易
		if len(trxs) > 0 {
			for _, trx := range trxs {
				if has, myFollowers := listMyFollowers(session, job.Log, trx.UserId); has {
					//保存通知

					if ok := dumpNotifyLogs(session, myFollowers, trx); ok {
						trx.NotifyStatus = 2
					} else {
						trx.NotifyStatus = 3
					}

				} else {
					job.Log.Printf("没有待通知的交易")
				}

			}
		}
	} else {
		job.Task.LastRunMsg = err.Error()
		job.Task.LastRunResult = model.TASK_FAIL
	}

}

/**
dump notify log
**/
func dumpNotifyLogs(s *xorm.Session, myFollowers []*model.UserFollow, trx *model.StockTrans) bool {

	nofifyLogs := make([]*model.NotifyLog, len(myFollowers))
	for i, follower := range myFollowers {
		nofifyLog := new(model.NotifyLog)
		nofifyLog.TrxId = trx.Id
		nofifyLog.FollowedId = trx.UserId        //被订阅人ID
		nofifyLog.FollowerId = follower.UserId   //订阅人ID
		nofifyLog.Content = model.TRX_NOTIFY_MSG //
		// "尊敬的客户 %s：您好，您订阅的用户%s于%s %s了股票%s %d 手 ,如有问题，请联系我们电话：%s"
		// GetRedisUser()
		fmt.Sprintf(model.TRX_NOTIFY_MSG)
		nofifyLogs[i] = nofifyLog
	}
	return true
}

/**
股票加权平均
**/
func doMixHoldingStocks(x *xorm.Engine, log *log.Logger) (bool, error) {

	//0、开启事务
	session := x.NewSession()
	session.Begin()
	defer session.Close()

	//today
	today := time.Now().Format(model.DATE_FORMAT)
	todayStocks := make([]*model.StockHolding, 0)
	err := session.Sql("select * from stock_holding where holding_status = 1 "+
		"and date_format(trans_time,'%Y-%m-%d') = ?", today).Find(&todayStocks)

	if err == nil {
		oldStocks := make([]*model.StockHolding, 0, len(todayStocks)) //之前的股票
		for _, stock := range todayStocks {
			oldStock := new(model.StockHolding)
			if has, _ := session.Where("user_id=?", stock.UserId).
				And("stock_code=?", stock.StockCode).
				And("trans_time<?", time.Now().Format(model.DATE_FORMAT_1)).
				And("holding_status=?", 1).Get(&oldStock); has {
				//求平均
				num, avgP := func(*model.StockHolding, *model.StockHolding) (int32, float64) {
					num := stock.StockNumber + oldStock.StockNumber
					pp := decimal.NewFromFloat(stock.TransPrice).Mul(decimal.New(int64(stock.StockNumber), 64)).
						Add(decimal.NewFromFloat(oldStock.TransPrice).
							Mul(decimal.New(int64(oldStock.StockNumber), 64)))
					f, _ := pp.DivRound(decimal.New(int64(num), 64), 2).Float64()
					return num, f
				}(stock, oldStock)
				//更新oldStock
				oldStock.StockNumber = num
				oldStock.TransPrice = avgP
				// _, err := session.Update(oldStock)
				// oldStocks[j] = oldStock
				//追加
				oldStocks = append(oldStocks, oldStock)
			}
		}
		if len(oldStocks) > 0 {
			_, err := session.InsertMulti(oldStocks)
			if err == nil {
				session.Commit()
			} else {
				session.Rollback()
			}
		}
		return true, nil
	} else {
		log.Printf("db err:%s", err.Error())
		session.Rollback()
		return false, err
	}

}
