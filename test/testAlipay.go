package test

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/ascoders/alipay"
)

func testInitForm() string {
	param := &alipay.AlipayParameters{}
	param.InputCharset = "utf-8"
	param.Body = "为mike充值0.01元"
	param.NotifyUrl = "http://www.goldgad.com/pay/alipay/notify"
	param.OutTradeNo = "8"
	param.Partner = "2088721921014090"
	param.PaymentType = 1
	param.ReturnUrl = "http://www.goldgad.com/pay/alipay/finish"
	param.SellerEmail = "jinxiuwl@aliyun.com"
	param.Service = "create_direct_pay_by_user"
	param.Subject = "100"
	param.TotalFee = 0.01

	//生成签名
	sign := sign(param)

	return sign

}

// 按照支付宝规则生成sign
func sign(param interface{}) string {
	//解析为字节数组
	paramBytes, err := json.Marshal(param)
	if err != nil {
		return ""
	}

	//重组字符串
	var sign string
	oldString := string(paramBytes)

	//为保证签名前特殊字符串没有被转码，这里解码一次
	oldString = strings.Replace(oldString, `\u003c`, "<", -1)
	oldString = strings.Replace(oldString, `\u003e`, ">", -1)

	//去除特殊标点
	oldString = strings.Replace(oldString, "\"", "", -1)
	oldString = strings.Replace(oldString, "{", "", -1)
	oldString = strings.Replace(oldString, "}", "", -1)
	paramArray := strings.Split(oldString, ",")

	for _, v := range paramArray {
		detail := strings.SplitN(v, ":", 2)
		//排除sign和sign_type
		if detail[0] != "sign" && detail[0] != "sign_type" {
			//total_fee转化为2位小数
			if detail[0] == "total_fee" {
				number, _ := strconv.ParseFloat(detail[1], 32)
				detail[1] = strconv.FormatFloat(number, 'f', 2, 64)
			}
			if sign == "" {
				sign = detail[0] + "=" + detail[1]
			} else {
				sign += "&" + detail[0] + "=" + detail[1]
			}
		}
	}

	//追加密钥
	sign += "rw2si4ejhwhw4nymm2fnvhlg34gtaxk5"

	fmt.Println("before:" + sign)
	//md5加密
	m := md5.New()
	m.Write([]byte(sign))
	sign = hex.EncodeToString(m.Sum(nil))
	return sign
}
