package main

import (
	"bytes"
	"flag"
	"fmt"
	"html/template"
	"log"
	"runtime"
	"time"

	"./db"
	"./handler"
	"./model"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/captcha"
	"github.com/go-macaron/session"
	_ "github.com/go-macaron/session/redis"
	redis "github.com/go-redis/redis"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
)

var engine *xorm.Engine
var myLogger *log.Logger
var redisCli *redis.Client
var m *macaron.Macaron

func ConfigEngine(x *xorm.Engine) {
	//打印sql
	x.ShowSQL(true)
	//映射类型
	x.SetMapper(core.SnakeMapper{})
	// engine.SetTableMapper(core.SameMapper{})

	//连接池
	x.SetMaxIdleConns(10)
	x.SetMaxOpenConns(20)
	//缓存
	// cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	// engine.SetDefaultCacher(cacher)
	//日志级别
	x.Logger().SetLevel(core.LOG_INFO)
}

func SyncTable(tableNames ...interface{}) {

	fmt.Printf("开始同步表 \n")
	for _, table := range tableNames {
		ok, _ := engine.IsTableExist(table)
		if !ok {
			engine.CreateTables(table)
		} else {
			fmt.Println("表已经存在")
		}
		engine.Sync(table)
	}
	fmt.Printf("开始同步表结束\n")
}

func initXormEngin() {
	// //映射数据库服务
	eg, err := db.ConnectDb(model.DriverOfMysql, model.DataSourceOfMysql)
	if err != nil {
		fmt.Println("connection db err,%s", err)
		panic(err)
	} else {
		engine = eg
		ConfigEngine(engine)
		log.Println("db-client init ok")
	}
	m.Map(engine)
}

/*
**/
func initRedisClient() {

	//映射redis
	fmt.Println("new redis-client... ")
	redisCli = handler.NewRedisClient()
	redisCli.Set("name", "mike1", 0)
	m.Map(redisCli)
	fmt.Printf("redis-client init ok%s", (redisCli == nil))
}

func initLogger() {
	//映射logger
	var buf bytes.Buffer
	myLogger = log.New(&buf, "logger: ", log.Lshortfile)
	// m.Map(myLogger)
}

/**
初始化服务
**/
func initCache() {

	//初始化用户
	log.Println("------缓存用户---------")
	users := make([]*model.User, 0)
	err := engine.Where("1=1").Find(&users)
	if err != nil {
		log.Fatalf("%s", err.Error())
		panic(0)
	} else {
		if len(users) > 0 {
			for _, u := range users {
				handler.SetRedisUser(redisCli, u)
			}
		}

	}
	log.Print("------缓存用户结束---------")
	log.Print("------缓存股票---------")
	//初始化股票
	stocks := make([]*model.Stock, 0)
	err = engine.Where("1=1").Find(&stocks)
	if err != nil {
		log.Printf("%s", err)
		panic(0)
	} else {
		if len(stocks) > 0 {
			stockCodes := make([]string, len(stocks))
			i := 0
			for _, s := range stocks {
				handler.SetRedisStock(redisCli, s)
				stockCodes[i] = s.Location + s.StockCode
				i++
			}
			handler.SetRedisStockCodes(redisCli, stockCodes)
		}
	}
	log.Print("------缓存股票结束---------")
	return
}

/**
启动server
**/
func webgo() {
	//2、启动模板引擎

	// m.Use(macaron.Renderer())
	// m.Use(pongo2.Pongoer())
	//session存储内存

	m.Use(session.Sessioner())

	// m.Use(session.Sessioner(session.Options{
	// 	Provider: "redis",
	// 	// e.g.: network=tcp,addr=model.RedisHost,password=macaron,db=0,pool_size=100,idle_timeout=180,prefix=session:
	// 	ProviderConfig: "addr=" + model.RedisHost + ",password=xceof",
	// }))

	//验证码验证
	m.Use(cache.Cacher())
	m.Use(captcha.Captchaer())
	//模版引擎配置
	m.Use(macaron.Renderer(macaron.RenderOptions{
		// 模板文件目录，默认为 "templates"
		Directory: "templates",
		// 模板文件后缀，默认为 [".tmpl", ".html"]
		Extensions: []string{".tmpl", ".html"},
		// 模板函数，默认为 []
		Funcs: []template.FuncMap{map[string]interface{}{
			"AppName": func() string {
				return "Macaron"
			},
			"AppVer": func() string {
				return "1.0.0"
			},
			"GetUserNickName": handler.GetUserNickName,
			"FormatDate":      handler.FormatDate,
			"FormatDateTime":  handler.FormatDateTime,
		}},
		// 模板语法分隔符，默认为 ["{{", "}}"]
		Delims: macaron.Delims{"{{", "}}"},
		// 追加的 Content-Type 头信息，默认为 "UTF-8"
		Charset: "UTF-8",
		// 渲染具有缩进格式的 JSON，默认为不缩进
		IndentJSON: true,
		// 渲染具有缩进格式的 XML，默认为不缩进
		IndentXML: true,
		// 渲染具有前缀的 JSON，默认为无前缀
		PrefixJSON: []byte("macaron"),
		// 渲染具有前缀的 XML，默认为无前缀
		PrefixXML: []byte("macaron"),
		// 允许输出格式为 XHTML 而不是 HTML，默认为 "text/html"
		HTMLContentType: "text/html",
	}))
	//静态资源配置
	m.Use(macaron.Static("public",
		macaron.StaticOptions{
			// 请求静态资源时的 URL 前缀，默认没有前缀
			Prefix: "public",
			// 禁止记录静态资源路由日志，默认为不禁止记录
			SkipLogging: true,
			// 当请求目录时的默认索引文件，默认为 "index.html"
			IndexFile: "index.html",
			// 用于返回自定义过期响应头，默认为不设置
			// https://developers.google.com/speed/docs/insights/LeverageBrowserCaching
			Expires: func() string {
				return time.Now().Add(24 * 60 * time.Minute).UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT")
			},
		}))

	// filter login status before and after a request

	m.Use(func(sess session.Store, ctx *macaron.Context, log *log.Logger) {

		log.Println("-----before a request--" + ctx.Req.RequestURI)
		ctx.Data["webpath"] = ctx.Req.Host
		if sess.Get("user") == nil {
			u := new(model.User)
			engine.Id(2).Get(u)
			sess.Set("user", u)
		}

		// url := ctx.Req.RequestURI
		// if strings.Contains(url, "/login") || strings.Contains(url, "index.htm") || strings.Contains(url, ".js") || strings.Contains(url, ".css") {
		// 	ctx.Next()
		// } else {
		// 	if sess.Get("user") == nil {
		// 		ctx.Redirect("/login.htm")
		// 	} else {
		// 		ctx.Next()
		// 	}
		// }
		ctx.Next()
		log.Println("-----before a request--" + ctx.Req.RequestURI)
	})

	/*------------------------routes-------------------------------------------*/

	m.Get("/login.htm", handler.LoginGetHandler)
	m.Post("/login.htm", binding.Bind(model.User{}), handler.LoginPostHandler)
	m.Post("/logout.htm", handler.LogoutHandler)
	m.Get("/logout.htm", handler.LogoutHandler)
	m.Get("/register.htm", handler.RegisterGetHandler)
	m.Post("/register.htm", binding.Bind(model.User{}), handler.RegisterPostHandler)
	m.Get("/index.htm", handler.IndexHandler)
	m.Get("/error.htm", func(ctx *macaron.Context) {
		ctx.HTML(200, "error")
	})
	m.Get("/success.htm", func(ctx *macaron.Context) {
		ctx.HTML(200, "success")
	})

	//用户
	m.Group("/user", func() {
		m.Get("/:id", handler.UserDetailHandler)
		m.Get("/update/:id", handler.UserUpdateGetHandler)
		m.Put("/update/:id", binding.Bind(model.User{}), handler.UserUpdateSaveHandler)
		m.Post("/update/:id", binding.Bind(model.User{}), handler.UserUpdateSaveHandler)
		m.Get("/search/:name", handler.SearchXUserHandler)
		m.Group("/account", func() {
			m.Get("/:id", handler.UserAccountHandler)
		})
		m.Group("/follow", func() {
			m.Get("/:id", handler.FollowStep1Handler)
			m.Post("/:id/:type", handler.FollowStep2Handler) //下单---》进入收银台
			m.Post("/:id/:type", handler.FollowStep2Handler) //下单---》进入收银台
		})

		//持仓
		m.Group("/holding", func() {
			m.Get("", handler.MyHoldingHandler)        //我的持仓
			m.Get("/:page", handler.MyPageableHandler) //分页持仓
		})
		//委托
		m.Group("/entrust", func() {
			m.Get("", handler.TodayEntrustHandler)       //当日委托
			m.Get("/:page", handler.TodayEntrustHandler) //当日成交
		})
		m.Group("/msg", func() {
			// m.Get("/:msgKey", handle.LatestMsgHandler) //查看消息

		})

	})
	//follow
	m.Get("/myFollow/:followStatus", handler.UserFollowListHandler)
	m.Get("/myOrder/:orderStatus", handler.OrderListHandler)

	//产品相关
	m.Group("/product", func() {
		m.Get("/:type", handler.GetProductHandler)
	})

	//交易相关
	m.Group("/trx", func() {
		m.Get("", handler.TrxHandler)
		m.Post("/cancel/:entId", handler.CancelEntrustHandler) //撤单
	})

	//股票基础数据
	m.Get("/stock5/:stockCode", handler.Stock5StageHander)

	//支付相关
	if macaron.Env == macaron.DEV {
		m.Post("/pay/:orderId", handler.TestPayHandler)
	} else {
		m.Post("/pay/:orderId", handler.AlipayNotifyHandler)
	}
	m.Post("/alipay/notify", handler.AlipayNotifyHandler)

	// m.Group("/weixin", func() {
	// 	m.Post("/order/(?P<id>[0-9a-z]{24})", midOrder, hanlder.WxCreatePrepayOrder)
	// 	m.Get("/result", hanlder.WxQueryPayResult)
	// 	m.Post("/notify", hanlder.WxNotifyHandler)
	// 	m.Get("/callback", hanlder.weixinCallback)
	// })

	//排名
	m.Get("/rank/:page", handler.RankListHandler)

	m.Get("/test.htm", func(ctx *macaron.Context, x *xorm.Engine, r *redis.Client) {
		user := new(model.User)
		user.NickName = "sophie"
		x.Insert(user)
		ctx.JSON(200, user)
	})

	//----------------------对外接口----------------------------------------------/

	/*------------------------routes-------------------------------------------*/
	m.Run()
}

func main() {

	m = macaron.Classic()
	// ConfigEngine()
	initXormEngin()
	//数据库同步

	SyncTable(new(model.NotifyFollow), new(model.StockTrxInfo),
		new(model.MessageLog))
	// initRedisClient()
	//·
	var numCores = flag.Int("n", 2, "number of CPU cores to use")
	flag.Parse()
	runtime.GOMAXPROCS(*numCores)
	// initCache()
	//开启定时任务
	// go handler.StartSchedule(handler.InitScheduleJobs(engine, redisCli))
	// webgo()
}
