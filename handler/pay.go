package handler

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"encoding/xml"
	"fmt"
	"log"
	"net/url"
	"regexp"
	"strings"
	"time"

	"../model"

	"github.com/go-xorm/xorm"
	"github.com/stephenlyu/wxpay"
	macaron "gopkg.in/macaron.v1"
)

//
// ************************AliPay************************
//

const ALIPAY_PARTNER = "2088812266839292"
const ALIPAY_PRIVATE_KEY = "MIICdgIBADANBgkqhkiG9w0BAQEFAASCAmAwggJcAgEAAoGBANRz4SVco9Wx9h7D" +
	"2dsfiGjqE14+iljL/xRjZdqP/38CnWndSJxwNs0nKJX+8x4jU3/Pd9/kcKwjXu67" +
	"oa2niHGNpF1eXkLxePstG82MHctCv7BcWxvAWpCqvG6yNQIb1uBdbMVjnR2eMHlm" +
	"THwMXxaeIs3w8fJX0OB5gI2GE3ajAgMBAAECgYB7kwual9AUNHdcXb8SXb0SiVTK" +
	"tMXz8HRmf4p3HtsWHYdCVJwvonW9ztEkri7rkNC4vwyTBmUjO0+0vR7Fy3Tox2gJ" +
	"pioCWOMHixd4EzyJnKxYNPNMsd9lompRcChKKAWnWsvbJP3HHget2qUmxXcMmU/c" +
	"MTOc2lrd1pKxhrT8wQJBAP/2TxCVbG9CVgSxFued6944eyBF+hBqh8IRfCOlJ2JR" +
	"auqejqiOI9ppU9Z7AgZmWYALnw7vvXszCnRsVoxXgHUCQQDUe+xclow9u0obQsZC" +
	"JtFO/4T0OoNg4WzFiClW/6QHzM9T1pvV6jM5muenpuq4kYTdbemEQnmJPLMkgSsl" +
	"h7e3AkBuH2Z82AzDAWNIuXgFRmhIPzyZ8gFYNr0ZvbQPEesT3buGHZl640yBl3c+" +
	"e8WvQzGWaWmRX4vCCX+h/0ptLuhRAkAxZsJ0YFgovhOjtOmtVaMST9wUgEotSxvj" +
	"7R1XacY0Pgzx/BJtMK9KNFappugpk0Oly7kgE+h33NH1qcZjSmOPAkEA5wzfB3w/" +
	"hugXkbjnnYJ7APR2n0F5xdkF2cfigiuMTdMjkpyXqmUAiSfB08AyxMONbGeuRN2L" +
	"3R6DQha7iXbT/g=="
const ALIPAY_PUBLIC_KEY = `-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQCnxj/9qwVfgoUh/y2W89L6BkRAFljhNhgPdyPuBV64bfQNN1PjbCzkIM6qRdKBoLPXmKKMiFYnkd6rAoprih3/PrQEB/VsW8OoM8fxn67UDYuyBTqA23MML9q1+ilIZwBC2AQ2UBVOrFXfFl75p6/B5KsiNG9zpgmLCUYuLkxpLQIDAQAB
-----END PUBLIC KEY-----
`

type Result struct {
	// 状态
	Status int
	// 本网站订单号
	OrderNo string
	// 支付宝交易号
	TradeNo string
	// 买家支付宝账号
	BuyerEmail string
	// 错误提示
	Message string
}

func alipayVerifySign(data string, sign string) bool {
	// Parse public key into rsa.PublicKey
	PEMBlock, _ := pem.Decode([]byte(ALIPAY_PUBLIC_KEY))
	if PEMBlock == nil {
		log.Print("Could not parse Public Key PEM")
		return false
	}
	if PEMBlock.Type != "PUBLIC KEY" {
		log.Print("Found wrong key type")
		return false
	}
	pubkey, err := x509.ParsePKIXPublicKey(PEMBlock.Bytes)
	if err != nil {
		log.Print(err)
		return false
	}

	// compute the sha1
	h := sha1.New()
	h.Write([]byte(data))

	// Read the signature from stdin
	b64, err := base64.StdEncoding.DecodeString(sign)
	if err != nil {
		log.Print(err)
		return false
	}

	// Verify
	err = rsa.VerifyPKCS1v15(pubkey.(*rsa.PublicKey), crypto.SHA1, h.Sum(nil), b64)
	if err != nil {
		log.Print(err)
		return false
	}

	return true
}

/* 被动接收支付宝异步通知 */
func AlipayNotifyHandler(ctx *macaron.Context, x *xorm.Engine, log *log.Logger) (int, string) {

	// /pay/notify/104_alipay?discount=0.00&payment_type=1&subject=%E5%9C%A8%E7%BA%BF%E6%94%AF%E4%BB%98%E8%AE%A2%E5%8D%95&trade_no=2015072800001000810060741985&buyer_email=***&gmt_create=2015-07-28%2001:24:19%C2%ACify_type=trade_status_sync&quantity=1&out_trade_no=146842585&seller_id=2088021187655650%C2%ACify_time=2015-07-28%2001:24:29&body=%E8%AE%A2%E5%8D%95%E5%8F%B7%EF%BC%9A146842585&trade_status=TRADE_SUCCESS&is_total_fee_adjust=N&total_fee=0.01&gmt_payment=2015-07-28%2001:24:29&seller_email=***&price=0.01&buyer_id=2088302384317810%C2%ACify_id=75e570fcc802c637d8cf1fdaa8677d046i&use_coupon=N&sign_type=MD5&sign=***
	//old body method
	// body, _ := ioutil.ReadAll(r.Body)
	bodyStr, _ := ctx.Req.Body().String()
	log.Println("alipay 支付回调：新的获取body方式：")

	if bodyStr == "" {
		return 200, "fail"
	}

	log.Printf("[alipay] body: %s", bodyStr)

	//从body里读取参数，用&切割
	postArray := strings.Split(bodyStr, "&")

	//实例化url
	urls := &url.Values{}

	//保存传参的sign
	var paramSign string
	var sign string

	//如果字符串中包含sec_id说明是手机端的异步通知
	if strings.Index(bodyStr, `alipay.wap.trade.create.direct`) == -1 { //快捷支付
		for _, v := range postArray {
			detail := strings.Split(v, "=")

			//使用=切割字符串 去除sign和sign_type
			if detail[0] == "sign" || detail[0] == "sign_type" {
				if detail[0] == "sign" {
					paramSign = detail[1]
				}
				continue
			} else {
				urls.Add(detail[0], detail[1])
			}
		}

		// url解码
		urlDecode, _ := url.QueryUnescape(urls.Encode())
		sign, _ = url.QueryUnescape(urlDecode)
		paramSign, _ = url.QueryUnescape(paramSign)
	} else { // 手机网页支付
		// 手机字符串加密顺序
		mobileOrder := []string{"service", "v", "sec_id", "notify_data"}
		for _, v := range mobileOrder {
			for _, value := range postArray {
				detail := strings.Split(value, "=")
				// 保存sign
				if detail[0] == "sign" {
					paramSign = detail[1]
				} else {
					// 如果满足当前v
					if detail[0] == v {
						if sign == "" {
							sign = detail[0] + "=" + detail[1]
						} else {
							sign += "&" + detail[0] + "=" + detail[1]
						}
					}
				}
			}
		}
		sign, _ = url.QueryUnescape(sign)
		paramSign, _ = url.QueryUnescape(paramSign)

		//获取<trade_status></trade_status>之间的request_token
		re, _ := regexp.Compile("\\<trade_status[\\S\\s]+?\\</trade_status>")
		rt := re.FindAllString(sign, 1)
		trade_status := strings.Replace(rt[0], "<trade_status>", "", -1)
		trade_status = strings.Replace(trade_status, "</trade_status>", "", -1)
		urls.Add("trade_status", trade_status)

		//获取<out_trade_no></out_trade_no>之间的request_token
		re, _ = regexp.Compile("\\<out_trade_no[\\S\\s]+?\\</out_trade_no>")
		rt = re.FindAllString(sign, 1)
		out_trade_no := strings.Replace(rt[0], "<out_trade_no>", "", -1)
		out_trade_no = strings.Replace(out_trade_no, "</out_trade_no>", "", -1)
		urls.Add("out_trade_no", out_trade_no)

		//获取<buyer_email></buyer_email>之间的request_token
		re, _ = regexp.Compile("\\<buyer_email[\\S\\s]+?\\</buyer_email>")
		rt = re.FindAllString(sign, 1)
		buyer_email := strings.Replace(rt[0], "<buyer_email>", "", -1)
		buyer_email = strings.Replace(buyer_email, "</buyer_email>", "", -1)
		urls.Add("buyer_email", buyer_email)

		//获取<trade_no></trade_no>之间的request_token
		re, _ = regexp.Compile("\\<trade_no[\\S\\s]+?\\</trade_no>")
		rt = re.FindAllString(sign, 1)
		trade_no := strings.Replace(rt[0], "<trade_no>", "", -1)
		trade_no = strings.Replace(trade_no, "</trade_no>", "", -1)
		urls.Add("trade_no", trade_no)
	}

	log.Printf("[alipay] sign: %s\n", sign)
	log.Printf("[alipay] paramSign: %s\n", paramSign)

	orderNo := urls.Get("out_trade_no")
	tradeNo := urls.Get("trade_no")

	if alipayVerifySign(sign, paramSign) { //传进的签名等于计算出的签名，说明请求合法
		log.Println("[alipay] sign verify success")
		//判断订单是否已完成
		if urls.Get("trade_status") == "TRADE_FINISHED" || urls.Get("trade_status") == "TRADE_SUCCESS" { //交易成功
			err := UpdateOrderPayStatus(x, log, model.ORDER_STATUS_NOT_PAYED, orderNo, "alipay")
			if err != nil {
				log.Print("error: update order status, order id: %s trade no: %s", orderNo, tradeNo)
			}

			return 200, "success"
		}
	}
	log.Println("[alipay] sign verify fail")

	return 200, "fail"
}

//
// *********************************微信支付*********************************
//

const WX_APP_KEY = "2ec3dfca00f46cce2c7e64e6cc71e70b"

const SERVER_ADDR string = "localhost：8080"
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
		err := UpdateOrderPayStatus(x, log, model.ORDER_STATUS_NOT_PAYED, notiData.OutTradeNo, "weixin")
		if err != nil {
			log.Print("error: update order status, order id: %s trade no: %s", notiData.OutTradeNo, notiData.TransactionId)
		}
	}

	result := &WxNotifyResult{ReturnCode: "SUCCESS", ReturnMsg: "OK"}
	s, err := xml.Marshal(result)
	Chk(err)
	return 200, string(s)
}
