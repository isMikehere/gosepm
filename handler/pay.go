package handler

import (
	"encoding/xml"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/ascoders/alipay"

	"../model"

	"strconv"

	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"github.com/stephenlyu/wxpay"
	macaron "gopkg.in/macaron.v1"
)

//
// ************************AliPay************************
//

func alipaySubmitOrder(client alipay.Client, ctx *macaron.Context) {

	form := client.Form(alipay.Options{
		OrderId:  "123",   // 唯一订单号
		Fee:      99.8,    // 价格
		NickName: "翱翔大空",  // 用户昵称，支付页面显示用
		Subject:  "充值100", // 支付描述，支付页面显示用
	})
	log.Printf(form)
	// ctx.HTMLString(form)

}

//
// *********************************微信支付*********************************
//

const WX_APP_KEY = "2ec3dfca00f46cce2c7e64e6cc71e70b"

const SERVER_ADDR string = "localhost：4000"
const SERVER_IP string = "10.0.0.1"

func wxNewAppTrans() *wxpay.AppTrans {
	//初始化
	cfg := &wxpay.WxConfig{
		AppId:         "wx1c355d3126305953",
		AppKey:        WX_APP_KEY,
		MchId:         "1284856501",
		NotifyUrl:     fmt.Sprintf("http://%s/weixin/notify", SERVER_ADDR),
		PlaceOrderUrl: "https://api.mch.weixin.qq.com/pay/unifiedorder",
		QueryOrderUrl: "https://api.mch.weixin.qq.com/pay/orderquery",
		TradeType:     "APP",
	}
	appTrans, err := wxpay.NewAppTrans(cfg)
	Chk(err)
	return appTrans
}

/*
支付宝预下单
*
*/
func WxCreatePrepayOrder(order *model.StockOrder, user *model.User, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {

	var prepayId string
	var err error
	appTrans := wxNewAppTrans()

	var desc string

	var ts = time.Now().Unix() % 0xFFFFFF
	var outTradeNo = fmt.Sprintf("%s%06x", Hex(order.Id), ts)

	reg := regexp.MustCompile("[<>&\"]")
	desc = string(reg.ReplaceAll([]byte(desc), []byte("|")))

	//获取prepay id，手机端得到prepay id后加上验证就可以使用这个id发起支付调用
	// var payAmount = (order.OrderAmount - order.BonusAmount) * 100
	prepayId, err = appTrans.Submit(outTradeNo, order.PayAmount, desc, SERVER_IP, user.OpenId)
	Chk(err)

	//预支付
	err = UpdateOrderTradeNo(x, order, outTradeNo, "weixin", prepayId)

	Chk(err)

	//加上Sign，已方便手机直接调用
	payRequest := appTrans.NewPaymentRequest(prepayId)
	ctx.JSON(200, payRequest)
}

func WxQueryPayResult(ctx *macaron.Context) {
	prepayOrderId := strings.TrimSpace(ctx.Req.FormValue("prepayId"))

	appTrans := wxNewAppTrans()
	queryResult, err := appTrans.Query(prepayOrderId)
	Chk(err)
	ctx.JSON(200, queryResult)
}

type WxNotifyResult struct {
	ReturnCode string `xml:"return_code"`
	ReturnMsg  string `xml:"return_msg"`
}

type WxNotifyData struct {
	AppId         string `xml:"appid"`
	MchId         string `xml:"mch_id"`
	DeviceInfo    string `xml:"device_info`
	NonceStr      string `xml:"nonce_str"`
	Sign          string `xml:"sign"`
	ResultCode    string `xml:"result_code"`
	ReturnCode    string `xml:"return_code"`
	ErrCode       string `xml:"err_code"`
	ErrCodeDes    string `xml:"err_code_des"`
	OpenId        string `xml:"openid"`
	IsSubscribe   string `xml:"is_subscribe"`
	TradeType     string `xml:"trade_type"`
	BankType      string `xml:"bank_type"`
	TotalFee      string `xml:"total_fee"`
	FeeType       string `xml:"fee_type"`
	CashFee       string `xml:"cash_fee"`
	CashFeeType   string `xml:"cash_fee_type`
	Coupon_fee    string `xml:"coupon_fee"`
	CouponCount   string `xml:"coupon_count"`
	CouponIdN     string `xml:"coupon_id_$n"`
	CouponFeeN    string `xml:"coupon_fee_$n"`
	TransactionId string `xml:"transaction_id"`
	OutTradeNo    string `xml:"out_trade_no"`
	Attach        string `xml:"attach"`
	TimeEnd       string `xml:"time_end"`
}

func WxNotifyHandler(ctx *macaron.Context, x *xorm.Engine, log *log.Logger) (int, string) {

	log.Printf("微信支付异步通知\n")
	data, err := ctx.Req.Body().Bytes()
	Chk(err)

	fmt.Printf("%s\n", string(data))

	notiData := WxNotifyData{}
	err = xml.Unmarshal(data, &notiData)

	Chk(err)

	nmap, err := wxpay.ToMap(&notiData)
	wantSign := wxpay.Sign(nmap, WX_APP_KEY)
	gotSign := nmap["sign"]
	if wantSign != gotSign {
		fmt.Println("[WxNotifyHandler] verify sign fail, gotSign: %s, wantSign: %s", gotSign, wantSign)
		result := &WxNotifyResult{ReturnCode: "FAIL", ReturnMsg: "sign error"}
		s, err := xml.Marshal(result)
		Chk(err)
		return 200, string(s)
	}

	if notiData.ResultCode == "SUCCESS" { //交易成功
		fmt.Printf("Update")
		err := UpdateOrderPayStatus(x, model.ORDER_STATUS_NOT_PAYED, notiData.OutTradeNo, "weixin")
		if err != nil {
			log.Print("error: update order status, order id: %s trade no: %s", notiData.OutTradeNo, notiData.TransactionId)
		}
	}

	result := &WxNotifyResult{ReturnCode: "SUCCESS", ReturnMsg: "OK"}
	s, err := xml.Marshal(result)
	Chk(err)
	return 200, string(s)
}

//测试支付路由
func DevPayHandler(ctx *macaron.Context, x *xorm.Engine, r *redis.Client) {

	orderId, _ := strconv.Atoi(ctx.Params(":orderId"))
	payType := ctx.Params(":payType")
	_, msg := OrderPayed(x, r, orderId, payType)
	log.Printf("%s", msg)
	ctx.Data["msg"] = msg
	ctx.HTML(200, "follow_step3")
}

/**
订单处理*
**/
func OrderPayed(x *xorm.Engine, r *redis.Client, orderId int, payType string) (bool, string) {

	order := new(model.StockOrder)
	if has, _ := x.Id(orderId).Get(order); has {
		//判断是否已经支付
		if order.OrderStatus != 0 {
			return false, "订单状态异常，请检查订单"
		}
	} else {
		return false, "订单不存在"
	}

	now := time.Now()
	s := x.NewSession()
	s.Begin()
	defer s.Close()

	order.OrderStatus = 1
	order.PayType = payType
	order.PayTime = now

	uf := new(model.UserFollow)
	uf.UserId = order.UserId
	uf.FollowedId = order.FollowedId
	uf.FollowType = order.ProductType
	uf.FollowStart = now
	uf.OrderId = order.Id
	var weeks = 1
	uf.FollowEnd = now.Add(1 * 7 * 24 * time.Hour)
	if order.ProductType == 1 {
		weeks = 4
		uf.FollowEnd = now.Add(4 * 7 * 24 * time.Hour)
	}
	uf.FollowStatus = 0
	//开始更新
	_, err := s.Id(orderId).Update(order)
	if err != nil {
		log.Printf("出现异常%s", err.Error())
		s.Rollback()
		return false, "订单更新失败,请联系客服"
	}

	user := new(model.User)         //订阅人
	followedUser := new(model.User) //被订阅人
	x.Id(order.UserId).Get(user)
	x.Id(order.UserId).Get(followedUser)

	//更新用户的订阅量
	userAccount := new(model.UserAccount)
	if has, _ := s.Where("user_id=?", followedUser.Id).Get(userAccount); has {
		userAccount.TotalFollow = userAccount.TotalFollow + 1
		_, err = s.Id(userAccount.Id).MustCols("total_follow").Update(userAccount)
		if err != nil {
			log.Printf("出现异常%s", err.Error())
			s.Rollback()
			return false, "订单更新失败,请联系客服"
		}
	} else {
		s.Rollback()
	}

	// 【金修网络】恭喜您：%s已经成功订阅您的为期%d周股票模拟交易提醒，有效期为%s-%s。请及时处理详情请参考订单须知。
	messageLog := new(model.MessageLog)
	messageLog.Mobile = followedUser.Mobile
	messageLog.InBatchId = time.Now().Format(model.DATE_ORDER_FORMAT)
	messageLog.Content =
		fmt.Sprintf(model.TOBEFOLLOWED_OK_MSG, user.NickName, weeks, uf.FollowStart.Format(model.DATE_TIME_FORMAT),
			uf.FollowEnd.Format(model.DATE_TIME_FORMAT))
	messageLog.SendStatus = 0
	//发布消息队列
	PublishMessage(r, model.R_MSG_SEND_CHAN, messageLog)

	_, err = s.Insert(uf)
	if err != nil {
		log.Printf("出现异常%s", err.Error())
		s.Rollback()
		return false, "订单更新失败,请联系客服"
	}
	s.Commit()
	return true, "订单更新成功"
}
