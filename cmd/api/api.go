package api

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/cobra"

	"github.com/darimuri/coll-news/pkg/handlers/summary"
	"github.com/darimuri/coll-news/pkg/handlers/wordcloud"
	"github.com/darimuri/coll-news/pkg/util"
)

var Command = &cobra.Command{
	Use:   "api",
	Short: "api server for coll",
	RunE: func(cmd *cobra.Command, args []string) error {
		return server()
	},

	Args: func(cmd *cobra.Command, args []string) error {
		return validateFlags()
	},
}

var (
	collectDirectoryPath string
	metricsPort          int
)

func init() {
	Command.Flags().StringVarP(&collectDirectoryPath, "collect-directory", "d", "coll_dir", "collect directory path for collected data")
	Command.Flags().IntVarP(&metricsPort, "port", "", 8080, "port to listen http request")

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func server() error {
	collectPath := filepath.Join(collectDirectoryPath)

	st, err := os.Stat(collectPath)
	if err != nil {
		panic(err)
	} else if !st.IsDir() {
		panic(fmt.Errorf("%s is not directory", collectPath))
		os.Exit(1)
	}

	ec := echo.New()
	ec.Use(middleware.Recover())
	ec.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
	ec.GET("/api/summary", summary.NewSummary(collectPath).Handle)
	ec.POST("/api/wordcloud", wordcloud.NewWordCloud(collectPath).Handle)
	if err = ec.Start(fmt.Sprintf(":%d", metricsPort)); err != nil {
		panic(err)
	}

	return nil
}

func validateFlags() error {
	if err := validateCollectPath(); err != nil {
		return err
	}

	return nil
}

func validateCollectPath() error {
	info, errStat := os.Stat(collectDirectoryPath)
	if errStat != nil {
		if os.IsNotExist(errStat) {
			return fmt.Errorf("collect-directory %s does not exist", collectDirectoryPath)
		}
		return errStat

	}

	if false == info.IsDir() {
		return fmt.Errorf("collect-directory %s is not a directory", collectDirectoryPath)
	}

	if err := util.SyscallAccessRead(collectDirectoryPath); err != nil {
		return fmt.Errorf("read permission is not set to collect-path %s", collectDirectoryPath)
	}

	return nil
}
