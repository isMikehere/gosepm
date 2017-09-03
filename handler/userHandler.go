package handler

import (
	"crypto/md5"
	"encoding/hex"
	"log"
	"reflect"
	"time"

	"../model"

	"fmt"

	"strconv"

	"github.com/go-macaron/captcha"
	"github.com/go-macaron/session"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

//hangers
//登录跳转
func LoginGetHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine) {

	if sess.Get("user") != nil {
		user := sess.Get("user").(*model.User)
		ctx.Data["user"] = user
		ctx.HTML(200, "index")
	} else {
		ctx.HTML(200, "login")
	}
}

//登录
func LoginPostHandler(user model.User, sess session.Store, ctx *macaron.Context, x *xorm.Engine, r *redis.Client, cpt *captcha.Captcha) {
	//验证码验证

	if !cpt.VerifyReq(ctx.Req) {
		ctx.Data["msg"] = "验证码错误"
		ctx.HTML(200, "login")
	}

	//登录
	hash := md5.New()
	hash.Write([]byte(user.Password)) // 需要加密的字符串为 123456
	cipherStr := hash.Sum(nil)
	log.Printf("%s,%s", user.UserName, hex.EncodeToString(cipherStr))

	u := new(model.User)
	if has, _ := x.Where("user_name=? and password=? ", user.UserName, hex.EncodeToString(cipherStr)).Get(u); has {
		//写入session
		u.LoginCount++
		u.LastLoginDate = time.Now()
		sess.Set("user", u)
		x.ID(u.Id).Update(u)
		// ctx.Redirect("/index.htm")
		IndexHandler(sess, ctx, x, r)
	} else {
		log.Println("用户名密码不正确")
		ctx.Data["msg"] = "用户名密码不正确"
		ctx.HTML(200, "login")
	}
}

//注销
func LogoutHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine) {

	if sess.Get("user") != nil {
		user := sess.Get("user").(*model.User)
		log.Printf("用户%s注销", user.UserName)
		sess.Delete("user")
		ctx.HTML(200, "login")
	} else {
		ctx.HTML(200, "index")
	}
}

/**
获取手机验证码
**/
func GetMobileCode(ctx *macaron.Context, x *xorm.Engine, r *redis.Client) {

	jr := new(model.JsonResult)
	//检查是否允许发短信
	mobile := ctx.Params(":mobile")
	if mobile != "" && checkSend(mobile) {
		code := RandomIntCode()
		fmt.Printf("生成的短信验证码：%s", code)
		r.Set(mobile, code, model.MSG_EXPIRE_DURATION) //expire
		expired := time.Now().Add(model.MSG_EXPIRE_DURATION)
		f, _ := sendMessage(mobile, fmt.Sprintf(model.REGISTER_MSG, code))
		if f {
			vc := new(model.VerifyCode)
			vc.Mobile = mobile
			vc.Code = code
			vc.Expired = expired
			x.Insert(vc)
			jr.Code = "200"
		} else {
			jr.Code = "100"
			jr.Msg = "发送短信失败"
		}
	}
	ctx.JSON(200, jr)
}
func checkSend(mobile string) bool {
	return true
}

//注册跳转
func RegisterGetHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, cpt *captcha.Captcha) {
	sess.Delete("user")
	ctx.HTML(200, "register")
}

//注册保存
func RegisterPostHandler(user model.User, sess session.Store, ctx *macaron.Context, x *xorm.Engine, redis *redis.Client, cpt *captcha.Captcha) {

	//检验
	mobileCode := ctx.Query("mobileCode")
	fmt.Printf("手机验证码:mobileCode:%s", mobileCode)
	if ok := checkMobileCode(redis, user.Mobile, mobileCode); !ok {
		ctx.Data["msg"] = "验证码有误"
		ctx.HTML(200, "register")
		return
	}

	//检查用户名和手机号是否存在
	if ok, err := registerPreCheck(&user, x); !ok {
		log.Printf("err:%s", err)
		ctx.Data["msg"] = err
		ctx.HTML(200, "register")
	} else {
		user.UserRole = model.CUSTOMER
		user.UserStatus = model.USER_STATUS_OK
		user.Password = Md5(user.Password)
		id, err := x.Insert(&user)
		if id > 0 {
			log.Printf("用户%s注册成功", user.UserName)
			userAccount := initUserAccount(&user)
			if _, err := x.Insert(userAccount); err == nil {
				//注册完毕后的动作--->
				go afterRegisterHandler(&user, x)
				//<-------------end
				log.Printf("用户%s注册结束", user.UserName)
				ctx.HTML(200, "register_success")
			} else {
				log.Printf("用户%s注册失败", user.UserName)
				ctx.HTML(200, "register")
			}

		} else {
			ctx.Data["msg"] = "注册失败！"
			log.Printf("注册失败:%s", err)
			ctx.HTML(200, "register")
		}
	}
}

/**
mobile code checkk
**/
func checkMobileCode(r *redis.Client, mobile, mobileCode string) bool {

	if code, _ := r.Get(mobile).Result(); code == mobileCode {
		return true
	}
	return false
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
func afterRegisterHandler(user *model.User, x *xorm.Engine) {
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

//跳转我的
func UserAccountHandler(sess session.Store, ctx *macaron.Context, x *xorm.Engine, log *log.Logger, cpt *captcha.Captcha) {

	//1、根据用户ID获取用户信息
	//2、判断查询用户是否是本人
	log.Printf("get user account %s")
	id := 0
	user := new(model.User)
	if id, _ = strconv.Atoi(ctx.Params(":id")); id > 0 {
		ctx.Data["self"] = false
		if has, _ := x.Id(id).Get(&user); !has { //get the user again
			ctx.HTML(200, "notfound")
			return
		}
	} else {
		ctx.Data["self"] = true
		loginUser := sess.Get("user")
		v := reflect.ValueOf(loginUser)
		user = v.Interface().(*model.User)
	}
	ctx.Data["user"] = user

	userAccount := new(model.UserAccount)
	//get useraccount
	if err := x.Where("user_id=?", user.Id).Find(&userAccount); err != nil { //ua
		ctx.Data["userAccount"] = userAccount
		// get the ranking data
		weekRank := new(model.WeekRank)
		if has, _ := x.Where("user_id=?", id).Desc("id").Limit(1, 0).Get(&weekRank); has {
			ctx.Data["weekRank"] = weekRank
		} else {
			weekRank.Rank = userAccount.Rank
			weekRank.EarningRate = userAccount.EarningRate
		}
		//get the month rank
		monthRank := new(model.MonthRank)
		if has, _ := x.Where("user_id=?", id).Desc("id").Limit(1, 0).Get(&monthRank); has {
			ctx.Data["monthRank"] = monthRank
		} else {
			monthRank.Rank = userAccount.Rank
			monthRank.EarningRate = userAccount.EarningRate
		}
		ctx.HTML(200, "user_account")
	} else {
		ctx.HTML(200, "notfound")
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
func registerPreCheck(user *model.User, x *xorm.Engine) (bool, string) {

	if count, _ := x.Where("user_name=? ", user.UserName).Count(new(model.User)); count > 0 {
		return false, fmt.Sprintf("用户名:%s已经存在", user.Mobile)
	}
	if count, _ := x.Where("mobile=? ", user.Mobile).Count(new(model.User)); count > 0 {
		return false, fmt.Sprintf("手机号:%s已经存在", user.Mobile)
	}
	fmt.Printf("%s,%d", user.Password, len(user.Password))
	if len(user.Password) < 8 {
		return false, "密码长度不够"
	}
	return true, ""
}

/**
根据用户ID获取用户
**/
func QueryUserByIdWithEngine(x *xorm.Engine, id interface{}) (bool, *model.User) {

	user := new(model.User)
	if has, _ := x.Id(id).Get(user); has {
		return user != nil, user

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
func hasFollowers(s *xorm.Session, followedId int64) bool {

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
func listMyFollowers(s *xorm.Session, followedId int64) (bool, []*model.UserFollow) {

	userFollows := make([]*model.UserFollow, 0)
	err := s.Where("followed_id=?", followedId).Find(&userFollows)
	if err == nil {
		return true, userFollows
	} else {
		log.Printf("查询异常：%s", err.Error())
		return false, nil
	}
}

/**
查找高手
**/
func SearchXUserHandler(ctx *macaron.Context, x *xorm.Engine) {
	name := ctx.Params("name")
	users := make([]*model.User, 0)
	log.Printf("%s", name)
	if err := x.Where("nick_name like ?", "%"+name+"%").Limit(5, 0).Find(&users); err == nil {
		log.Print(users)
		ctx.JSON(200, users)
	} else {
		ctx.JSON(200, nil)
	}
}

/**
统计一个用户对订阅量
**/
func countFollowNumbers(s *xorm.Session, uid int64) int64 {
	c1, _ := s.Where("followed_id=?", uid).Count(new(model.UserFollow))
	return c1
}

func GetUserById(x interface{}, s *redis.Client, id int64) *model.User {

	sid := strconv.Itoa(int(id))
	if user := GetRedisUser(s, sid); user == nil {
		var user *model.User
		switch t := x.(type) {
		case *xorm.Engine:
			{
				user = new(model.User)
				if has, _ := t.Id(id).Get(user); has {
					return user
				} else {
					return nil
				}
			}
		case *xorm.Session:
			{
				user = new(model.User)
				if has, _ := t.Id(id).Get(user); has {
					return user
				} else {
					return nil
				}
			}
		default:
			return nil
		}
	} else {
		return user
	}

}
