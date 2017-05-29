package handler

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"../model"

	"fmt"

	"strconv"

	"github.com/go-macaron/captcha"
	"github.com/go-macaron/session"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

//hangers
//登录跳转
func LoginGetHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {

	if sess.Get("user") != nil {
		user := sess.Get("user").(model.User)
		ctx.HTML(200, "index", user)
	} else {
		ctx.HTML(200, "login")
	}
}

//登录
func LoginPostHandler(user model.User, sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger, cpt *captcha.Captcha) {
	//验证码验证

	if !cpt.VerifyReq(ctx.Req) {
		ctx.HTML(200, "login")
	}

	//登录
	hash := md5.New()
	hash.Write([]byte(user.Password)) // 需要加密的字符串为 123456
	cipherStr := hash.Sum(nil)
	log.Printf("%s,%s", user.UserName, hex.EncodeToString(cipherStr))

	u := new(model.User)
	if has, err := x.Where("user_name=? and password=? ", user.UserName, hex.EncodeToString(cipherStr)).Get(u); has {
		//写入session
		u.LoginCount++
		u.LastLoginDate = time.Now()
		sess.Set("user", u)
		x.Update(u)
		ctx.HTML(200, "user", u)
	} else {
		log.Printf("没有找到用户。。。%s", err)
		ctx.HTML(200, "register")
	}
}

//注销
func LogoutHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {

	if sess.Get("user") != nil {
		user := sess.Get("user").(model.User)
		log.Printf("用户%s注销", user.UserName)
		sess.Delete("user")
		ctx.HTML(200, "login")
	} else {
		ctx.HTML(200, "index")
	}
}

//注册跳转
func RegisterGetHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger, cpt *captcha.Captcha) {
	sess.Delete("user")
	ctx.HTML(200, "register")
}

//注册保存
func RegisterPostHandler(user model.User, sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger, cpt *captcha.Captcha) {

	//检验
	if !cpt.VerifyReq(ctx.Req) {
		ctx.HTMLSet(200, "register", "mike")
	}

	//检查用户名和手机号是否存在
	if ok, err := registerPreCheck(&user, x); !ok {
		log.Printf("err:", err)
		ctx.HTML(200, "register")
	} else {
		user.UserRole = model.CUSTOMER
		user.UserStatus = model.USER_STATUS_OK
		id, err := x.Insert(&user)
		if id > 0 {
			log.Printf("用户%s注册成功", user.UserName)
			userAccount := initUserAccount(&user)
			_, err := x.Insert(userAccount)
			Chk(err)
			//注册完毕后的动作--->
			go afterRegisterHandler(&user, x, log)
			//<-------------end
			log.Printf("用户%s注册结束", user.UserName)
		} else {
			log.Printf("注册失败:%s", err)
			ctx.HTML(200, "register")
		}
	}
}

/**
初始化账户表
**/
func initUserAccount(user *model.User) *model.UserAccount {

	//生成一个userAccount
	userAccount := new(model.UserAccount)
	userAccount.UserId = user.Id
	userAccount.UserName = user.UserName
	userAccount.InitAmount = model.INIT_AMOUNT
	userAccount.AvailableAmount = model.INIT_AMOUNT
	userAccount.LockAmount = 0
	userAccount.Gold = 0
	userAccount.Integral = 0
	userAccount.EarningRate = 0
	userAccount.TransFrequency = 0
	userAccount.SuccessRate = 0
	userAccount.UserLevel = 1
	return userAccount

}

/**
异步操作注册成功动作
**/
func afterRegisterHandler(user *model.User, x *xorm.Engine, log *log.Logger) {
	log.Println("------email confirm----")
}

//跳转用户详情
func UserDetailHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger, cpt *captcha.Captcha) {

	//1、根据用户ID获取用户信息
	//2、判断产讯用户是否是本人

	if id, _ := strconv.Atoi(ctx.Params(":id")); id > 0 {
		user := new(model.User)
		ctx.Data["user"] = user
		ctx.HTML(200, "user_center")
	}
}

//跳转我的账户
func UserAccountHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger, cpt *captcha.Captcha) {

	//1、根据用户ID获取用户信息
	//2、判断查询用户是否是本人

	if id, _ := strconv.Atoi(ctx.Params(":id")); id > 0 {
		user := new(model.User)
		x.Id(id).Get(user)
		ctx.Data["user"] = user
		ctx.HTML(200, "user_account")
	}
}

/**
用户资料更新跳转
**/
func UserUpdateGetHandler(sess session.Store, ctx *macaron.Context, log *log.Logger) {
	//用户资料更新
	ctx.HTML(200, "user_update", sess.Get("user").(model.User))

}

/**
用户资料更新保存
**/
func UserUpdateSaveHandler(user model.User, sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger) {
	//用户资料更新
	log.Printf("更新用户资料....%s", user.UserName)
	x.Id(user.Id).Cols("true_name,sex,mobile,address,birthday,id_card,email,nick_name,updated").Update(&user)
	x.Id(user.Id).Find(&user)
	ctx.HTML(200, "user_center", user)
	log.Printf("更新用户资料结束....%s", user.UserName)
}

/**
注册检查
**/
func registerPreCheck(user *model.User, x *xorm.Engine) (bool, error) {

	if count, _ := x.Where("user_name=? ", user.UserName).Count(new(model.User)); count > 0 {
		return false, errors.New(fmt.Sprintf("用户名:%s已经存在", user.Mobile))
	}
	if count, _ := x.Where("mobile=? ", user.Mobile).Count(new(model.User)); count > 0 {
		return false, errors.New(fmt.Sprintf("手机号:%s已经存在", user.Mobile))
	}
	return true, nil
}

/**
根据用户ID获取用户
**/
func QueryUserByIdWithEngine(x *xorm.Engine, id interface{}) (bool, *model.User) {

	user := new(model.User)
	if has, _ := x.Id(id).Get(user); has {
		return user == nil, user

	} else {
		return false, nil
	}
}

/**
根据用户ID获取用户的资金账户信息
**/
func QueryUserAccoutByUserIdWithEngine(x *xorm.Engine, userId int64) (bool, *model.UserAccount) {

	userAccount := new(model.UserAccount)
	flag, _ := x.Where("user_id=?", userId).Get(userAccount)
	return flag, userAccount
}

/**
根据用户ID获取用户
**/
func QueryUserByIdWithSession(x *xorm.Session, id interface{}) (bool, *model.User) {

	user := new(model.User)
	if has, _ := x.Id(id).Get(user); has {
		return user == nil, user

	} else {
		return false, nil
	}
}

/**
根据用户ID获取用户的资金账户信息
**/
func QueryUserAccoutByUserIdWithSession(x *xorm.Session, userId int64) (bool, *model.UserAccount) {

	userAccount := new(model.UserAccount)
	flag, _ := x.Where("user_id=?", userId).Get(userAccount)
	return flag, userAccount
}

/**
查询用户是否用订阅用户
**/
func hasFollowers(s *xorm.Session, log *log.Logger, followedId int64) bool {

	len, err := s.Where("followed_id=?", followedId).Count(new(model.UserFollow))
	if err == nil {
		return len > 0
	} else {
		log.Printf("查询异常：%s", err.Error())
		return false
	}
}

/**
查询所有订阅者
**/
func listMyFollowers(s *xorm.Session, log *log.Logger, followedId int64) (bool, []*model.UserFollow) {

	userFollows := make([]*model.UserFollow, 0)
	err := s.Where("followed_id=?", followedId).Find(&userFollows)
	if err == nil {
		return true, userFollows
	} else {
		log.Printf("查询异常：%s", err.Error())
		return false, nil
	}
}
