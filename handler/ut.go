package handler

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"reflect"

	"strconv"
	"strings"
	"time"

	"github.com/axgle/mahonia"
	redis "github.com/go-redis/redis"

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
md5 加密
**/
func Md5(text string) string {
	hash := md5.New()
	hash.Write([]byte(text)) // 需要加密的字符串为 123456
	cipherStr := hash.Sum(nil)
	return hex.EncodeToString(cipherStr)
}

/*
 格式化比例
*/
func FormateRate(rate float64) string {
	return fmt.Sprintf("%s%s", decimal.NewFromFloat(rate).Mul(decimal.New(100, 0)).StringFixed(2), "%")
}

/*
 掩码
*/
func MaskStockCode(code string) string {

	if len := len(code); len > 0 {
		return strings.Repeat("*", len)
	}
	return "N/A"
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
		user := sess.Get("user").(*model.User)
		return true, user
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

	path := model.STOCK_5_STAGES_API + "list=" + stockList
	client := &http.Client{}
	reqest, _ := http.NewRequest("GET", path, nil)

	// reqest.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	reqest.Header.Set("Accept-Charset", "utf-8;q=0.7,*;q=0.3")
	// reqest.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6")
	reqest.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_12_1) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.36")

	// resp, _ := http.Get(path)
	resp, _ := client.Do(reqest)
	//一定要关闭
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode == 200 { //相应成功

		bodyStr := string(body)

		// 定义切片
		var stockCodes []string
		var result = make(map[string][]string, 0)

		if strings.Index(stockList, ",") > 0 { //多个股票
			stockCodes = strings.Split(stockList, ",")
		} else {
			stockCodes = make([]string, 1)
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

/*
协程获取5档数据
**/
func GetStock5StagesWithChan(ch chan map[string]interface{}, stockList string) {

	if len(stockList) == 0 {
		return
	}
	resp, _ := http.Get(model.STOCK_5_STAGES_API + "list=" + stockList)

	//一定要关闭
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	if resp.StatusCode == 200 { //相应成功
		bodyStr := string(body)
		// 定义切片
		var stockCodes []string
		var result = make(map[string]interface{}, 0)
		if strings.Index(stockList, ",") > 0 { //多个股票
			stockCodes = strings.Split(stockList, ",")
		} else {
			stockCodes = make([]string, 1)
			stockCodes[0] = stockList
		}

		x := strings.Split(strings.Trim(bodyStr, "\n"), "\n")
		for i, s := range x {
			if !strings.Contains(s, ",") {
				continue
			} else {
				stockCode := strings.Trim(strings.Trim(stockCodes[i], model.EXC_SH), model.EXC_SZ)
				result[stockCode] = s
			}
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
	return decimal.NewFromFloat(closePrice).Mul((decimal.NewFromFloat(percent).Mul(decimal.New(int64(operator), 0))).Add(decimal.New(int64(1), 0))).Round(2)
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
				buffer.WriteString(t[0].StockCode)
			} else {
				for i, v := range t {
					if i < len-1 {
						buffer.WriteString(v.StockCode + ",")
					} else {
						buffer.WriteString(v.StockCode)
					}
				}
			}
		}
	case []string:
		{
			len := len(t)
			if len == 1 {
				buffer.WriteString(t[0])
			} else {
				for i, v := range t {
					if i < len-1 {
						buffer.WriteString(v + ",")
					} else {
						buffer.WriteString(v)
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

/**
通过redis 实现同获取订单号自增
**/
func OrderSnGenerator(redisCli *redis.Client) string {

	var lock = "order_sn_lock"
	t := time.Now().Format(model.DATE_ORDER_FORMAT)
	for {
		if f, _ := redisCli.SetNX(lock, "lock", 3*time.Second).Result(); f {
			s, _ := redisCli.Get("order_sn").Result()
			if s == "" {
				s = "0001"
				redisCli.Set("order_sn", s, 24*time.Hour)
				redisCli.Del(lock)
				return t + s
			} else {
				i, _ := strconv.Atoi(s)
				i++
				d := formateSn(strconv.Itoa(i))
				fmt.Print("%s", d)
				redisCli.Set("order_sn", d, 24*time.Hour)
				redisCli.Del(lock)
				return t + d
			}
		} else {
			time.Sleep(500 * time.Millisecond)
		}
	}
}

func formateSn(i string) string {
	len0 := (4 - len(i))
	switch len0 {
	case 0:
		return i
	case 1:
		return "0" + i
	case 2:
		return "00" + i
	case 3:
		return "000" + i
	default:
		return i
	}
}

/*
获取用户昵称
*/
func GetUserNickName(x *xorm.Engine, r *redis.Client, id interface{}) string {

	var i int64
	switch id.(type) {
	case int64:
		i = id.(int64)
	case int32:
		i = int64(id.(int32))
	case int:
		i = int64(id.(int))
	case string:
		d, _ := strconv.Atoi(id.(string))
		i = int64(d)
	default:
		i = 0
	}
	if u := GetUserById(x, r, i); u != nil {
		return u.NickName
	}
	return ""
}

/*
格式化标准日期
**/
func FormatDate(d time.Time) string {
	return d.Format(model.DATE_FORMAT)
}

/*
格式化日期时间
*/
func FormatDateTime(d time.Time) string {
	return d.Format(model.DATE_TIME_FORMAT)
}

func ConvertToString(src string, srcCode string, tagCode string) string {

	srcCoder := mahonia.NewDecoder(srcCode)
	srcResult := srcCoder.ConvertString(src)
	tagCoder := mahonia.NewDecoder(tagCode)
	_, cdata, _ := tagCoder.Translate([]byte(srcResult), true)
	result := string(cdata)
	return result
}

func TestN() {
	fmt.Print(strings.Count("ss,ss,ss", ","))
}

/**
地址缩短工具
**/
func ShortMe(path string) (bool, string) {
	x := url.URL{}
	x.Host = model.ME_HOST
	x.Scheme = model.ME_SCHEMA
	x.Path = path
	d := fmt.Sprintf(model.SHORT_API, x.EscapedPath())
	_, shortPath := HttpGet(d)
	if shortPath == "" {
		return false, path
	} else {
		return true, shortPath
	}
}

//计算股票市值
func StockValue(r *redis.Client, stockCode string, num int32) string {

	if j := GetRedisStockDetail(r, stockCode); j != "" {
		if x := strings.Split(j, ","); len(x) > 0 {
			p, _ := decimal.NewFromString(x[3])
			return decimal.New(int64(num), 0).Mul(p).StringFixed(2)
		}
	}
	return model.NA
}

/**
根据 @readme中股票的下表获取股票某一个字段的值
-1 表示整个信息
**/
func StockDetail(r *redis.Client, stockCode string, index int) string {

	if j := GetRedisStockDetail(r, stockCode); j != "" {
		if x := strings.Split(j, ","); len(x) > 0 && index < len(x) {
			return x[index]
		}
	}
	return model.NA
}

/**
浮动盈亏,盈亏比例
返回值：earning
**/
func FloatEarning(cp interface{}, transPrice interface{}, num int32) string {

	var earning string
	var x, y decimal.Decimal
	switch t := cp.(type) {
	case string:
		{
			x, _ = decimal.NewFromString(t)
			switch tp := transPrice.(type) {
			case string:
				{
					y, _ = decimal.NewFromString(tp)
				}
			case float32:
				{
					y = decimal.NewFromFloat(float64(tp))

				}
			case float64:
				{
					y = decimal.NewFromFloat(tp)
				}
			}
		}

	case float32:
		{
			x = decimal.NewFromFloat(float64(t))

			switch tp := transPrice.(type) {
			case string:
				{
					y, _ = decimal.NewFromString(tp)
				}
			case float32:
				{
					y = decimal.NewFromFloat(float64(tp))

				}
			case float64:
				{
					y = decimal.NewFromFloat(tp)
				}
			}
		}
	case float64:
		{
			x = decimal.NewFromFloat(t)
			switch tp := transPrice.(type) {
			case string:
				{
					y, _ = decimal.NewFromString(tp)
				}
			case float32:
				{
					y = decimal.NewFromFloat(float64(tp))
				}
			case float64:
				{
					y = decimal.NewFromFloat(tp)
				}
			}
		}
	}

	earning = decimal.New(int64(num), 0).Mul(x.Sub(y)).StringFixed(2)
	return earning
}

/**
盈亏比例
rate
**/
func EarningRate(cp interface{}, transPrice interface{}) string {

	var rate string
	var x, y decimal.Decimal
	switch t := cp.(type) {
	case string:
		{
			x, _ = decimal.NewFromString(t)
			switch tp := transPrice.(type) {
			case string:
				{
					y, _ = decimal.NewFromString(tp)
				}
			case float32:
				{
					y = decimal.NewFromFloat(float64(tp))

				}
			case float64:
				{
					y = decimal.NewFromFloat(tp)
				}
			}
		}

	case float32:
		{
			x = decimal.NewFromFloat(float64(t))

			switch tp := transPrice.(type) {
			case string:
				{
					y, _ = decimal.NewFromString(tp)
				}
			case float32:
				{
					y = decimal.NewFromFloat(float64(tp))

				}
			case float64:
				{
					y = decimal.NewFromFloat(tp)
				}
			}
		}
	case float64:
		{
			x = decimal.NewFromFloat(t)
			switch tp := transPrice.(type) {
			case string:
				{
					y, _ = decimal.NewFromString(tp)
				}
			case float32:
				{
					y = decimal.NewFromFloat(float64(tp))
				}
			case float64:
				{
					y = decimal.NewFromFloat(tp)
				}
			}
		}
	}
	if y.Cmp(decimal.Zero) == 0 {
		return model.NA
	}
	f, _ := x.Sub(y).DivRound(y, 2).Float64()
	rate = FormateRate(f)
	return rate
}

/*

 */
func isMobile(mobile string) bool {
	return true
}

/**
数字``
**/
func RandomIntCode() string {
	var ret string
	rand.Seed(int64(time.Now().Nanosecond()))
	for i := 0; i < 4; i++ {
		ret += strconv.Itoa(rand.Intn(10))
	}
	return ret
}

/**
字母
**/
func RandomStringCode(len int) string {
	var ret string
	rand.Seed(int64(time.Now().Nanosecond()))
	arrays := strings.Split("abcdefghijklmnopqrstuvwxyz", "")
	for i := 0; i < len; i++ {
		ret += arrays[rand.Intn(26)]
	}
	return ret
}

// 字符串判断空
func StringNul(x string) bool {
	if x == "" {
		return true
	}
	return false
}

func checkInService() bool {
	return true
}

func GenerateRandomChar() {

}
