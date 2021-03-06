package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	"github.com/sirupsen/logrus"
	"gopkg.in/kataras/iris.v6"
	"gopkg.in/kataras/iris.v6/adaptors/httprouter"
	"gopkg.in/kataras/iris.v6/adaptors/sessions"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:  %s [options] ...\n", os.Args[0])
	flag.PrintDefaults()
}

var (
	// Flags
	helpShort  = flag.Bool("h", false, "Show usage text (same as --help).")
	helpLong   = flag.Bool("help", false, "Show usage text (same as -h).")
	serverIp   = flag.String("ip", "127.0.0.1", "the server ip")
	serverPort = flag.Int("p", 7200, "the server port")
	configFile = flag.String("c", "config.json", "config file")
	debug      = flag.Bool("debug", false, "debug log")
	srvConfig  Config
)

var DB *gorm.DB

func main() {

	flag.Usage = usage
	flag.Parse()
	if *helpShort || *helpLong {
		flag.Usage()
		return
	}

	if err := srvConfig.Load(*configFile); err != nil {
		logrus.Error(err.Error())
		panic(err.Error())
	}
	//	log.InitLog(config.ServerConfig.Name)
	logrus.Debugf("%v", srvConfig)

	app := iris.New(iris.Configuration{
		Gzip:    true,
		Charset: "UTF-8",
	})

	app.Adapt(iris.DevLogger())

	app.Adapt(sessions.New(sessions.Config{
		Cookie:  srvConfig.Server.SessionID,
		Expires: time.Minute * 20,
	}))

	app.Adapt(httprouter.New())

	Route(app)

	app.OnError(iris.StatusNotFound, func(ctx *iris.Context) {
		ctx.JSON(iris.StatusOK, iris.Map{
			"errNo": 4003,
			"msg":   "Not Found",
			"data":  iris.Map{},
		})
	})

	app.OnError(500, func(ctx *iris.Context) {
		ctx.JSON(iris.StatusInternalServerError, iris.Map{
			"errNo": 5000,
			"msg":   "error",
			"data":  iris.Map{},
		})
	})

	app.Listen(":" + strconv.Itoa(*serverPort))
}
