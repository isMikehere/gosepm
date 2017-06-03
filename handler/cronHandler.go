package handler

import (
	"log"
	"math"
	"strings"
	"time"

	"../model"

	"fmt"

	"strconv"

	redis "github.com/go-redis/redis"
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
	R      *redis.Client
}

type keyType []float64
type kvType map[float64]string

//任务执行
func (job *MyJob) Run() {

	if ok := job.IsRunabled(); ok { //
		log.Printf("开始任务%s，上次执行时间：%s, 执行结果：%d",
			job.Task.Name, job.Task.LastRuntime.Format(model.DATE_TIME_FORMAT), job.Task.LastRunResult)
		InvokeObjectMethod(job, job.Task.Func)
		//更新任务数据库
		job.UpdateJobResult()
		log.Printf("结束任务%s，结束运行时间：%s, 执行结果：%d",
			job.Task.Name, job.Task.LastRuntime.Format(model.DATE_TIME_FORMAT), job.Task.LastRunResult)
	} else {
		fmt.Printf("任务已经被停止运行...")
	}
}

/**
1、判断任务是有可用
2、判断任务是否已经执行
**/
func (job *MyJob) IsRunabled() bool {

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

/**
更新任务
**/
func (job *MyJob) UpdateJobResult() {
	log.Printf("更新任务：%s,%s", job.Task.Name, time.Now().Format(model.DATE_TIME_FORMAT))
	job.Task.LastRuntime = time.Now()
	job.Engine.Id(job.Task.Id).Update(job.Task)
}

/**
 获取数据库中的有效的task
 并且进行转换成job
**/
func ExtractEnabledTask2MyJob(c *cron.Cron, s *xorm.Engine, r *redis.Client) {
	tasks := make([]*model.Task, 0)
	if err := s.Where("enable=?", 1).Find(&tasks); err == nil {
		for _, task := range tasks {
			job := new(MyJob)
			job.Task = task
			job.Engine = s
			job.R = r
			//添加到队列
			AddToRun(c, job, "job", task.Crontab)
		}
	}
}

/**
init
**/
func InitScheduleJobs(s *xorm.Engine, r *redis.Client) *cron.Cron {
	c := cron.NewWithLocation(time.Now().Location())
	ExtractEnabledTask2MyJob(c, s, r)
	return c
}

//运行调度器
func StartSchedule(c *cron.Cron) {
	fmt.Printf("start...")
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
			// c.AddJob("@every 10s", job)
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
		if ok, err := doMixHoldingStocks(job.Engine); ok {
			job.Task.LastRunResult = 0
		} else {
			job.Task.LastRunResult = 1
			job.Task.LastRunMsg = err.Error()
		}

		//更新执行任务结果
		log.Println("任务之行结束：%s", job.Task)
		job.Engine.Update(job.Task)
	}
}

/**
收盘后，定时将所有今日委托进行取消
**/
func (job *MyJob) J_CancelAllEntrust() {
	CanCellAllEntrust(job.Engine.NewSession())
}

/**
抓取每天的股票动态数据
**/
func (job *MyJob) J_CurrentStockDetail() {

	log.Printf("获取最新的股票数据:%s", time.Now().Format(model.DATE_TIME_FORMAT))

	// 	stocks := make([]interface{}, 0)
	// if stocks == nil {
	// 		err := job.Engine.
	// 			Sql("select concat(s.location,s.stock_code) stock_code from stock s").Find(&stocks)
	// 	}

	stocks := GetRedisStockCodes(job.R)
	if stocks != nil {

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
				ch := make(chan map[string]interface{})

				//获取数据
				go GetStock5StagesWithChan(ch, stringList)
				//进行存储
				go DumpStockDetail(job, ch)

				time.Sleep(2e9)
			}
		}
		job.Task.LastRunResult = 0
	} else {
		job.Task.LastRunResult = 0
		job.Task.LastRunMsg = "没有数据"
		log.Printf("执行失败%s,%s", job.Task.Name, job.Task.LastRunMsg)
	}
	log.Printf("获取最新的股票数据结束:%s", time.Now().Format(model.DATE_TIME_FORMAT))
}

/**
保存当前股票详情数据库或者缓存
**/
func DumpStockDetail(job *MyJob, ch chan map[string]interface{}) {

	if job.Task.DumpType == 0 { //落地数据库
		v5 := <-ch
		for k, v := range v5 {

			detail := new(model.CurrentStockDetail)

			v55 := make([]string, 0)
			if !strings.Contains(v.(string), ",") {
				continue
			}

			v55 = strings.Split(v.(string), ",")

			if has, _ := job.Engine.Where("stock_code=?", k).Get(detail); has {
				f, _ := strconv.ParseFloat(v55[3], 64)
				detail.CurrentPrice = f
				v55[0] = detail.StockName
				detail.Detail = strings.Join(v55, ",")
			} else {
				detail.StockCode = k
				if d := GetRedisStock(job.R, k); d != nil {
					detail.StockName = d.StockName
					v55[0] = d.StockName
					detail.Detail = strings.Join(v55, ",")
				}
				f, _ := strconv.ParseFloat(v55[3], 64)
				detail.CurrentPrice = f
				detail.Detail = strings.Join(v55, ",")
			}

			if detail.Id <= 0 {
				job.Engine.Insert(detail)
			} else {
				job.Engine.Id(detail.Id).Update(detail)
			}

		}
	} else if job.Task.DumpType == 1 { //redis
		v5 := <-ch
		v5 = func(v5 map[string]interface{}) map[string]interface{} {
			for k, v := range v5 {
				v55 := make([]string, 0)
				if !strings.Contains(v.(string), ",") {
					continue
				}
				v55 = strings.Split(v.(string), ",")
				if d := GetRedisStock(job.R, k); d != nil {
					v55[0] = d.StockName
					x := strings.Join(v55, ",")
					v5[k] = x
				}
			}
			return v5
		}(v5)

		_, err := job.R.HMSet(model.R_KEY_STOCKS_DETAIL, v5).Result()

		if err != nil {
			job.Task.LastRunResult = 1
			job.Task.LastRunMsg = err.Error()
			fmt.Printf("更新失败：%s\n", err.Error())
		}
	}
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
	earningSql := "SELECT t.*,(@i:=@i+1) AS rank from ( " +
		"SELECT ors.user_id, ors.stock_code,u.nick_name,s.stock_name, (cs.current_price-ors.trans_price ) / ors.trans_price earning_rate" +
		",date_format(curdate(),'%Y-%m-%d') day " +
		"FROM  (SELECT user_id, stock_code, sum(stock_number) stock_number, trans_price FROM stock_holding " +
		"WHERE holding_status = 1 AND date_format(trans_time,'%Y-%m-%d')= date_format(curdate() ,'%Y-%m-%d') " +
		" GROUP BY user_id, stock_code, trans_price ) ors " +
		"LEFT JOIN  current_stock_detail cs  ON ors.stock_code = cs.stock_code " +
		"LEFT JOIN  stock s ON ors.stock_code = s.stock_code " +
		"LEFT JOIN  user u ON u.id  = ors.user_id " +
		"GROUP BY  ors.user_id,ors.stock_code " +
		"ORDER BY  earning_rate DESC " +
		"LIMIT  ?) t,(SELECT @i:=0) AS it"

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
				log.Println("日排名计算异常：%s\n", err)
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
				log.Println("日排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
			}
		}

	} else {
		log.Printf("查询今日买入股票列表异常\n")
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
	earningSql := "SELECT  ors.user_id, ors.stock_code,u.nick_name,s.stock_name, (cs.current_price-ors.trans_price ) / ors.trans_price earning_rate" +
		",yearweek(curdate()) week, date_format(curdate(),'%Y-%m') month " +
		"FROM  (SELECT user_id, stock_code, sum(stock_number) stock_number, trans_price FROM stock_holding " +
		"WHERE holding_status = 1 AND user_id = ? GROUP BY user_id, stock_code, trans_price  ) ors " +
		"LEFT JOIN  current_stock_detail cs  ON ors.stock_code = cs.stock_code " +
		"LEFT JOIN  stock s  ON ors.stock_code = s.stock_code " +
		"LEFT JOIN  user u ON u.id  = ors.user_id " +
		"GROUP BY  ors.user_id,ors.stock_code " +
		"ORDER BY  earning_rate DESC " +
		"LIMIT  1"

	err := session.Desc("earning", "earning_rate").Limit(model.RANK_SIZE, 0).Find(&userAccounts)
	if err == nil {
		for i, userAccount := range userAccounts {
			weekRank := new(model.WeekRank)
			has, err := session.Sql(earningSql, userAccount.UserId).Get(&weekRank)
			if err == nil {
				if has {
					weekRank.Rank = i + 1
					weekRank.TotalFollow = countFollowNumbers(session, weekRank.UserId)
					dailyRanks[i] = weekRank
				}
			} else {
				log.Printf("周排名计算异常：%s\n", err)
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
				log.Printf("周排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
			}
		}

	} else {
		log.Printf("查询今日买入股票列表异常\n")
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
	earningSQL := "SELECT  ors.user_id, ors.stock_code,u.nick_name,s.stock_name, (cs.current_price-ors.trans_price ) / ors.trans_price earning_rate" +
		" ,date_format(curdate(),'%Y-%m') month " +
		"FROM  (SELECT user_id, stock_code, sum(stock_number) stock_number, trans_price  FROM stock_holding " +
		"WHERE holding_status = 1 AND user_id = ? GROUP BY user_id, stock_code, trans_price  ) ors " +
		"LEFT JOIN  current_stock_detail cs  ON ors.stock_code = cs.stock_code " +
		"LEFT JOIN  stock s  ON ors.stock_code = s.stock_code " +
		"LEFT JOIN  user u ON u.id  = ors.user_id " +
		"GROUP BY  ors.user_id,ors.stock_code " +
		"ORDER BY  earning_rate DESC " +
		"LIMIT  1"

	err := session.Desc("earning", "earning_rate").Limit(model.RANK_SIZE, 0).Find(&userAccounts)
	if err == nil {
		for i, userAccount := range userAccounts {

			monthRank := new(model.MonthRank)
			has, err := session.Sql(earningSQL, userAccount.UserId).Get(&monthRank)
			if err == nil {
				if has {
					monthRank.Rank = i + 1
					monthRank.TotalFollow = countFollowNumbers(session, monthRank.UserId)
					monthRanks[i] = monthRank
				}
			} else {
				log.Printf("月排名计算异常：%s\n", err)

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
				log.Printf("月排名计算异常：%s\n", err)
				job.Task.LastRunResult = model.TASK_FAIL
				job.Task.LastRunMsg = err.Error()
				session.Rollback()
			}
		}

	} else {
		log.Printf("查询今日买入股票列表异常\n")
		return
	}
}

type EarningStruct struct {
	SuccessTimes   int32
	TotalTimes     int32
	SuccessRate    float64
	TransFrequency float64
}

/**
总收益排名
一天执行一次，收盘结束执行
**/
func (job *MyJob) J_TotalEarningCalc() {

	userAccounts := make([]*model.UserAccount, 0)
	session := job.Engine.NewSession()
	session.Begin()
	defer session.Close()
	err := session.Desc("earning").Find(&userAccounts)

	if err == nil {

		//盈亏Sql
		earningSQL := "SELECT  t.*,round(t.success_times/t.total_times,2) success_rate, " +
			" round(datediff(t.max_time,t.min_time) / t.total_times,2) trans_frequency from " +
			" (SELECT count(1) total_times,sum(case when i.earning>=0 then 1 else 0 end) as success_times, " +
			" max(i.buy_time) max_time,min(i.buy_time) min_time" +
			" from stock_trx_info i where user_id = ? and trans_status= 1) as t"

			//计算排名
		for i, userAccount := range userAccounts {
			//排名
			userAccount.Rank = i + 1
			//计算盈亏笔数，成功率，交易频率
			earning := new(EarningStruct)
			if has, _ := session.Sql(earningSQL, userAccount.UserId).Get(earning); has {
				userAccount.SuccessTimes = earning.SuccessTimes
				userAccount.TotalTimes = earning.TotalTimes
				userAccount.SuccessRate = earning.SuccessRate
				userAccount.TransFrequency = earning.TransFrequency
			}
			_, err := session.Id(userAccount.Id).Update(userAccount)

			if err != nil {
				session.Rollback()
				log.Printf("[%s]执行更新出现异常 %s", job.Task.Name, err.Error())
				job.Task.LastRunResult = 1
				job.Task.LastRunMsg = "执行更新出现异常 "
				break
			}
		}
		session.Commit()
	} else {
		log.Printf("查询今日买入股票列表异常\n")
		return
	}
}

/**
如果自己有订阅用户，则自己交易完毕后，短信通知客户
5秒轮询机制
**/
func (job *MyJob) J_NotifyFollowersAfterTrx() {

	//0、开启事务
	session := job.Engine.NewSession()
	session.Begin()
	defer session.Close()
	//1、查询所有没有通知过的交易
	//买进的交易
	//0:sale ,1:buy

	notifyStatuss := []int{0, 2}
	for _, s := range notifyStatuss {
		trxInfos := make([]*model.StockTrxInfo, 0)
		err := session.Where("notify_status=?", 1).And("notify_status = ?", s).Find(&trxInfos)
		if err == nil {
			//遍历所有交易
			if len(trxInfos) > 0 {
				for _, trxInfo := range trxInfos {
					//查询待通知的follow列表
					notifyFollows := make([]*model.NotifyFollow, 0)
					followedUser := GetUserById(s, job.R, trxInfo.UserId) //被订阅者
					//通知列表
					err := session.Where("trx_info_id =?", trxInfo.Id).Find(&notifyFollows)
					if err != nil && len(notifyFollows) > 0 && followedUser != nil {
						//订阅用户列表
						if has, myFollowers := listMyFollowers(session, trxInfo.UserId); has {
							//查找交易要通知的具体交易
							trx := new(model.StockTrans)
							if has, _ := session.Where("id=?", trxInfo.LastTrxId).And("notify_status=?", 1).Get(trx); has {
								// //保存通知
								if dumpOk := dumpNotifyLogs(session, myFollowers, followedUser, trx, job.R); dumpOk {
									trx.NotifyStatus = 2
									trxInfo.NotifyStatus = int8(s + 1)
								} else {
									job.Task.LastRunResult = 1
									job.Task.LastRunMsg = "dump msg执行失败"
									fmt.Print("dump msg执行失败....")
									break
								}
							} else {
								trxInfo.NotifyStatus = int8(s + 1)
							}
						} else {
							log.Printf("没有待通知的交易")
						}
					} else {
						trxInfo.NotifyStatus = 3
					}
					//更新交易对
					session.Id(trxInfo.Id).Update(trxInfo)
				}
			}
		} else {
			job.Task.LastRunMsg = err.Error()
			job.Task.LastRunResult = model.TASK_FAIL
			break
		}
	}
}

/**
dump notify log
**/
func dumpNotifyLogs(s *xorm.Session, myFollowers []*model.UserFollow, followedUesr *model.User, trx *model.StockTrans, redis *redis.Client) bool {

	//批次号
	//根据数量进行切片.500一次
	len := len(myFollowers)
	if len > 0 {
		in_batch_id := time.Now().Format(model.DATE_ORDER_FORMAT)
		messageLogs := make([]*model.MessageLog, len)

		//一次500个短信
		step := len / 500
		if len%500 != 0 {
			step++
		}
		for i := 0; i < step; i++ {
			//数据分片查询
			start := i * 500
			end := (i + 1) * 500
			if end > len {
				end = len
			}

			temp := myFollowers[start:end] //取 start~(end-1)
			for _, follower := range temp {
				var u *model.User
				if u = GetRedisUser(redis, strconv.Itoa(int(follower.UserId))); u == nil {
					if has, _ := s.Id(follower.UserId).Get(u); !has {
						continue
					}
				}
				messageLog := new(model.MessageLog)
				messageLog.Mobile = u.Mobile
				messageLog.SendStatus = 0
				messageLog.InBatchId = in_batch_id
				msgKey := strconv.Itoa(int(follower.UserId)) + strconv.Itoa(int(trx.Id))
				// "【金修网络】尊敬的用户:您在本平台中订阅的用户%s有了新交易动态，查看请点击链接 %s"
				messageLog.Content = fmt.Sprintf(model.TRX_NOTIFY_MSG, u)
				// 【金修网络】尊敬的用户:您在模拟股票交易中订阅的用户%s在%s买入%s，成交价格%s %s股。以上数据来自于模拟股票交易仅供参考
				num := trx.TransNumber * 100
				messageLog.Detail = fmt.Sprintf(model.TRX_NOTIFY_DETAIL,
					u.NickName, trx.TransTime.Format(model.DATE_TIME_FORMAT),
					trx.StockCode, trx.TransPrice, strconv.Itoa(int(num)))
				redis.Set(msgKey, messageLog.Detail, time.Hour*24*2) //设置消息redis
				messageLogs = append(messageLogs, messageLog)
			}

		}
	}
	return true
}

/**
发送短信任务
15秒轮训一次
**/
func (job *MyJob) J_SendMsg() {

	batchs := make([]*model.MessageLog, 0)
	s := job.Engine.NewSession()
	s.Begin()
	defer s.Close()
	//根据批次号进行发送
	if err := s.NoCache().Where("send_status=?", 0).Distinct("in_batch_id").GroupBy("in_batch_id").Find(&batchs); err == nil {
		if len(batchs) > 0 {
			for _, inBatchId := range batchs {
				messages := make([]*model.MessageLog, 0)
				s.Cols("id,mobile,content").Where("in_batch_id=?", inBatchId).And("send_status=?", 0).Find(&messages)
				if len(messages) > 0 {
					content := ""
					mobiles := make([]string, len(messages))
					for i, m := range messages {
						if content == "" {
							content = m.Content
						}
						mobiles[i] = m.Mobile
					}
					//join mobile
					var err error
					mobileList := strings.Join(mobiles, ",")
					if flag, ret := sendMessage(mobileList, content); flag && strings.Contains(ret, "0:") {
						_, err = s.Exec("update message_log set send_status =?,ret_batch_id = ?,updated=now() "+
							" where in_batch_id=?", 1, ret, inBatchId)
					} else {
						_, err = s.Exec("update message_log set send_status =?,err_msg =?,updated=now()"+
							" where in_batch_id=?", 2, ret, inBatchId)
					}
					if err != nil {
						log.Printf("短信发送失败%s", inBatchId)
						s.Rollback()
					} else {
						s.Commit()
					}
				}
			}
		}
	} else {
		log.Printf("查询待发送短信异常:%s", err.Error())
	}
}

/*
获取短信状态
**/
func (job *MyJob) J_GetMsgStatus() {

}

/**
股票加权平均
**/
func doMixHoldingStocks(x *xorm.Engine) (bool, error) {

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
					pp := decimal.NewFromFloat(stock.TransPrice).Mul(decimal.New(int64(stock.StockNumber), 0)).
						Add(decimal.NewFromFloat(oldStock.TransPrice).
							Mul(decimal.New(int64(oldStock.StockNumber), 0)))
					f, _ := pp.DivRound(decimal.New(int64(num), 0), 0).Float64()
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
