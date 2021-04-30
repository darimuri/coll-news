package coll

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/darimuri/coll-news/pkg/coll"
	"github.com/darimuri/coll-news/pkg/types"
	"github.com/spf13/cobra"
	"golang.org/x/sys/unix"
)

const (
	listTypesDesc = "t(tsv)/m(markdown table)/b(both)"

	listTypeBoth = "b"
	listTypeTsv  = "t"
	listTypeMD   = "m"
)

var availableListTypes = map[string]string{
	listTypeBoth: listTypeBoth,
	listTypeTsv:  listTypeTsv,
	listTypeMD:   listTypeMD,
}

var (
	listHeader = []string{
		"No.",
		"NumComment",
		"Author",
		"Publisher",
		"Category",
		"Title",
		"Location",
		"CollectedAt",
		"PostedAt",
		"ModifiedAt",
		"Emotions",
		"URL",
	}
	listHeaderLine = []string{
		"---",
		"---",
		"---",
		"---",
		"---",
		"---",
		"---",
		"---",
		"---",
		"---",
		"---",
		"---",
	}
)

var (
	collectPeriod   time.Duration
	collectType     string
	collectSource   string
	collectSavePath string
	listOutputType  string
	chromeBin       string
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
	Command.Flags().StringVarP(&chromeBin, "chrome-bin", "b", "/usr/bin/chromium-browser", "chrome browser binary path")
	Command.Flags().StringVarP(&collectSavePath, "save-path", "s", "", "save path for collected data")
	Command.Flags().StringVarP(&listOutputType, "list-output", "o", "b", fmt.Sprintf("list output of collected news(%s)", listTypesDesc))
	Command.Flags().BoolVarP(&disableHeadless, "no-headless", "", false, "collect news in non-headless mode")

	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("type")
	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("news-source")
	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("save-path")
}

func collect() error {
	savePath := filepath.Join(collectSavePath, collectSource, collectType)

	s := make(chan os.Signal, 1)
	e := make(chan error, 1)

	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	var finished time.Time

	log.Println("start collect news for period", collectPeriod)
	nextTrigger := time.Now().Add(collectPeriod)
	go collectAndSave(e, savePath)
	for {
		select {
		case sig := <-s:
			log.Println("stop collection with signal", sig)
			os.Exit(0)
		case <-time.After(time.Second):
			if false == finished.IsZero() && nextTrigger.Before(time.Now()) {
				finished = time.Time{}
				nextTrigger = time.Now().Add(collectPeriod)
				go collectAndSave(e, savePath)
			}
		case collErr := <-e:
			if collErr != nil {
				log.Println("failed to collect for error", collErr.Error())
				os.Exit(1)
			}
			if nextTrigger.After(time.Now()) {
				log.Println("next collection will start at", nextTrigger.Format(types.LogDateTimeFormat))
			}
			finished = time.Now()

		}
	}

	return nil
}

func collectAndSave(res chan error, savePath string) {
	now := time.Now()

	log.Println("collect news", collectSource, collectType, "to", savePath)

	dumpPath := filepath.Join(savePath, "dump", now.Format(types.FileYearFormat))
	listPath := filepath.Join(savePath, "list", now.Format(types.FileYearFormat), now.Format(types.FileDateFormat))

	if errMkdir := checkDirWritable(listPath); errMkdir != nil {
		res <- errMkdir
		return
	}

	c, errColl := coll.NewCollector(collectSource, collectType, coll.Option{
		ChromeBin: chromeBin,
		SavePath:  dumpPath,
		Headless:  !disableHeadless,
	})
	if errColl != nil {
		res <- errColl
		return
	}

	defer func() {
		if c != nil {
			c.Cleanup()
		}
	}()

	news := make([]types.News, 0)

	log.Println("get top news list")

	c.Top()
	collectedAt := time.Now().Format(types.DataDateTimeFormat)

	topNews, errTop := c.GetTopNewsList()
	if errTop != nil {
		res <- errTop
		return
	}

	for i := range topNews {
		topNews[i].CollectedAt = collectedAt
		topNews[i].Location = types.Top
	}

	news = append(news, topNews...)

	log.Println("get news home news list")

	c.NewsHome()
	collectedAt = time.Now().Format(types.DataDateTimeFormat)

	homeNews, errHome := c.GetNewsHomeNewsList()
	if errHome != nil {
		res <- errHome
		return
	}

	for i := range homeNews {
		homeNews[i].CollectedAt = collectedAt
		homeNews[i].Location = types.Home
	}

	news = append(news, homeNews...)

	log.Printf("get %d numbers of news ends\n", len(news))

	for idx := range news {
		if e := c.GetNewsEnd(&news[idx]); e != nil {
			res <- e
			return
		}
		if idx > 10 && idx%10 == 1 {
			log.Printf("processed %d percent of news end\n", (idx*100)/len(news))
		}
	}

	tableRows := make([][]string, 0)
	for idx, n := range news {
		emotions := make([]string, 0)
		author := ""
		publisher := n.Publisher
		numComment := uint64(0)
		title := strings.TrimSpace(n.Title)
		location := n.Location
		collectedAt = ""
		postedAt := ""
		modifiedAt := ""
		category := ""

		if n.End != nil {
			author = n.End.Author
			publisher = n.End.Provider
			numComment = n.End.NumComment
			category = n.End.Category
			collectedAt = n.End.CollectedAt
			postedAt = n.End.PostedAt
			modifiedAt = n.End.ModifiedAt

			for _, e := range n.End.Emotions {
				emotions = append(emotions, fmt.Sprintf("%s(%d)", e.Name, e.Count))
			}
		}

		if author == "" {
			author = "-"
		}

		if publisher == "" {
			publisher = "-"
		}

		if category == "" {
			category = "-"
		}

		if postedAt == "" {
			postedAt = "-"
		}

		if modifiedAt == "" {
			modifiedAt = "-"
		}

		author = strings.TrimSpace(author)
		publisher = strings.TrimSpace(publisher)

		row := []string{
			fmt.Sprintf("%d", idx),
			fmt.Sprintf("%d", numComment),
			author,
			publisher,
			category,
			title,
			string(location),
			collectedAt,
			postedAt,
			modifiedAt,
			emotionsToString(emotions),
			n.URL,
		}

		tableRows = append(tableRows, row)
	}

	filePrefix := fmt.Sprintf("%s-%s", now.Format(types.FileDateFormat), now.Format(types.FileTimeFormat))

	switch listOutputType {
	case listTypeTsv:
		dumpToFile(tableRows, listPath, filePrefix, "tsv", '\t', false)
	case listTypeMD:
		dumpToFile(tableRows, listPath, filePrefix, "md", '|', true)
	case listTypeBoth:
		dumpToFile(tableRows, listPath, filePrefix, "tsv", '\t', false)
		dumpToFile(tableRows, listPath, filePrefix, "md", '|', true)
	}

	log.Println("collected news", collectSource, collectType, "to", collectSavePath)

	res <- nil
}

func emotionsToString(emotions []string) string {
	if len(emotions) == 0 {
		return "-"
	}
	return fmt.Sprintf("%v", emotions)
}

func checkDirWritable(dir string) error {
	for {
		s, errStat := os.Stat(dir)
		if errStat == nil {
			if false == s.IsDir() {
				return fmt.Errorf("%s should be a directory", dir)
			} else if runtime.GOOS == "linux" && unix.Access(dir, unix.W_OK) != nil {
				return fmt.Errorf("directory %s should be writable", dir)
			}
			break
		} else if true == os.IsNotExist(errStat) {
			if errMkdir := os.MkdirAll(dir, os.FileMode(0700)); errMkdir != nil {
				return errMkdir
			}
		} else {
			return errStat
		}
	}

	return nil
}

func dumpToFile(rows [][]string, listPath, filePrefix, ext string, sep rune, headerLine bool) {
	buffer := &bytes.Buffer{}
	table := csv.NewWriter(buffer)
	table.Comma = sep
	_ = table.Write(listHeader)
	if headerLine {
		_ = table.Write(listHeaderLine)
	}
	_ = table.WriteAll(rows)
	table.Flush()

	outputFile := filepath.Join(listPath, fmt.Sprintf("%s.%s", filePrefix, ext))
	if err := ioutil.WriteFile(outputFile, buffer.Bytes(), os.FileMode(0600)); err != nil {
		log.Println("failed to write to", outputFile, "for", err.Error())
	}
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

	if _, ok := availableListTypes[listOutputType]; false == ok {
		return fmt.Errorf("list output type %s is not supported", listOutputType)
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
