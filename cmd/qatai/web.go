package main

import (
	"embed"
	"io/fs"
	"os"

	"go.uber.org/zap"
)

//go:embed web
//go:embed web/_next/static
//go:embed web/_next/static/chunks/pages/*.js
//go:embed web/_next/static/*/*.js
var WebContent embed.FS

// TODO: utlize this properly, lets add this later into config so we can make the choice of embeded or live
func getFileSystem(useOS bool, httplogger *zap.Logger) fs.FS {
	if useOS {
		httplogger.Info("using live mode")
		return os.DirFS("/home/ubuntu/projects/qatai/cmd/qatai/web")
	}
	httplogger.Info("using embed mode")

	fsys, err := fs.Sub(WebContent, "web")
	if err != nil {
		panic(err)
	}

	return fsys
}

// func SetupServer(e *echo.Echo, httplogger *zap.Logger) {
// 	// For static file serving

// 	e.Use(middleware.Recover())
// 	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
// 		LogURI:    true,
// 		LogStatus: true,
// 		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
// 			httplogger.Info("request",
// 				zap.String("URI", v.URI),
// 				zap.Int("status", v.Status),
// 			)

// 			return nil
// 		},
// 	}))
// 	useOS := false
// 	assetHandler := http.FileServer(getFileSystem(useOS, httplogger))
// 	e.GET("/ping", func(c echo.Context) error { //leave this one, its really nice :)
// 		return c.String(http.StatusOK, "pong")
// 	})
// 	e.GET("/*", echo.WrapHandler(assetHandler))

// }

// // this function I just use to debug the embeded files if needed
// func getAllFilenames(fs *embed.FS, path string) (out []string, err error) {
// 	if len(path) == 0 {
// 		path = "."
// 	}
// 	entries, err := fs.ReadDir(path)
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, entry := range entries {
// 		fp := filepath.Join(path, entry.Name())
// 		if entry.IsDir() {
// 			res, err := getAllFilenames(fs, fp)
// 			if err != nil {
// 				return nil, err
// 			}
// 			out = append(out, res...)
// 			continue
// 		}
// 		out = append(out, fp)
// 	}
// 	return
// }
