package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/peterbourgon/ff/v3/ffcli"
	"go.uber.org/zap"
)

var version = "1.0.0"

//go:embed web
//go:embed web/_next/static
//go:embed web/_next/static/chunks/pages/*.js
//go:embed web/_next/static/*/*.js
var webContent embed.FS

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

Visit https://qatai.xyz to learn more about qatai.
`

type QataiCommand struct {
	config *Config

	cert    string
	key     string
	db      string
	addr    string
	version bool
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
	fs.StringVar(&cmd.db, "db", "~/.qatai/db", "Database directory path.")
	fs.StringVar(&cmd.addr, "addr", ":8080", "TCP address to listen on, in the form \"host:port\".")
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

	go func() {
		mainLogger.Info(fmt.Sprintf("Hetty (v%v) is running on %v ...", version, cmd.addr))
		mainLogger.Info(fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(32), "Get started at "+url))

		//you main blocking routine here
		for {
			time.Sleep(1 * time.Second)
		}
		// if err != http.ErrServerClosed {
		// 	mainLogger.Fatal("HTTP server closed unexpected.", zap.Error(err))
		// }
	}()

	// Wait for interrupt signal.
	<-ctx.Done()
	// Restore signal, allowing "force quit".
	stop()

	mainLogger.Info("Shutting down HTTP server. Press Ctrl+C to force quit.")

	// Note: We expect httpServer.Handler to handle timeouts, thus, we don't
	// need a context value with deadline here.
	//nolint:contextcheck
	// err = httpServer.Shutdown(context.Background())
	// if err != nil {
	// 	return fmt.Errorf("failed to shutdown HTTP server: %w", err)
	// }

	return nil
}
