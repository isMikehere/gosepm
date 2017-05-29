package handler

import (
	"log"

	"../model"

	"github.com/go-xorm/xorm"
)

/**
我订阅成功后，则消息通知我
**/
func SendMessage(x *xorm.Engine, log *log.Logger, mobile, content string) {

	messageLog := new(model.MessageLog)
	messageLog.Mobile = mobile
	messageLog.Content = content
	if ok := sendMessage(mobile, content); ok {
		messageLog.SendStatus = 1
	} else {
		messageLog.SendStatus = 0
	}
	_, err := x.Insert(messageLog)
	Chk(err)
}

/**
发送短信等外置接口
**/
func sendMessage(mobile, content string) bool {

	log.Println("---------短信发送：%s,%s", mobile, content)
	return true
}

/**
发送短信等外置接口
**/
func batchSendMessage(mobile []string, content string) bool {
	return true
}
