package api

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"qatai/pkg/db"
	"qatai/pkg/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/valyala/fasthttp"
)

type Client struct {
	name   string
	events chan string
}
type DashBoard struct {
	User uint
}

var mydb db.QataiDatabase

func StartGeneartionServer(WebFS http.FileSystem, mydb db.QataiDatabase) {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://gpu01.yawal.io:3000/, http://gpu01.yawal.io",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(c.App().Stack())
	})
	app.Post("/v1/chat/completions", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")
		var uniReq models.UniversalRequest

		err := json.Unmarshal(c.Request().Body(), &uniReq)
		if err != nil {
			log.Fatalln(err)
		}
		c.Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {

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
					n, err := fmt.Fprintf(w, "data: %v\n\n", ev)
					if err != nil {
						log.Println(err)
					}

					fmt.Printf("written for buffer: %d, with data: %v\n", n, ev)

					err = w.Flush()
					if err != nil {
						// Refreshing page in web browser will establish a new
						// SSE connection, but only (the last) one is alive, so
						// dead connections must be closed here.
						fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

						break Loop
					}

				case <-timeout:
					_, _ = fmt.Fprintf(w, ":\n\n") // Apparently this forces close SSE

					err = w.Flush()
					if err != nil {
						// Refreshing page in web browser will establish a new
						// SSE connection, but only (the last) one is alive, so
						// dead connections must be closed here.
						fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

						break
					}

					break Loop
				}
			}

		}))

		return nil
	})
	// app.Post("/v1/chat/completions", adaptor.HTTPHandler(handler(generationHandler)))
	// always keep this one last, otherwise my get request didn't work
	app.Use(filesystem.New(filesystem.Config{
		Root:         WebFS,
		Browse:       false,
		Index:        "index.html",
		NotFoundFile: "404.html",
		MaxAge:       3600,
	}))

	app.Listen(":5050")
}
