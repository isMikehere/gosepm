package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"time"

	"./db"

	"./handler"
	"./model"
	"github.com/go-macaron/binding"
	"github.com/go-macaron/cache"
	"github.com/go-macaron/captcha"
	"github.com/go-macaron/session"
	"github.com/go-xorm/core"
	"github.com/go-xorm/xorm"
	macaron "gopkg.in/macaron.v1"
	redis "gopkg.in/redis.v2"
)

var engine *xorm.Engine
var myLogger *log.Logger
var redisCli *redis.Client
var m *macaron.Macaron

func ConfigEngine() {
	//打印sql
	engine.ShowSQL(true)
	//映射类型
	// engine.SetMapper(core.SameMapper{})
	//连接池
	engine.SetMaxIdleConns(10)
	engine.SetMaxOpenConns(20)
	//缓存
	cacher := xorm.NewLRUCacher(xorm.NewMemoryStore(), 1000)
	engine.SetDefaultCacher(cacher)
	//日志级别
	engine.Logger().SetLevel(core.LOG_DEBUG)
}

func SyncTable(tableNames ...interface{}) {

	fmt.Printf("开始同步表 \n")
	for _, table := range tableNames {
		ok, _ := engine.IsTableExist(table)
		if !ok {
			engine.CreateTables(table)
		}
		engine.Sync(table)
	}
	fmt.Printf("开始同步表结束\n")
}

func TestQuery() {
	postStocks := make([]model.PostStock, 0)
	engine.NoCache().Table("sz_post_stock_2010").Limit(1, 0).Find(&postStocks)
	fmt.Printf("%s", postStocks)
}

/**
数据库同步
*/
func DbSync() {
	ConfigEngine()
	SyncTable(new(model.News))
	// TestQuery()
}

/**
启动server
**/
func webgo() {
	//2、启动模板引擎

	m.Use(macaron.Renderer())
	// m.Use(pongo2.Pongoer())
	//session存储内存

	m.Use(session.Sessioner())

	// m.Use(session.Sessioner(session.Options{
	// 	Provider: "redis",
	// 	// e.g.: network=tcp,addr=127.0.0.1:6379,password=macaron,db=0,pool_size=100,idle_timeout=180,prefix=session:
	// 	ProviderConfig: "addr=127.0.0.1:6379,password=xceof",
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

		log.Println("mike-----before a request--" + ctx.Req.RequestURI)

		ctx.Next()

		log.Println("mike-----before a request--" + ctx.Req.RequestURI)
	})

	/*------------------------routes-------------------------------------------*/

	m.Get("/login.htm", handler.LoginGetHandler)
	m.Post("/login.htm", binding.Bind(model.User{}), handler.LoginPostHandler)
	m.Post("/logout.htm", handler.LogoutHandler)
	m.Get("/logout.htm", handler.LogoutHandler)
	m.Get("/register.htm", handler.RegisterGetHandler)
	m.Post("/register.htm", binding.Bind(model.User{}), handler.RegisterPostHandler)
	m.Get("/test.htm", indexHandler)

	//用户
	m.Group("/user", func() {
		m.Get("/:id", handler.UserDetailHandler)
		m.Get("/update/:id", handler.UserUpdateGetHandler)
		m.Put("/update/:id", binding.Bind(model.User{}), handler.UserUpdateSaveHandler)
		m.Post("/update/:id", binding.Bind(model.User{}), handler.UserUpdateSaveHandler)
		m.Group("/account", func() {
			m.Get("/:id", handler.UserAccountHandler)
		})
		m.Group("/follow", func() {
			m.Get("/:id", handler.FollowStep1Handler)
			m.Post("/:id", handler.FollowStep2Handler)
			m.Post("/:id/:type", handler.FollowStep2Handler)
		})
	})

	//交易相关
	m.Group("/trx", func() {
		m.Get("", func(ctx *macaron.Context) {
			ctx.JSON(200, "")
		})
		m.Post("/cancel/:entId", handler.CancelEntrustHandler) //撤单
	})
	//支付相关
	m.Post("/alipay/notify", handler.AlipayNotifyHandler)

	// m.Group("/weixin", func() {
	// 	m.Post("/order/(?P<id>[0-9a-z]{24})", midOrder, hanlder.WxCreatePrepayOrder)
	// 	m.Get("/result", hanlder.WxQueryPayResult)
	// 	m.Post("/notify", hanlder.WxNotifyHandler)
	// 	m.Get("/callback", hanlder.weixinCallback)
	// })

	//----------------------对外接口----------------------------------------------/

	/*------------------------routes-------------------------------------------*/
	m.Run()
}

func indexHandler(ctx *macaron.Context) {
	// user := new(model.User)
	// user.NickName = "mikke"
	// stocks := make([]*model.Stock, 2)
	// stock := new(model.Stock)
	// stock2 := new(model.Stock)
	// stock.StockCode = "00001"
	// stock.StockName = "中国平安"
	// stock2.StockCode = "00002"
	// stock2.StockName = "中国移动"
	// stocks[0] = stock
	// stocks[1] = stock2
	// var data = &model.Data{Id: 1, Name: "Mike", User: user, Stocks: stocks}
	ctx.Data["name"] = "mike"
	ctx.Data["age"] = 10
	ctx.HTML(200, "test")
}

func initXormEngin() {
	// //映射数据库服务
	eg, err := db.ConnectDb(model.DriverOfMysql, model.DataSourceOfMysql)
	if err != nil {
		fmt.Println("connection db err,%s", err)
		panic(err)
	} else {
		engine = eg
		log.Println("db-client init ok")
	}
	m.Map(engine)
}

/*
**/
func initRedisClient() {

	//映射redis
	fmt.Println("new redis-client... ")
	redisCli = NewRedisClient()
	if err == nil {
		m.Map(redisCli)
		fmt.Println("redis-client init ok")
	} else {
		fmt.Println("redis-client init fail")
		panic(err)
	}
}

func initLogger() {
	//映射logger
	var buf bytes.Buffer
	myLogger = log.New(&buf, "logger: ", log.Lshortfile)
	// m.Map(myLogger)
	log.Println("-----------mike......")
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
				ReSet()
			}
		}

	}
	log.Print("------缓存用户结束---------")
	log.Print("------缓存股票---------")
	//初始化股票
	stocks := make([]*model.Stock, 0)
	err = engine.Where("1=1").Find(&stocks)
	if err != nil {
		panic(0)
	} else {
		if len(stocks) > 0 {
			for _, s := range stocks {
				SetRedisStock(s)
			}
		}
	}
	log.Print("------缓存股票结束---------")
}
func main() {

	m = macaron.Classic()
	initLogger()
	initXormEngin()
	//数据库同步
	SyncTable(new(model.User))
	initRedisClient()
	initCache()
	webgo()

	//开启定时任务
	// job := new(model.MyJob)
	// c := new(cron.Cron)
	// job.AddToRun(c, "job", "@every 1s")
}
