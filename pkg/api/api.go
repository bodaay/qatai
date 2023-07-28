package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"qatai/pkg/db"
	"qatai/pkg/models"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

func StartGeneartionServer(addr string, WebFS http.FileSystem, mydb db.QataiDatabase, httplogger *zap.Logger) error {

	assetHandler := http.FileServer(WebFS)

	e := echo.New()

	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		// AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogURI:    true,
		LogStatus: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			httplogger.Info("request",
				zap.String("URI", v.URI),
				zap.Int("status", v.Status),
			)

			return nil
		},
	}))
	e.POST("/v1/chat/completions", chatCompletionHandler(mydb, httplogger))
	// e.File("/chat/*", "/index.html") // this is still not working
	e.GET("/*", echo.WrapHandler(assetHandler))
	return e.Start(addr)

}

func chatCompletionHandler(mydb db.QataiDatabase, logger *zap.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("X-Accel-Buffering", "no")
		c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
		c.Response().Header().Set("Transfer-Encoding", "chunked")

		var uniReq models.UniversalRequest

		err := json.NewDecoder(c.Request().Body).Decode(&uniReq)
		if err != nil {
			return err
		}

		llmmodel, err := db.GetModelByName(mydb, "gpt-4-0613")
		if err != nil {
			return err
		}
		// llmmodel.Stops = []string{"</s>"}
		events := make(chan string, 100)
		go models.DoGenerate(&uniReq, llmmodel, events)
		timeoutInterval := 10
		timeout := time.After(time.Duration(timeoutInterval) * time.Second)
		GeneratedTokens := 0
	Loop:
		for {
			select {
			case ev := <-events:
				n, err := fmt.Fprintf(c.Response(), "data: %v\n\n", ev)
				if err != nil {
					logger.Error(err.Error())
					continue
				}
				_ = n
				GeneratedTokens += 1
				// fmt.Printf("written for buffer: %d, with data: %v\n", n, ev)

				c.Response().Flush()
				if ev == "[DONE]" {
					logger.Info(fmt.Sprintf("Total Number of Tokens Generated: %d", GeneratedTokens-1))
					_, _ = fmt.Fprintf(c.Response(), ":\n\n") // Apparently this forces close SSE
					close(events)
					break Loop
				}

			case <-timeout:
				_, _ = fmt.Fprintf(c.Response(), ":\n\n") // Apparently this forces close SSE
				logger.Debug(fmt.Sprintf("Timed out while waiting for generation to completed, Timeout: %d seconds", timeoutInterval))
				// close(events) //TODO: //I still dont know whats the best approach for this, how do I time out
				c.Response().Flush()
				break Loop
			}
		}

		return nil
	}
}
