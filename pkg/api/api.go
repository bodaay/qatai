package api

import (
	"encoding/json"
	"fmt"
	"log"
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
	e.POST("/v1/chat/completions", chatCompletionHandler(mydb))
	e.GET("/*", echo.WrapHandler(assetHandler))
	return e.Start(addr)

}

func chatCompletionHandler(mydb db.QataiDatabase) echo.HandlerFunc {
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
			log.Fatalln(err)
		}
		// llmmodel.Stops = []string{"</s>"}
		events := make(chan string, 100)
		go models.DoGenerate(&uniReq, llmmodel, events)
		timeout := time.After(10 * time.Second)
	Loop:
		for {
			select {
			case ev := <-events:
				n, err := fmt.Fprintf(c.Response(), "data: %v\n\n", ev)
				if err != nil {
					log.Println(err)
				}

				fmt.Printf("written for buffer: %d, with data: %v\n", n, ev)

				c.Response().Flush()
				if ev == "[DONE]" {
					_, _ = fmt.Fprintf(c.Response(), ":\n\n") // Apparently this forces close SSE
					close(events)
					break Loop
				}

			case <-timeout:
				_, _ = fmt.Fprintf(c.Response(), ":\n\n") // Apparently this forces close SSE
				close(events)
				c.Response().Flush()
				break Loop
			}
		}

		return nil
	}
}
