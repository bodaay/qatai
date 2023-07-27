package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qatai/pkg/db"
	"qatai/pkg/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/labstack/echo/v4"
)

func StartGeneartionServer(WebFS http.FileSystem, mydb db.QataiDatabase) {
	app := fiber.New(
		fiber.Config{
			// EnablePrintRoutes: true,
			StrictRouting: true,
			AppName:       "QatAI",
		},
	)
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "*",
	}))
	// Initialize default config
	app.Use(logger.New())

	// Or extend your config for customization
	// Logging remote IP and Port
	app.Use(logger.New(logger.Config{
		Format: "[${ip}]:${port} ${status} - ${method} ${path}\n",
	}))

	// Logging Request ID
	app.Use(requestid.New())
	app.Use(logger.New(logger.Config{
		// For more options, see the Config section
		Format: "${pid} ${locals:requestid} ${status} - ${method} ${path}​\n",
	}))

	// Changing TimeZone & TimeFormat
	app.Use(logger.New(logger.Config{
		Format:     "${pid} ${status} - ${method} ${path}\n",
		TimeFormat: "02-Jan-2006",
		TimeZone:   "America/New_York",
	}))

	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(c.App().Stack())
	})

	assetHandler := http.FileServer(WebFS)

	e := echo.New()

	e.POST("/v1/chat/completions", sseHandler(mydb))
	e.GET("/*", echo.WrapHandler(assetHandler))
	e.Start(":5050")

}

func sseHandler(mydb db.QataiDatabase) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Response().Header().Set(echo.HeaderContentType, "text/event-stream")
		// c.Response().Header().Set(echo.HeaderXAccelBuffering, "no")
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
