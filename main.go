package main

import (
	"github.com/dilshat/sms-sender/controller"
	"github.com/dilshat/sms-sender/dao"
	_ "github.com/dilshat/sms-sender/docs"
	"github.com/dilshat/sms-sender/log"
	"github.com/dilshat/sms-sender/service"
	"github.com/dilshat/sms-sender/sms"
	"github.com/dilshat/sms-sender/util"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Sms service HTTP API
// @description Simple sms service

// @contact.name Dilshat Aliev
// @contact.email dilshat.aliev@gmail.com

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Error.Println(err)
	}
}

func main() {
	//create db client
	dbClient, err := dao.GetClient(util.GetEnv("DB_PATH", "sms.db"))
	if err != nil {
		log.Fatal(err)
	}

	//create smpp client
	smppClient := sms.NewClient(util.GetEnv("SMS_IP", ""),
		util.GetEnvAsInt("SMS_PORT", 8018),
		util.GetEnv("SMS_ID", ""),
		util.GetEnv("SMS_PWD", ""),
		util.GetEnvAsInt("ENQ_LNK_SEC", 30),
		util.GetEnvAsInt("TRX_PER_SEC", 100))

	smsSender := sms.NewSender(smppClient)

	//start sms sender
	err = smsSender.Start()
	if err != nil {
		log.Fatal(err)
	}

	smsService := service.NewService(
		smsSender,
		dao.NewMessageDao(dbClient),
		dao.NewRecipientDao(dbClient),
		util.GetEnvAsInt("STATUS_STORE_DAYS", 7),
		util.GetEnvAsInt("SMS_MAX_LEN", 300),
		util.GetEnv("WEB_HOOK", ""),
		util.GetEnv("PHONE_MASK", "996\\d{9}"),
	)

	//attach http handlers
	e := echo.New()
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	e.HideBanner = true
	e.Use(middleware.BodyLimit("2K"))

	bindRoutes(e, smsService)

	//start http server
	log.Fatal(e.Start(":" + util.GetEnv("HTTP_PORT", "8080")))
}

func bindRoutes(e *echo.Echo, service service.Service) {

	e.POST("/sms", controller.GetSendSmsFunc(service))

	e.GET("/sms/:id", controller.GetCheckSmsFunc(service))
}
