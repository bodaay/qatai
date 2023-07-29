package api

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"qatai/pkg/db"
	"qatai/pkg/models"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/pocketbase/pocketbase"
	"github.com/pocketbase/pocketbase/apis"
	"github.com/pocketbase/pocketbase/core"
	"go.uber.org/zap"
)

func StartGeneartionServer(addr string, WebFS fs.FS, mydb db.QataiDatabase, httplogger *zap.Logger) error {
	// I love pocketbase <3
	app := pocketbase.New()
	app.OnBeforeServe().Add(func(e *core.ServeEvent) error {
		e.Router.GET("/*", apis.StaticDirectoryHandler(WebFS, false))

		e.Router.POST("/v1/chat/completions", chatCompletionHandler(mydb, httplogger))
		return nil
	})
	app.RootCmd.SetArgs([]string{"serve", "--http=0.0.0.0:5050", "--origins=*"})
	if err := app.Start(); err != nil {
		log.Fatal(err)
	}
	for {
		time.Sleep(1 * time.Second)
	}
	return nil

}

func chatCompletionHandler(mydb db.QataiDatabase, logger *zap.Logger) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		c.Response().Header().Set("X-Accel-Buffering", "no")
		// c.Response().Header().Set(echo.HeaderCacheControl, "no-cache")
		// c.Response().Header().Set(echo.HeaderConnection, "keep-alive")
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
