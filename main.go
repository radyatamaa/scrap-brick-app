package main

import (
	"github.com/tebeka/selenium"
	"net/http"
	"strings"
	"time"

	"github.com/radyatamaa/scrap-brick-app/internal"
	"github.com/radyatamaa/scrap-brick-app/pkg/database"

	beego "github.com/beego/beego/v2/server/web"
	beegoContext "github.com/beego/beego/v2/server/web/context"
	"github.com/beego/beego/v2/server/web/filter/cors"
	"github.com/beego/i18n"
	"github.com/radyatamaa/scrap-brick-app/internal/domain"
	"github.com/radyatamaa/scrap-brick-app/internal/middlewares"
	"github.com/radyatamaa/scrap-brick-app/pkg/response"
	"github.com/radyatamaa/scrap-brick-app/pkg/zaplogger"

	productHandler "github.com/radyatamaa/scrap-brick-app/internal/product/delivery/http/v1"
	productRepository "github.com/radyatamaa/scrap-brick-app/internal/product/repository"
	productUsecase "github.com/radyatamaa/scrap-brick-app/internal/product/usecase"
)

// @title Api Gateway V1
// @version v1
// @contact.name radyatama
// @contact.email mohradyatama24@gmail.com
// @description api "API Gateway v1"
// @BasePath /api
// @query.collection.format multi

func main() {
	err := beego.LoadAppConfig("ini", "conf/app.ini")
	if err != nil {
		panic(err)
	}
	// global execution timeout
	serverTimeout := beego.AppConfig.DefaultInt64("serverTimeout", 60)
	// global execution timeout
	requestTimeout := beego.AppConfig.DefaultInt("executionTimeout", 5)
	// web hook to slack error log
	slackWebHookUrl := beego.AppConfig.DefaultString("slackWebhookUrlLog", "")
	// app version
	appVersion := beego.AppConfig.DefaultString("version", "1")
	// log path
	logPath := beego.AppConfig.DefaultString("logPath", "./logs/api.log")

	// database initialization
	db := database.DB()

	// language
	lang := beego.AppConfig.DefaultString("lang", "en|id")
	languages := strings.Split(lang, "|")
	for _, value := range languages {
		if err := i18n.SetMessage(value, "./conf/"+value+".ini"); err != nil {
			panic("Failed to set message file for l10n")
		}
	}

	// initialize a Chrome browser instance on port 4444
	service, err := selenium.NewChromeDriverService("./chromedriver", 4444)

	if err != nil {
		panic(err)
	}

	defer service.Stop()

	// global execution timeout to second
	timeoutContext := time.Duration(requestTimeout) * time.Second

	// beego config
	beego.BConfig.Log.AccessLogs = false
	beego.BConfig.Log.EnableStaticLogs = false
	beego.BConfig.Listen.ServerTimeOut = serverTimeout

	// zap logger
	zapLog := zaplogger.NewZapLogger(logPath, slackWebHookUrl)

	if beego.BConfig.RunMode == "dev" {
		// db auto migrate dev environment
		if err := db.AutoMigrate(
			&domain.Product{},
		); err != nil {
			panic(err)
		}

		// static files swagger
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	if beego.BConfig.RunMode != "prod" {
		// static files swagger
		beego.BConfig.WebConfig.DirectoryIndex = true
		beego.BConfig.WebConfig.StaticDir["/swagger"] = "swagger"
	}

	// middleware init
	beego.InsertFilter("*", beego.BeforeRouter, cors.Allow(&cors.Options{
		AllowMethods:    []string{http.MethodGet, http.MethodPost},
		AllowAllOrigins: true,
	}))

	beego.InsertFilterChain("*", middlewares.RequestID())
	beego.InsertFilterChain("/api/*", middlewares.BodyDumpWithConfig(middlewares.NewAccessLogMiddleware(zapLog, appVersion).Logger()))

	// health check
	beego.Get("/health", func(ctx *beegoContext.Context) {
		ctx.Output.SetStatus(http.StatusOK)
		ctx.Output.JSON(
			beego.M{
				"status":  "alive",
				"version": beego.AppConfig.DefaultString("version", "1.1.0"),
			}, beego.BConfig.RunMode != "prod", false)
	})

	// default error handler
	beego.ErrorController(&response.ErrorController{})

	// init repository
	productRepo := productRepository.NewPgProductRepository(db, zapLog)

	// init usecase
	productUcase := productUsecase.NewProductUseCase(timeoutContext,
		productRepo,
		zapLog)

	// init handler
	productHandler.NewProductHandler(productUcase, zapLog)

	// default error handler
	beego.ErrorController(&internal.BaseController{})

	beego.Run()
}
