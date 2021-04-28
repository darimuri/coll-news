package coll

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/darimuri/coll-news/pkg/coll"
	"github.com/darimuri/coll-news/pkg/types"
	"github.com/spf13/cobra"
)

var (
	collectPeriod   time.Duration
	collectType     string
	collectSource   string
	collectSavePath string
	disableHeadless bool
)

var Command = &cobra.Command{
	Use:   "coll",
	Short: "Collect portal news in a given period",
	RunE: func(cmd *cobra.Command, args []string) error {
		return collect()
	},

	Args: func(cmd *cobra.Command, args []string) error {
		return validateFlags()
	},
}

func init() {
	Command.Flags().DurationVarP(&collectPeriod, "period", "p", time.Minute*10, "period between every news collection")
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
	res := make(chan error, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	var finished time.Time

	log.Println("start collect news for period", collectPeriod)
	nextTrigger := time.Now().Add(collectPeriod)
	go collectAndSave(c, res)
	for {
		select {
		case sig := <-sigs:
			log.Println("stop collection with signal", sig)
			os.Exit(0)
		case <-time.After(time.Second):
			if false == finished.IsZero() && nextTrigger.Before(time.Now()) {
				finished = time.Time{}
				nextTrigger = time.Now().Add(collectPeriod)
				go collectAndSave(c, res)
			}
		case collErr := <-res:
			if collErr != nil {
				log.Println("failed to collect for error", collErr.Error())
				os.Exit(1)
			}
			if nextTrigger.After(time.Now()) {
				log.Println("next collection will start at", nextTrigger.Format("2006/01/02 15:04:05"))
			}
			finished = time.Now()

		}
	}

	return nil
}

func collectAndSave(c types.Collector, res chan error) {
	log.Println("collect news", collectSource, collectType, "to", collectSavePath)

	news := make([]types.News, 0)

	log.Println("get top news list")

	c.Top()
	topNews, errTop := c.GetTopNewsList()
	if errTop != nil {
		res <- errTop
		return
	}

	news = append(news, topNews...)

	log.Println("get news home news list")

	c.NewsHome()
	homeNews, errHome := c.GetNewsHomeNewsList()
	if errHome != nil {
		res <- errHome
		return
	}

	news = append(news, homeNews...)

	log.Println("get news ends", len(news))

	for idx := range news {
		if e := c.GetNewsEnd(&news[idx]); e != nil {
			res <- e
			return
		}
	}

	for idx, n := range news {
		emotions := make([]string, 0)
		author := ""
		publisher := n.Publisher
		numComment := uint64(0)
		title := strings.TrimSpace(n.Title)
		postedAt := ""
		modifiedAt := ""

		if n.End != nil {
			author = n.End.Author
			publisher = n.End.Provider
			numComment = n.End.NumComment
			postedAt = n.End.PostedAt
			modifiedAt = n.End.ModifiedAt

			for _, e := range n.End.Emotions {
				emotions = append(emotions, fmt.Sprintf("%s(%d)", e.Name, e.Count))
			}
		}

		author = strings.TrimSpace(author)
		publisher = strings.TrimSpace(publisher)

		fmt.Printf("%3d|%d|%v|%v|%v|%v|%v|%v\n", idx, numComment, author, publisher, emotions, title, postedAt, modifiedAt)
	}

	log.Println("collected news", collectSource, collectType, "to", collectSavePath)

	c.Cleanup()

	res <- nil
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
