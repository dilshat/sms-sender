package main

import (
	"log"

	"github.com/dilshat/sms-sender/controller"
	"github.com/dilshat/sms-sender/dao"
	_ "github.com/dilshat/sms-sender/docs"
	"github.com/dilshat/sms-sender/service"
	"github.com/dilshat/sms-sender/sms"
	"github.com/dilshat/sms-sender/util"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// @title Sms service HTTP API
// @description Simple sms service

// @contact.name Dilshat Aliev
// @contact.email dilshat.aliev@gmail.com

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading env variables", err)
	}
	initZapLog()
	if err != nil {
		log.Fatal("Error initializing logger", err)
	}
}

func initZapLog() error {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	//set log level
	logLevel := zap.InfoLevel
	logLevel.Set(util.GetEnv("LOG_LEVEL", "ERROR"))
	config.Level.SetLevel(logLevel)
	logger, err := config.Build()
	if err != nil {
		return err
	}
	zap.ReplaceGlobals(logger)
	return nil
}

func main() {
	defer zap.L().Sync() // flushes buffer, if any

	//create db client
	dbClient, err := dao.GetClient(util.GetEnv("DB_PATH", "sms.db"))
	if err != nil {
		zap.L().Fatal("Error connecting to db", zap.Error(err))
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
		zap.L().Fatal("Error connecting to SMSC", zap.Error(err))
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
	err = e.Start(":" + util.GetEnv("HTTP_PORT", "8080"))
	zap.L().Fatal("Error starting http server", zap.Error(err))
}

func bindRoutes(e *echo.Echo, service service.Service) {

	e.POST("/sms", controller.GetSendSmsFunc(service))

	e.GET("/sms/:id", controller.GetCheckSmsFunc(service))
}
