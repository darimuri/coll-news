package coll

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/darimuri/coll-news/pkg/coll"
	"github.com/darimuri/coll-news/pkg/types"
	"github.com/spf13/cobra"
)

var (
	collectDuration time.Duration
	collectType     string
	collectSource   string
	collectSavePath string
	disableHeadless bool
)

var Command = &cobra.Command{
	Use:   "coll",
	Short: "Collect portal news in a given duration",
	RunE: func(cmd *cobra.Command, args []string) error {
		return collect()
	},

	Args: func(cmd *cobra.Command, args []string) error {
		return validateFlags()
	},
}

func init() {
	Command.Flags().DurationVarP(&collectDuration, "period", "p", time.Minute, "period between every news collection")
	Command.Flags().StringVarP(&collectType, "type", "t", "", fmt.Sprintf("collect news type(%s)", coll.Types))
	Command.Flags().StringVarP(&collectSource, "news-source", "n", "", fmt.Sprintf("news source(%s)", coll.Sources))
	Command.Flags().StringVarP(&collectSavePath, "save-path", "s", "", "save path for collected data")
	Command.Flags().BoolVarP(&disableHeadless, "no-headless", "", false, "collect news in non-headless mode")

	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("type")
	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("news-source")
	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("save-path")
}

func collect() error {
	c, err := coll.NewCollector(collectSource, collectType, coll.Option{
		SavePath: collectSavePath,
		Headless: !disableHeadless,
	})
	if err != nil {
		return err
	}

	sigs := make(chan os.Signal, 1)
	collError := make(chan error, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case sig := <-sigs:
			log.Println("stop collection with signal", sig)
			os.Exit(0)
		case <-time.After(collectDuration):
			collectAndSave(c, collError)
		case ce := <-collError:
			if ce != nil {
				fmt.Println("failed to collect for error", ce.Error())
				os.Exit(1)
			}
		}
	}

	return nil
}

func collectAndSave(c types.Collector, collError chan error) {
	log.Println("collect")

	c.Top()
	news, err := c.GetTopNewsList()
	if err != nil {
		collError <- err
		return
	}

	for _, n := range news {
		if e := c.GetNewsEnd(&n); e != nil {
			collError <- e
			return
		}
	}

	c.NewsHome()
	news, err = c.GetNewsHomeNewsList()
	if err != nil {
		collError <- err
		return
	}

	for _, n := range news {
		if e := c.GetNewsEnd(&n); e != nil {
			collError <- e
			return
		}
	}

	log.Println("collected")

	c.Cleanup()
}

func validateFlags() error {
	switch collectType {
	case coll.PC, coll.Mobile:
	default:
		return fmt.Errorf("type should be %s. not %s", coll.Types, collectType)
	}

	switch collectSource {
	case coll.Daum, coll.Naver:
	default:
		return fmt.Errorf("news-source should be %s. not %s", coll.Sources, collectSource)
	}

	if err := validateSavePathWritable(); err != nil {
		return err
	}

	return nil
}

func validateSavePathWritable() error {
	info, errStat := os.Stat(collectSavePath)
	if errStat != nil {
		if false == os.IsNotExist(errStat) {
			return errStat
		}

		if info == nil {
			if err := os.MkdirAll(collectSavePath, 0755); err != nil {
				return err
			}

			info, _ = os.Stat(collectSavePath)
		}

		if false == info.IsDir() {
			return fmt.Errorf("save-path %s is not a directory", collectSavePath)
		}

		if info.Mode().Perm()&(1<<(uint(7))) == 0 {
			return fmt.Errorf("write permission is not set to save-path %s", collectSavePath)
		}
	}

	return nil
}
