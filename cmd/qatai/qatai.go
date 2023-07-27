package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"

	"qatai/pkg/api"
	"qatai/pkg/db"

	"github.com/labstack/gommon/log"
	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
)

var version = "1.0.0"

//  you can use the below, to embed your whole website content into the binary itself, Thanks to: https://v0x.nl/articles/portable-apps-go-nextjs/ , https://github.com/dstotijn/hetty

var qataiUsage = `
Usage:
    qatai [flags] [subcommand] [flags]

Runs an HTTP server with (MITM) proxy, GraphQL service, and a web based admin interface.

Options:
    --cert         Path to root CA certificate. Creates file if it doesn't exist. (Default: "~/.qatai/qatai_cert.pem")
    --key          Path to root CA private key. Creates file if it doesn't exist. (Default: "~/.qatai/qatai_key.pem")
    --db           Database directory path. (Default: "~/.qatai/db")
    --addr         TCP address for HTTP server to listen on, in the form \"host:port\". (Default: ":8080")
    --chrome       Launch Chrome with proxy settings applied and certificate errors ignored. (Default: false)
    --verbose      Enable verbose logging.
    --json         Encode logs as JSON, instead of pretty/human readable output.
    --version, -v  Output version.
    --help, -h     Output this usage text.

Subcommands:
    - cert  Certificate management

Run ` + "`qatai <subcommand> --help`" + ` for subcommand specific usage instructions.


`

type QataiCommand struct {
	config    *Config
	bboltPath string
	useMongo  bool
	mongoHost string
	cert      string
	key       string
	db        string
	addr      string
	version   bool
}

func NewqataiCommand() (*ffcli.Command, *Config) {
	cmd := QataiCommand{
		config: &Config{},
	}

	fs := flag.NewFlagSet("qatai", flag.ExitOnError)

	// fs.StringVar(&cmd.cert, "cert", "~/.qatai/qatai_cert.pem",
	// 	"Path to root CA certificate. Creates a new certificate if file doesn't exist.")
	// fs.StringVar(&cmd.key, "key", "~/.qatai/qatai_key.pem",
	// 	"Path to root CA private key. Creates a new private key if file doesn't exist.")
	fs.StringVar(&cmd.bboltPath, "bbolt_path", "~/.qatai/db", "Database directory path.")
	fs.BoolVar(&cmd.useMongo, "use_mongo", false, "use mongo db instead of bbolt")
	fs.StringVar(&cmd.mongoHost, "mongo_host", "mongodb://localhost:27017", "mongo db connection string")
	fs.StringVar(&cmd.db, "db", "~/.qatai/db", "Database directory path.")
	fs.StringVar(&cmd.addr, "addr", ":5050", "TCP address to listen on, in the form \"host:port\".")
	fs.BoolVar(&cmd.version, "version", false, "Output version.")
	fs.BoolVar(&cmd.version, "v", false, "Output version.")

	cmd.config.RegisterFlags(fs)

	return &ffcli.Command{
		Name:    "qatai",
		FlagSet: fs,
		// Subcommands: []*ffcli.Command{ //you can add more subcommand here, Thanks Hetty <3. https://github.com/dstotijn/hetty/
		// 	NewCertCommand(cmd.config),
		// },
		Exec: cmd.Exec,
		UsageFunc: func(*ffcli.Command) string {
			return qataiUsage
		},
	}, cmd.config
}

func (cmd *QataiCommand) Exec(ctx context.Context, _ []string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt)
	defer stop()

	if cmd.version {
		fmt.Fprint(os.Stdout, version+"\n")
		return nil
	}

	mainLogger := cmd.config.logger.Named("main")

	listenHost, listenPort, err := net.SplitHostPort(cmd.addr)
	if err != nil {
		mainLogger.Fatal("Failed to parse listening address.", zap.Error(err))
	}

	url := fmt.Sprintf("http://%v:%v", listenHost, listenPort)
	if listenHost == "" || listenHost == "0.0.0.0" || listenHost == "127.0.0.1" || listenHost == "::1" {
		url = fmt.Sprintf("http://localhost:%v", listenPort)
	}
	//db

	//webserver
	// e := echo.New()
	// e.HideBanner = true
	go func() {
		mainLogger.Info(fmt.Sprintf("qatai (v%v) is running on %v ...", version, cmd.addr))
		mainLogger.Info(fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(32), "Get started at "+url))
		//start the Generation API
		mdb, err := db.InitNewMongoDB("mongodb://localhost:27017", "qatai")
		if err != nil {
			panic(err)
		}
		os.MkdirAll("data", os.ModePerm)
		bboltDbPath := path.Join("data", "bbolt.db")
		bdb, err := db.InitNewBoltDB(bboltDbPath)
		if err != nil {
			panic(err)
		}
		TestDB(mdb, bdb)
		err = api.StartGeneartionServer(cmd.addr, getFileSystem(false, cmd.config.logger.Named("http")), bdb, cmd.config.logger.Named("http"))

		if err != http.ErrServerClosed {
			mainLogger.Fatal("HTTP server closed unexpected.", zap.Error(err))
		}

	}()

	// Wait for interrupt signal.
	<-ctx.Done()
	// Restore signal, allowing "force quit".
	stop()

	mainLogger.Info("Shutting down HTTP server. Press Ctrl+C to force quit.")

	// Note: We expect httpServer.Handler to handle timeouts, thus, we don't
	// need a context value with deadline here.
	//nolint:contextcheck
	// err = e.Shutdown(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	// }

	return nil
}

func TestDB(mdb db.QataiDatabase, bdb db.QataiDatabase) {

	db.ClearAllConfig(mdb)
	db.ClearAllConfig(bdb)
	db.SetConfig(mdb, &db.Config{Key: "TestKey", Value: "TestValue"})
	db.SetConfig(mdb, &db.Config{Key: "TestKey2", Value: "TestValue2"})
	db.SetConfig(bdb, &db.Config{Key: "TestKey", Value: "TestValue"})
	db.SetConfig(bdb, &db.Config{Key: "TestKey2", Value: "TestValue2"})
	fmt.Println(db.GetConfig(mdb, "TestKey"))
	fmt.Println(db.GetConfig(bdb, "TestKey"))

	fmt.Println(db.GetAllConfig(mdb))
	fmt.Println(db.GetAllConfig(bdb))

	//Test Creating models:
	db.ClearAllModels(mdb)
	db.ClearAllModels(bdb)

	endpoints := []db.LLMEndPoint{
		{Host: "gpu01.yawal.io", Port: 8080, UseSSL: false},
		// {Host: "localhost", Port: 8081, UseSSL: false},
	}

	prompts := []db.LLMPrompts{
		{Title: "Test title", Prompt: "Test prompt", PromptImage: "Test Image"},
	}

	params := db.LLMParameters{
		Temperature:        0.95,
		Top_P:              0.9,
		Top_K:              40,
		RepetitionPenality: 1.0,
		Truncate:           2048,
		MaxNewTokens:       1024,
	}
	tokens := db.LLMTokens{
		SystemToken:    "<<sys>>",
		UserToken:      "[INST]",
		AssistantToken: "[/INST]",
		FunctionToken:  "",
	}
	model := db.NewLLMModel("gpt-4-0613", "LLaMa V2 13B parameters", db.HFTGI, "<<SYS>>\n You are a helpful, respectful and honest assistant. <</SYS>>", tokens, []string{"</s>"}, endpoints, prompts, params)

	if err := db.AddUpdateModel(mdb, model, false); err != nil {
		log.Errorf("Failed to add/update model: %s", err.Error())
	}
	if err := db.AddUpdateModel(bdb, model, false); err != nil {
		log.Errorf("Failed to add/update model: %s", err.Error())
	}
	fmt.Println(db.GetAllModels(mdb))
	fmt.Println(db.GetAllModels(bdb))

	// uReq := &models.UniversalRequest{
	// 	Messages: []models.Message{
	// 		{
	// 			Role:    "assistant", //this later we have to pull it from the config of the LLM Model
	// 			Content: "How can I help you today?",
	// 		},
	// 		{
	// 			Role:    "user",
	// 			Content: "what is the biggest country in the world?",
	// 		},
	// 	},
	// 	Stream:           true,
	// 	Model:            "gpt-4",
	// 	Temperature:      1,
	// 	TopP:             0.95,
	// 	Stop:             []string{"</s>"},
	// 	N:                1,
	// 	PresencePenalty:  0,
	// 	FrequencyPenalty: 1.2,
	// }
	// models.DoGenerate(uReq, model, nil)
}
