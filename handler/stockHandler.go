package handler

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"../model"
	"github.com/go-macaron/session"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"github.com/shopspring/decimal"
	macaron "gopkg.in/macaron.v1"
)

/**
跳转交易页面
**/
func TrxGetHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {
	//获取当前用户
	if ok, user := GetSessionUser(sess); ok {
		//当前用户
		ctx.Data["i"] = user
		//获取用户的资金账户信息
		if has, userAccount := QueryUserAccoutByUserIdWithEngine(x, user.Id); has {
			ctx.Data["ua"] = userAccount
		}
		// 账户信息
		ctx.HTML(200, "trx")
	}
}

/**
用户持仓handler
TODO:需要更新
**/
//FIXME:增加持仓数据
func UserHoldingHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {
	//获取当前用户
	if ok, user := GetSessionUser(sess); ok {
		//当前用户
		ctx.Data["i"] = user
		//获取用户的资金账户信息
		if has, userAccount := QueryUserAccoutByUserIdWithEngine(x, user.Id); has {
			ctx.Data["ua"] = userAccount
		}
		// 账户信息
		ctx.HTML(200, "trx")
	}
}

/**
委托一个股票,
委托成功后，如果判断可以交易，则直接交易，如果不可以择继续委托状态
**/
func TrxEntrustPostHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {
	//根据股票的代码，数量，价格，进行委托---》委托前检查
	//获取参数

	jr := new(model.JsonResult)
	quantity := ctx.Params("quantity")                //数量 单位：手
	buyPrice := ctx.Params("buyPrice")                //价格
	trxType, _ := strconv.Atoi(ctx.Params("trxType")) //交易类型 0:卖，1:买
	stockCode := ctx.Params("stockCode")              //股票代码

	if stoop := CheckStockStopService(x, log, stockCode); stoop {
		jr.Code = "-100"
		jr.Msg = model.TRX_STOCK_NOT_SALE
		ctx.JSON(200, jr)
		return
	}

	//当前用户
	_, user := GetSessionUser(sess)
	ctx.Data["i"] = user
	//获取用户的资金账户信息
	if has, userAccount := QueryUserAccoutByUserIdWithEngine(x, user.Id); has {
		// can [1,2] -->[可以委托，不可以委托]
		if can, r := canEntrust(x, userAccount, stockCode, quantity, buyPrice, trxType); can == 1 {
			//委托成功
			code, msg := doEntrust(x, user.Id, stockCode, r["rbq"].(decimal.Decimal), r["rbp"].(decimal.Decimal), int8(trxType), int8(can))
			jr.Code = code
			jr.Msg = msg
			ctx.JSON(200, jr)
			return

		} else {
			jr.Msg = r["msg"].(string)
			jr.Code = "100"
			ctx.JSON(200, jr)
			return
		}
	}
}

/**
股票委托检查trxType 0:卖，1:买
ret （0:可以立即交易 ，1:可以委托，2:委托失败)
**/
func canEntrust(x *xorm.Engine, ua *model.UserAccount, stockCode string, quantity string, buyPrice string, trxType int) (int, map[string]interface{}) {

	//1、股票自身情况，停盘、涨停、跌停
	//2、购买用户的资金情况前端做第一次判断，提交后则后端在此校验
	ret := make(map[string]interface{}, 0)   //
	bq, _ := decimal.NewFromString(quantity) // 购买量
	bp, _ := decimal.NewFromString(buyPrice) // 购买价格
	//获取当前股票的买卖五档
	ok, infoArr := GetStock5Stages(AddExcToStockCode(stockCode))
	v5 := infoArr[stockCode]

	//1、判断当前股价是否已经
	cp, _ := strconv.ParseFloat(v5[3], 2)  //当前价
	yCp, _ := strconv.ParseFloat(v5[2], 2) //昨日收盘价
	dm, _ := decimal.NewFromString(v5[8])
	dealAmount := dm.Div(decimal.New(int64(100), 0))           //交易量(单位：手)
	var rBp, rBq, upLimitPrice, downLimitPrice decimal.Decimal //实际交易价格,数量,涨停价，跌停价

	var percent float64 = 0.1
	if IsST(v5[0]) {
		percent = 0.05
	}
	upLimitPrice = GetLimitPrice(yCp, percent, 1)    //今日涨停价
	downLimitPrice = GetLimitPrice(yCp, percent, -1) //今日跌停价

	if ok { //获取买卖五档成功
		if trxType == 1 { // buy

			//1、涨停价和当前价格进行比较，
			if decimal.NewFromFloat(cp).Cmp(upLimitPrice) >= 0 {
				ret["msg"] = model.TRX_STOCK_INC_TOP
				return 2, ret
			}
			//2、价格是大于涨停价格
			if bp.Cmp(upLimitPrice) > 0 {
				ret["msg"] = model.TRX_STOCK_X_PRICE
				return 2, ret
			}

			//3、如果交易量为0 或者不足的话，则按照交易量进行部分交易

			if dealAmount.Cmp(decimal.Zero) == 0 {
				ret["msg"] = model.TRX_STOCK_ENT_FAIL
				return 2, ret
			} else {
				rBq = decimal.Min(dealAmount, bq)
			}
			//4、判断价格
			//如果最新成交价等于委托价，按照委托价成交，如果最新价小于委托价，按照最新价撮合成交，涨停不能买入
			if decimal.NewFromFloat(cp).Cmp(bp) <= 0 {
				rBp = decimal.NewFromFloat(cp)
			} else {
				rBp = bp
			}

			//5、判断是否有足够资金购买
			cost := rBq.Mul(rBp).Mul(decimal.New(100, 0)).Round(2)
			if decimal.NewFromFloat(ua.AvailableAmount).Cmp(cost) < 0 {
				ret["msg"] = model.TRX_SHORT_OF_MONEY
				return 2, ret
			} else {
				//TODO:这里暂时停止 委托立即交易的模式
				// ret["msg"] = model.TRX_STOCK_ENT_OK
				// ret["rBq"] = rBq
				// ret["rBp"] = rBp
				// return 0, ret
			}
			//将实际的委托价格、数量、消息返回
			ret["msg"] = model.TRX_STOCK_ENT_OK
			return 1, ret

		} else { //sale
			//1、当前价跟跌停价比较，判断是否涨停
			if decimal.NewFromFloat(cp).Cmp(downLimitPrice) <= 0 {
				ret["msg"] = model.TRX_STOCK_DEC_TOP
				return 2, ret
			}
			//2、买入价格是小于涨停价格
			if bp.Cmp(downLimitPrice) < 0 {
				ret["msg"] = model.TRX_STOCK_O_PRICE
				return 2, ret
			}
			//3、如果交易量为0 或者不足的话，则按照交易量进行部分交易
			if dealAmount.Cmp(decimal.Zero) == 0 {
				ret["msg"] = model.TRX_STOCK_ENT_FAIL
				return 2, ret
			} else {
				rBq = decimal.Min(dealAmount, bq)
			}
			//4、判断价格
			//如果最新成交价等于委托价，按照委托价成交，如果最新价高于委托价，按照最新价撮合成交，跌停不能卖出
			if decimal.NewFromFloat(cp).Cmp(bp) >= 0 {
				rBp = decimal.NewFromFloat(cp)
			} else {
				rBp = bp
			}

			//5、判断当前持仓是否有足够大股票数量
			stockHoldings := make([]*model.StockHolding, 0) //当前该持仓股票记录，存在多个的可能

			today := time.Now().Format(model.DATE_FORMAT)
			err := x.Where("stock_code=?", stockCode).And("holding_status=?", 1).And("trans_time<?", today).Cols("id,stock_number,available_number").Find(&stockHoldings)
			if err != nil && len(stockHoldings) > 0 {

				var availableNumber int32 = 0
				for _, sh := range stockHoldings {
					availableNumber += sh.AvailableNumber
				}
				if decimal.New(int64(availableNumber), 0).Cmp(rBq) < 0 { //可售数量小于持仓数量
					ret["msg"] = model.TRX_STOCK_ENT_MORE
					return 2, ret
				} else {
					//TODO:这里暂时停止 委托立即交易的模式
					// //只有这种情况可以直接出售
					// ret["msg"] = model.TRX_STOCK_ENT_OK
					// ret["rBq"] = rBq
					// ret["rBp"] = rBp
					// return 1, ret
				}

			} else {
				ret["msg"] = model.TRX_STOCK_ENT_FAIL
				return 2, ret
			}
			ret["msg"] = model.TRX_STOCK_ENT_OK

			return 1, ret
		}
	} else {
		ret["msg"] = model.SYS_ERR_Q_STOCK
		return 1, ret
	}
}

/*
*定时撮合交易
*撮合的时间段是有效交易的时间段
**/
func TrxMatch(x *xorm.Engine, log *log.Logger) {

	log.Println("撮合交易开始%s", time.Now())
	//0、统计所有待交易的数据，

	today := time.Now().Format(model.DATE_FORMAT)
	se := new(model.StockEntrust)
	//如果存在待交易数据

	if count, _ := x.Where("entrust_status=?", 0).And("trans_time<?", today).Count(se); count > 0 {

		//0-1、数据分片，开启协程处理
		page := int(count) / model.MATCH_LIMIT

		if int(count)%model.MATCH_LIMIT != 0 {
			page++
		}

		for i := 0; i < page; i++ { //分片获取数据
			//1、撮合第一步，获取所有买卖的委托
			stockEntrusts := make([]*model.StockEntrust, 0)
			err := x.Where("entrust_status=?", 0).And("trans_time<?", today).
				Limit(model.MATCH_LIMIT, i*model.MATCH_LIMIT).Find(&stockEntrusts)
			if err == nil {
				//开启lambda协程
				go func(x *xorm.Engine, entrustList []*model.StockEntrust) {
					//拼接股票代码
					stockList := ConcatStockList(stockEntrusts)
					if f, mapp := GetStock5Stages(stockList); f {
						for _, ent := range entrustList {
							v5 := mapp[ent.StockCode]
							log.Printf("判断是否可以进行交易：%s", ent.StockCode)
							if canTrx(ent, log, v5) { //判断是否可以进行交易
								log.Printf("可以进行交易：%s", ent.StockCode)
								doTrx(x.NewSession(), log, ent, v5) //进行交易
							}
						}
					}
				}(x, stockEntrusts)

			} else {
				log.Printf("查询出错了")
			}
		}

		//2、买卖开启协程进行匹配当前价格
	} else {
		log.Printf("没有待交易数据...")
	}
}

/**
是否可以进行交易
v5 买卖5档的详情
0:已交易，1:委托中，2:已取消，3:没有交易
//交易类型0:sale ,1:buy
**/
func canTrx(ent *model.StockEntrust, log *log.Logger, v5 []string) bool {

	if ent.EntrustStatus == 0 {
		return false
	}

	if ent.TransType == 0 { //sale
		//可以卖出的条件是 当前价格>=委托价格,并按照委托价格
		rp, _ := decimal.NewFromString(v5[3])
		entP := decimal.NewFromFloat(ent.EntrustPrice)
		//0、价格第一步判断
		if rp.Cmp(entP) < 0 {
			return false
		}
		//1、数量第二步判断
		dm, _ := decimal.NewFromString(v5[8])
		dealAmount := dm.Div(decimal.New(int64(100), 0)) //交易量(单位：手)
		if dealAmount.Cmp(decimal.Zero) <= 0 {
			return false
		}
		return true

	} else { //buy
		//可以买进的条件是 当前价格<=委托价格，并按照最新当前价格
		rp, _ := decimal.NewFromString(v5[3])
		entP := decimal.NewFromFloat(ent.EntrustPrice)
		if rp.Cmp(entP) >= 0 {
			return true
		}
		//1、数量第二步判断
		dm, _ := decimal.NewFromString(v5[8])
		dealAmount := dm.Div(decimal.New(int64(100), 0)) //交易量(单位：手)
		if dealAmount.Cmp(decimal.Zero) <= 0 {
			return false
		}
		return true
	}
}

/**
委托后的股票进行交易
处理流程 0、根据售价、售量 从持仓表中减去
事务处理 1、增加交易记录
		2、修改资金账户信息
**/
func doTrx(s *xorm.Session, log *log.Logger, ent *model.StockEntrust, v5 []string) {

	log.Printf("处理委托交易:%d,开启事务处理...\n", ent.Id)
	now := time.Now()
	s.Begin()
	//0、获取股票用户
	userAccount := new(model.UserAccount)
	if has, _ := s.Id(ent.UserId).Get(userAccount); !has {
		log.Printf("没有找到用户ID:%d的资金记录", ent.UserId)
		return
	}
	var trx *model.StockTrans
	var stockHolding *model.StockHolding

	//已经交易
	if ent.EntrustStatus == 0 {
		return
	}

	//查找持仓记录
	if has, _ := s.Where("user_id=?", ent.UserId).
		And("stock_code=?", ent.StockCode).
		And("holding_status=?", 1).Get(stockHolding); has {
		//存在纪录
		if ent.EntrustStatus == 0 { //继续委托
			stockHolding.HoldingStatus = 1
		} else {
			stockHolding.HoldingStatus = 0
		}

		stockHolding.StockNumber -= ent.RentrustNumber

		if stockHolding.StockNumber < 0 {
			stockHolding.StockNumber = 0 //全部买出
			stockHolding.HoldingStatus = 0
		}

	} else {
		return
	}

	//*********************增加交易流水************************
	trx = new(model.StockTrans)
	trx.StockCode = ent.StockCode
	trx.TransType = ent.TransType
	trx.TransStatus = 1
	trx.TransFrom = 0
	trx.TransNumber = ent.RentrustNumber
	trx.TransPrice = ent.RentrustPrice
	trx.TransTime = now

	//判断是否有订阅者是否需要通知
	if has := hasFollowers(s, log, userAccount.UserId); has {
		trx.NotifyStatus = 1
	} else {
		trx.NotifyStatus = 0
	}

	if _, err := s.Insert(trx); err != nil {
		log.Printf("插入交易失败,回滚,%s", trx)
		s.Rollback()
	}
	//*********************增加交易流水结束************************
	if ent.TransType == 0 { //sale

		//可以卖出的条件是 当前价格>=委托价格,并按照委托价格
		cp, _ := decimal.NewFromString(v5[3])
		entP := decimal.NewFromFloat(ent.EntrustPrice)
		var sp = cp
		var sq decimal.Decimal
		//0、价格第一步判断
		if cp.Cmp(entP) >= 0 {
			sp = entP
		}
		ssp, _ := sp.Float64() //出售价格
		ent.RentrustPrice = ssp

		//1、数量第二步判断
		dm, _ := decimal.NewFromString(v5[8])            //成交量
		dealAmount := dm.Div(decimal.New(int64(100), 0)) //交易量(单位：手)
		entq := decimal.New(int64(ent.EntrustNumber), 0) //委托量

		if dealAmount.Cmp(entq) >= 0 {

			ent.EntrustStatus = 0                     //全量交易
			q, _ := strconv.Atoi(entq.StringFixed(0)) //实际交易量
			ent.RentrustNumber = int32(q)
			sq = entq

			//当全量卖出 交易才算完全结束则进行统计，本次交易是否亏损
			// stockHolding.TransPrice

		} else { //剩余量继续委托
			ent.EntrustStatus = 0                                     //部分交易，继续委托
			q, _ := strconv.Atoi(entq.Sub(dealAmount).StringFixed(0)) //剩余量
			ent.EntrustNumber = int32(q)
			qq, _ := strconv.Atoi(dealAmount.StringFixed(0)) //实际交易量
			ent.RentrustNumber = int32(qq)
			sq = dealAmount
		}

		//结算资金
		saleMoney := sq.Mul(sp).Mul(decimal.New(100, 0))
		sm, _ := saleMoney.Float64() //进入可用资金中
		userAccount.AvailableAmount += sm

		_, err := s.Update(stockHolding)

		if err != nil {
			log.Printf("更新持仓表失败：%s", stockHolding.ToString())
			s.Rollback()
			return
		}
		//交易对处理
		trxInfos := make([]*model.StockTrxInfo, 0)
		s.Where("user_id = ?", ent.UserId).And("stock_code=?", ent.StockCode).
			And("trans_status=?", 0).Desc("id").Find(&trxInfos)
		//匹配交易对

		func() {

			num := decimal.Zero
			for _, trxInfo := range trxInfos {
				n := decimal.New(int64(trxInfo.LeftNumber), 0)
				num = num.Add(n)

				if num.Cmp(dealAmount) < 0 {
					trxInfo.TransStatus = 1 //交易结束
					trxInfo.SaleTime = now
					earings := sp.Sub(decimal.NewFromFloat(trxInfo.BuyPrice)).Mul(decimal.New(int64(100), 0)).Mul(n)
					trxInfo.Earning, _ = decimal.NewFromFloat(trxInfo.Earning).Add(earings).Float64()
					trxInfo.NotifyStatus = 2 //卖出待提醒
					trxInfo.LastTrxId = trx.Id
					s.Id(trxInfo.Id).Update(trxInfo)
				} else {
					if left := num.Sub(dealAmount); left.Cmp(decimal.Zero) > 0 {
						trxInfo.LeftNumber = int32(left.IntPart())
						trxInfo.SaleTime = now
						earings := sp.Sub(decimal.NewFromFloat(trxInfo.BuyPrice)).
							Mul(decimal.New(int64(100), 0)).
							Mul(decimal.New(int64(trxInfo.TransNumber-trxInfo.LeftNumber), 0))
						trxInfo.Earning, _ = decimal.NewFromFloat(trxInfo.Earning).Add(earings).Float64()
						trxInfo.NotifyStatus = 2 //卖出待提醒
						trxInfo.LastTrxId = trx.Id
					} else {
						trxInfo.TransStatus = 1 //交易结束
						trxInfo.SaleTime = now
						earings := sp.Sub(decimal.NewFromFloat(trxInfo.BuyPrice)).Mul(decimal.New(int64(100), 0)).Mul(n)
						trxInfo.Earning, _ = decimal.NewFromFloat(trxInfo.Earning).Add(earings).Float64()
						trxInfo.NotifyStatus = 2 //卖出待提醒
						trxInfo.LastTrxId = trx.Id
					}
					s.Id(trxInfo.Id).Update(trxInfo)
					break
				}
			}
		}()

	} else { //buy

		//可以买进的条件是 当前价格<=委托价格，并按照最新当前价格
		cp, _ := decimal.NewFromString(v5[3])
		entP := decimal.NewFromFloat(ent.EntrustPrice)
		var sp = entP
		if cp.Cmp(entP) <= 0 {
			sp = cp
		}
		ssp, _ := sp.Float64() //出售价格
		ent.RentrustPrice = ssp

		//1、数量第二步判断
		dm, _ := decimal.NewFromString(v5[8])            //成交量
		dealAmount := dm.Div(decimal.New(int64(100), 0)) //交易量(单位：手)
		entq := decimal.New(int64(ent.EntrustNumber), 0) //委托量

		if dealAmount.Cmp(entq) >= 0 {
			ent.EntrustStatus = 0                     //全量交易
			q, _ := strconv.Atoi(entq.StringFixed(0)) //实际交易量
			ent.RentrustNumber = int32(q)
		} else { //剩余量继续委托
			ent.EntrustStatus = 0                                     //部分交易，继续委托
			q, _ := strconv.Atoi(entq.Sub(dealAmount).StringFixed(0)) //剩余量
			ent.EntrustNumber = int32(q)
			qq, _ := strconv.Atoi(dealAmount.StringFixed(0)) //实际交易量
			ent.RentrustNumber = int32(qq)
		}

		//结算资金---->在委托的时候已经进行减去
		// saleMoney := sq.Mul(sp).Mul(decimal.New(100, 0))
		// sm, _ := saleMoney.Float64() //进入可用资金中
		// userAccount.AvailableAmount -= sm

		//持仓记录新增
		stockHolding := new(model.StockHolding)
		stockHolding.StockCode = ent.StockCode
		stockHolding.StockNumber = ent.RentrustNumber
		stockHolding.AvailableNumber = ent.RentrustNumber
		stockHolding.HoldingStatus = 1
		stockHolding.TransPrice = ent.RentrustPrice
		stockHolding.TransTime = now

		if _, err := s.Insert(stockHolding); err != nil {
			log.Printf("插入持仓表失败：%s", stockHolding.ToString())
			s.Rollback()
			return
		}

	}

	//判断首次交易时间
	if userAccount.TransStart.IsZero() {
		userAccount.TransStart = now
	}

	//记录交易对
	trxInfo := new(model.StockTrxInfo)
	trxInfo.StockCode = ent.StockCode
	trxInfo.UserId = ent.UserId
	trxInfo.TransNumber = ent.RentrustNumber
	trxInfo.LeftNumber = ent.RentrustNumber
	trxInfo.BuyTime = now
	trxInfo.BuyPrice = ent.EntrustPrice
	trxInfo.TransStatus = 0  //买入中
	trxInfo.NotifyStatus = 0 //买入待提醒
	trxInfo.LastTrxId = trx.Id

	if _, err := s.Insert(trxInfo); err != nil {
		log.Printf("插入交易对失败：%s", err.Error())
		s.Rollback()
		return
	}

	//更新数据库
	if _, err := s.Update(ent, userAccount); err == nil {
		if err == nil {
			s.Commit()
		} else {
			log.Printf("插入交易失败,回滚,%s", trx)
			s.Rollback()
		}
	} else {
		log.Printf("更新失败，回滚%s", trx)
		s.Rollback()
	}

	log.Printf("处理委托交易结束:%d,开启事务处理...\n", ent.Id)
}

/*
进行委托
或者，直接委托交易
事务处理
*/
func doEntrust(x *xorm.Engine, userId int64, stockCode string, quantity decimal.Decimal, buyPrice decimal.Decimal, trxType int8, entrustStatus int8) (string, string) {

	session := x.NewSession()
	session.Begin() //开启事务
	defer session.Close()
	//资金账户信息
	userAccount := new(model.UserAccount)
	var err error

	if has, _ := session.Where("user_id = ?", userId).Get(userAccount); !has {
		return "100", "没有找到对应用户账户信息"
	}

	now := time.Now()
	q, _ := strconv.Atoi(quantity.StringFixed(0))
	f, _ := buyPrice.Float64()

	//0、写入委托表
	stockEntrust := new(model.StockEntrust)
	stockEntrust.StockCode = stockCode
	stockEntrust.UserId = userId
	stockEntrust.EntrustPrice = f
	stockEntrust.EntrustNumber = int32(q)
	stockEntrust.EntrustTime = now
	stockEntrust.EntrustStatus = entrustStatus
	stockEntrust.TransType = trxType

	//修改资金账户信息
	if trxType == 1 {
		a, _ := (decimal.NewFromFloat(userAccount.AvailableAmount).Sub(quantity.Mul(buyPrice.Mul(decimal.New(100, 0))))).Float64()
		if a < 0 {
			a = 0
		}
		userAccount.AvailableAmount = a
	}

	//判断首次交易时间
	if userAccount.TransStart.IsZero() {
		userAccount.TransStart = now
	}

	_, err = session.Update(userAccount)
	if err == nil {

		_, err = session.Insert(stockEntrust)

		if err != nil {
			session.Rollback()
			fmt.Errorf("%d委托失败:%d,%s,%s,%s", userId, trxType, stockCode, buyPrice, quantity)
			return "100", model.TRX_STOCK_FAIL
		} else {

			stockHolding := new(model.StockHolding)
			if trxType == 0 { //交易卖出
				if has, _ := x.Where("user_id=?", userId).And("stock_code=?", stockCode).Get(stockHolding); !has {
					session.Rollback()
					fmt.Errorf("%d委托失败:%d,%s,%s,%s", userId, trxType, stockCode, buyPrice, quantity)
					return "100", model.TRX_STOCK_FAIL
				} else { //委托卖出成功，修改可以委托数量
					stockHolding.AvailableNumber = stockHolding.StockNumber - int32(q)
				}
			}
			_, err = x.Update(stockHolding)

		}
	} else {
		session.Rollback()
		fmt.Errorf("%d委托失败:%d,%s,%s,%s", userId, trxType, stockCode, buyPrice, quantity)
		return "100", model.TRX_STOCK_FAIL
	}

	//提交
	err = session.Commit()
	if err == nil {
		fmt.Printf("%d委托成功:%d,%s,%s,%s", userId, trxType, stockCode, buyPrice, quantity)
		return "200", model.TRX_STOCK_OK
	} else {
		fmt.Errorf("%d委托失败:%d,%s,%s,%s", userId, trxType, stockCode, buyPrice, quantity)
		return "100", model.TRX_STOCK_FAIL
	}

}

/**
用户撤单
事务处理
**/
func CancelEntrustHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {

	jr := new(model.JsonResult)
	jr.Code = "100"
	jr.Msg = model.TRX_STOCK_CAN_ENT_FAIL
	entId := ctx.Params(":entId")

	log.Printf("用户撤单%s", entId)
	session := x.NewSession()
	session.Begin()
	defer session.Close()
	var err error

	stockEntrust := new(model.StockEntrust)
	if has, _ := session.Id(entId).Get(stockEntrust); has {
		if stockEntrust.EntrustStatus == 0 { //可以取消

			stockEntrust.EntrustStatus = 2
			_, err = session.Update(stockEntrust)

			if err == nil {
				//恢复可用金额
				_, userAccount := QueryUserAccoutByUserIdWithSession(session, stockEntrust.UserId)
				if stockEntrust.TransType == 1 { //买入委托取消后，将买入的金额恢复到可用金额
					d, _ := (decimal.NewFromFloat(userAccount.AvailableAmount).Add(decimal.New(int64(stockEntrust.EntrustNumber), 0).Mul(decimal.New(100, 0)).Mul(decimal.NewFromFloat(stockEntrust.EntrustPrice)))).Float64()
					userAccount.AvailableAmount = d
					_, err = session.Update(userAccount)
					if err != nil {
						session.Rollback()
						ctx.JSON(200, jr)
					} else {
						jr.Code = "200"
						jr.Msg = model.TRX_STOCK_CAN_ENT_OK
						ctx.JSON(200, jr)
						session.Commit()
						ctx.JSON(200, jr)
					}
				}

			} else {
				session.Rollback()
				ctx.JSON(200, jr)
			}

		} else {
			ctx.JSON(200, jr)
		}
	} else {
		ctx.JSON(200, jr)
	}
}

/**
取消所有委托
**/
func CanCellAllEntrust(session *xorm.Session) {

	session.Begin()
	defer session.Close()
	_, err := session.Exec("update stock_entrust set entrust_status = 2 where entrust_status = 1")
	if err == nil {
		_, err = session.Exec("update stock_holding set available_number = stock_number where holding_status=1")
		if err == nil {
			session.Commit()
		} else {
			session.Rollback()
		}
	} else {
		session.Rollback()
	}
}

/**
判断一个股票是否处于一个停盘的状态
**/
func CheckStockStopService(x *xorm.Engine, log *log.Logger, stockCode string) bool {

	taskExcludeDate := new(model.TaskExcludeDate)
	if has, _ := x.Where("stock_code=?", stockCode).Get(taskExcludeDate); has {

		if taskExcludeDate.EndDate.After(time.Now()) {
			return true
		}
		return false
	}
	return false
}

func Stock5StageHander(sess session.Store, ctx *macaron.Context, redis *redis.Client) {

	r := new(model.JsonResult)
	stockCode := ctx.Params(":stockCode")

	if stockDetail := GetRedisStockDetail(redis, stockCode); stockDetail != "" {
		r.Code = "200"
		r.Data = stockDetail
	} else {
		r.Code = "100"
		r.Data = stockDetail
	}
	ctx.JSON(200, r)
}
