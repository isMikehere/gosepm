package handler

import (
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/go-macaron/session"
	"github.com/isMikehere/alipay"

	"../model"

	"strconv"

	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	"github.com/stephenlyu/wxpay"
	macaron "gopkg.in/macaron.v1"
)

/*
Pay router
*/
//noinspection GoUnusedExportedFunction
func Pay(ctx *macaron.Context, x *xorm.Engine, sess session.Store, client alipay.Client) {
	//check the order status whether is payed
	orderID := ctx.Params(":orderId")
	id, _ := strconv.Atoi(orderID)
	if has, order := GetOrderByOrderId(x, int64(id)); has {
		if order.OrderStatus == model.ORDER_STATUS_NOT_PAYED {

			//get the buyer
			_, buyer := GetSessionUser(sess)
			payType := ctx.Params(":payType")
			if payType == model.AliPay {
				AlipaySubmitOrder(ctx, x, client, buyer, order)
			} else {

			}

		} else {
			return
		}

	} else {
		ctx.Data["msg"] = "订单不存在"
		ctx.HTML(200, "error")
	}

}

/*
 AlipaySubmitOrder impletation
*/
func AlipaySubmitOrder(ctx *macaron.Context, x *xorm.Engine, client alipay.Client, user *model.User, order *model.StockOrder) {
	// var typeName = "周"
	// if order.ProductType == int8(1) {
	// 	typeName = "月"
	// }
	subject := "subscribe"
	opts := alipay.Options{
		OrderId:  strings.Replace(order.OutTradeNo, "-", "", -1), // 唯一订单号
		Fee:      float32(order.OrderAmount),                     // 价格
		NickName: user.UserName,                                  // 用户昵称，支付页面显示用
		Subject:  subject,                                        // 支付描述，支付页面显示用
	}
	fmt.Println("------pay with alipay ----------")

	form := client.Form(opts)
	fmt.Println(form)
	ctx.Data["form"] = template.HTML(form)
	ctx.HTML(200, "alipay")
}

/*
AlipayFinishHandler sync callback
*/
func AlipayFinishHandler(ctx *macaron.Context, x *xorm.Engine, r *redis.Client, client alipay.Client) {
	result := client.NativeReturn(ctx.Req.Request)
	fmt.Println("alipay result:", result)
	flag := "fail"
	if result.Status == 1 { //付款成功，处理订单
		//处理订单
		OrderPayed(x.NewSession(), r, result.OrderNo, model.AliPay)
		flag = "ok"
	}
	jsonResult := new(model.JsonResult)
	jsonResult.Code = "200"
	jsonResult.Data = flag
	ctx.JSON(200, jsonResult)
}

/*
AlipayNotifyHandler async callback
*/
func AlipayNotifyHandler(ctx *macaron.Context, x *xorm.Engine, r *redis.Client, client alipay.Client) {
	result := client.NativeNotify(ctx.Req.Request)
	fmt.Println("alipay result:", result)
	flag := "fail"
	if result.Status == 1 { //付款成功，处理订单
		//处理订单
		OrderPayed(x.NewSession(), r, result.OrderNo, model.AliPay)
		flag = "ok"
	}
	jsonResult := new(model.JsonResult)
	jsonResult.Code = "200"
	jsonResult.Data = flag
	ctx.JSON(200, jsonResult)
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

/*
 OrderPayed 订单支付成功处理
*/
func OrderPayed(s *xorm.Session, r *redis.Client, OutTradeNo string, payType string) (bool, string) {

	order := new(model.StockOrder)
	if has, _ := s.Where("out_trade_no=?", OutTradeNo).Get(order); has {
		//判断是否已经支付
		if order.OrderStatus != 0 {
			return false, "订单状态异常，请检查订单"
		}

		now := time.Now()
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
		_, err := s.ID(order.Id).Update(order)
		if err != nil {
			log.Printf("出现异常%s", err.Error())
			s.Rollback()
			return false, "订单更新失败,请联系客服"
		}

		user := new(model.User)         //订阅人
		followedUser := new(model.User) //被订阅人
		s.ID(order.UserId).Get(user)
		s.ID(order.UserId).Get(followedUser)

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
	return false, "更新订单失败"
}
