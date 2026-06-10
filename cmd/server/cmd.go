package server

import (
	"context"
	"embed"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/Guo-Chenxu/pay-log/config"
	"github.com/Guo-Chenxu/pay-log/dal"
	redisdb "github.com/Guo-Chenxu/pay-log/dal/redis"
	"github.com/Guo-Chenxu/pay-log/pkg/auth"
	"github.com/Guo-Chenxu/pay-log/pkg/logger"
	"github.com/Guo-Chenxu/pay-log/pkg/middleware"
	"github.com/Guo-Chenxu/pay-log/pkg/snowflake"
	"github.com/Guo-Chenxu/pay-log/router"
	"github.com/Guo-Chenxu/pay-log/scheduler"
)

var (
	configFile string
	staticFS   embed.FS
)

var rootCmd = &cobra.Command{
	Use:   "pay-log",
	Short: "Bill parsing and summarization tool",
	Run:   run,
}

func init() {
	rootCmd.Flags().StringVarP(&configFile, "config", "f", "config.yaml", "config file path")
}

// Execute accepts the embedded static filesystem from main and starts the server.
func Execute(files embed.FS) {
	staticFS = files
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	config.Init(configFile)
	logger.Init(config.GetLoggerConfig())

	if err := snowflake.Init(1); err != nil {
		panic(fmt.Sprintf("snowflake init failed: %v", err))
	}

	dal.Init()

	ctx := context.Background()
	auth.Init(ctx, auth.NewRedisStorage(redisdb.GetClient()), config.GetAuthConfig())

	r := gin.New()
	r.Use(middleware.Cors())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.ErrorHandlerMiddleware())
	r.Use(middleware.AuthMiddleware())

	// serve embedded frontend under /
	sub, err := fs.Sub(staticFS, "static")
	if err == nil {
		r.NoRoute(func(c *gin.Context) {
			p := c.Request.URL.Path
			// API routes and real asset files get a real 404
			if strings.HasPrefix(p, "/api/") {
				c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "not found"})
				return
			}
			// Try to serve the file; if it doesn't exist serve index.html (SPA fallback)
			f, err := http.FS(sub).Open(strings.TrimPrefix(p, "/"))
			if err == nil {
				f.Close()
				http.FileServer(http.FS(sub)).ServeHTTP(c.Writer, c.Request)
				return
			}
			// SPA fallback: serve index.html
			idx, err := sub.Open("index.html")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			defer idx.Close()
			stat, _ := idx.Stat()
			c.Header("Content-Type", "text/html; charset=utf-8")
			http.ServeContent(c.Writer, c.Request, "index.html", stat.ModTime(), idx.(io.ReadSeeker))
		})
	}

	router.Register(r)
	scheduler.Start()

	addr := fmt.Sprintf(":%d", config.GetServerConfig().Port)
	logger.Infof("server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		panic(err)
	}
}
