package handler

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"github.com/go-macaron/session"
	"github.com/shopspring/decimal"

	"../model"
	"github.com/go-xorm/xorm"
)

func Chk(err error) {
	if err != nil {
		panic(err)
	}
}

/**
16进制转
**/
func Hex(id int64) string {
	return hex.EncodeToString([]byte(strconv.FormatInt(id, 10)))
}

/**
session 中获取用户信息
**/
func GetSessionUser(sess session.Store) (bool, *model.User) {
	if sess.Get("user") != nil {
		user := sess.Get("user").(model.User)
		return true, &user
	} else {
		return false, nil
	}
}

//判断一天是否可以交易
func DayCanTrx(x *xorm.Engine, day string) bool {
	return true
}

/**
定时获取国家法定假日
**/
func GetNationHolidays() {

}

/*
获取股票的买卖五档
http://hq.sinajs.cn/list=sh601006,sh000911
var hq_str_sh601006="大秦铁路,8.440,8.440,8.310,8.480,8.300,8.300,8.310,26974939,225811186.000,613793,8.300,58700,8.290,245100,8.280,93600,8.270,192200,8.260,103400,8.310,151577,8.320,161700,8.330,157600,8.340,85300,8.350,2017-05-17,15:00:00,00";
var hq_str_sh000911="300可选,5006.1141,5011.2990,4988.1654,5043.7715,4982.8774,0,0,11218124,15794661106,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,2017-05-17,15:01:03,00";
*/
func GetStock5Stages(stockList string) (bool, map[string][]string) {

	if len(stockList) == 0 {
		return false, nil
	}
	resp, _ := http.Get(model.STOCK_5_STAGES_API + "list=" + stockList)
	//一定要关闭
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.Status == "200" { //相应成功
		bodyStr := string(body)
		// 定义切片
		var stockCodes = make([]string, 0)
		var result = make(map[string][]string, 0)

		if strings.Index(stockList, ",") > 0 { //多个股票
			stockCodes = strings.Split(stockList, ",")
		} else {
			stockCodes[0] = stockList
		}
		x := strings.Split(strings.Trim(bodyStr, "\n"), "\n")

		for i, s := range x {
			stockCode := strings.Trim(strings.Trim(stockCodes[i], model.EXC_SH), model.EXC_SZ)
			result[stockCode] = strings.Split(s, ",")
		}
		return true, result

	} else {
		return false, nil
	}
}

func GetStock5StagesWithChan(ch chan map[string]string, stockList string) {

	if len(stockList) == 0 {
		return
	}

	resp, _ := http.Get(model.STOCK_5_STAGES_API + "list=" + stockList)

	//一定要关闭
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	if resp.Status == "200" { //相应成功
		bodyStr := string(body)
		// 定义切片
		var stockCodes = make([]string, 0)
		var result = make(map[string]string, 0)

		if strings.Index(stockList, ",") > 0 { //多个股票
			stockCodes = strings.Split(stockList, ",")
		} else {
			stockCodes[0] = stockList
		}
		x := strings.Split(strings.Trim(bodyStr, "\n"), "\n")

		for i, s := range x {
			stockCode := strings.Trim(strings.Trim(stockCodes[i], model.EXC_SH), model.EXC_SZ)
			result[stockCode] = s
		}
		ch <- result
	}
}

/**
判断一个股票是上证A股，--》并在股票代码加上sh
判断一个股票是深证A股，--》并在股票代码加上sz
**/
func AddExcToStockCode(stockCode string) string {

	if strings.Index(stockCode, "60") == 0 {
		return model.EXC_SH + stockCode
	} else if strings.Index(stockCode, "000") == 0 {
		return model.EXC_SZ + stockCode
	} else {
		return ""
	}
}

/**
判断一个股票是上证A股，--》并在带sh股票代码减去sh
判断一个股票是深证A股，--》并在带sz股票代码减去sz
**/
func TrimExcFromStockCode(stockCode string) string {
	return strings.Trim(strings.Trim(stockCode, model.EXC_SH), model.EXC_SZ)
}

/**
计算涨停板，跌停板
closePrice : 昨日收盘价格
percent=0.1  -->up
percent=-0.1 -->down
**/
func GetLimitPrice(closePrice float64, percent float64, operator int) decimal.Decimal {
	return decimal.NewFromFloat(closePrice).Mul((decimal.NewFromFloat(percent).Mul(decimal.New(int64(operator), 1))).Add(decimal.New(int64(1), 64))).Round(2)
}

/**
判断一个股票是否是特殊处理的股票
加 *ST
**/
func IsST(stockName string) bool {
	return strings.Index(stockName, "ST") > 0
}

/**
拼接股票代理
**/
func ConcatStockList(stockList interface{}) string {

	var buffer bytes.Buffer
	switch t := stockList.(type) {
	case []*model.StockEntrust:
		{
			len := len(t)
			if len == 1 {
				buffer.WriteString(AddExcToStockCode(t[0].StockCode))
			} else {
				for i, v := range t {
					if i < len-1 {
						buffer.WriteString(AddExcToStockCode(v.StockCode) + ",")
					} else {
						buffer.WriteString(AddExcToStockCode(v.StockCode))
					}
				}
			}

		}
	case []*model.StockTrans:
		{
			len := len(t)
			if len == 1 {
				buffer.WriteString(AddExcToStockCode(t[0].StockCode))
			} else {
				for i, v := range t {
					if i < len-1 {
						buffer.WriteString(AddExcToStockCode(v.StockCode) + ",")
					} else {
						buffer.WriteString(AddExcToStockCode(v.StockCode))
					}
				}
			}
		}
	case []*model.Stock:
		{
			len := len(t)
			if len == 1 {
				buffer.WriteString(AddExcToStockCode(t[0].StockCode))
			} else {
				for i, v := range t {
					if i < len-1 {
						buffer.WriteString(AddExcToStockCode(v.StockCode) + ",")
					} else {
						buffer.WriteString(AddExcToStockCode(v.StockCode))
					}
				}
			}
		}
	case nil:
		fmt.Printf("nil value: nothing to check?\n")
	default:
		fmt.Printf("Unexpected type %T\n", t)
	}

	return buffer.String()
}

/**
反射调用方法
**/
func InvokeObjectMethod(object interface{}, methodName string, args ...interface{}) {
	inputs := make([]reflect.Value, len(args))
	for i, _ := range args {
		inputs[i] = reflect.ValueOf(args[i])
	}
	reflect.ValueOf(object).MethodByName(methodName).Call(inputs)
}
