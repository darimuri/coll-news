package coll

import (
	"bytes"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
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
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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
	collectPeriod          time.Duration
	collectType            string
	collectSource          string
	collectDirectoryPath   string
	listOutputFormat       string
	chromeBin              string
	disableHeadless        bool
	endGetIgnoreError      bool
	enableChromeLogging    bool
	listGetRetryCount      int
	chromeLoggingVerbosity int
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
	Command.Flags().StringVarP(&chromeBin, "chrome-bin", "b", "/usr/bin/chromium-browser", "chrome browser binary path")
	Command.Flags().DurationVarP(&collectPeriod, "collect-period", "p", time.Minute*10, "period between every news collection")
	Command.Flags().StringVarP(&collectType, "collect-type", "t", "", fmt.Sprintf("collect news type(%s)", coll.Types))
	Command.Flags().StringVarP(&collectSource, "collect-news-source", "s", "", fmt.Sprintf("news source(%s)", coll.Sources))
	Command.Flags().StringVarP(&collectDirectoryPath, "save-directory-path", "d", "", "save path for collected data")
	Command.Flags().StringVarP(&listOutputFormat, "list-output-format", "f", "b", fmt.Sprintf("list output format of collected news(%s)", listTypesDesc))
	Command.Flags().BoolVarP(&disableHeadless, "no-headless", "n", false, "collect news in non-headless mode")
	Command.Flags().BoolVarP(&endGetIgnoreError, "end-get-ignore-error", "e", false, "continue collect end when error occurs")
	Command.Flags().IntVarP(&listGetRetryCount, "list-get-retry-count", "l", 0, "retry count while getting list")
	Command.Flags().BoolVarP(&enableChromeLogging, "enable-chrome-logging", "", false, "run chrome using --enable-logging")
	Command.Flags().IntVarP(&chromeLoggingVerbosity, "chrome-logging-verbosity", "", 1, "run chrome using --v=1")

	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("collect-type")
	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("collect-news-source")
	//goland:noinspection GoUnhandledErrorResult
	Command.MarkFlagRequired("save-directory-path")

	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func collect() error {
	ec := echo.New()
	go func() {
		ec.Use(middleware.Recover())
		ec.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
		if err := ec.Start(":3000"); err != nil {
			panic(err)
		}
	}()

	savePath := filepath.Join(collectDirectoryPath, collectSource, collectType)

	s := make(chan os.Signal, 1)
	e := make(chan error, 1)

	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	finished := time.Now()
	nextTrigger := finished

	for {
		select {
		case sig := <-s:
			log.Println("stop collection with signal", sig)
			ec.Close()
			os.Exit(0)
		case <-time.After(time.Second):
			if false == finished.IsZero() && nextTrigger.Before(time.Now()) {
				finished = time.Time{}
				nextTrigger = time.Now().Add(collectPeriod)
				go func() {
					e <- collectAndSave(savePath)
				}()
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

func collectAndSave(rootPath string) error {
	started := nowInLocalZone()

	log.Println("collect news", collectSource, collectType, "to", rootPath)

	dumpPath := filepath.Join(rootPath, "dump", started.Format(types.FileYearFormat))
	fullDumpPath := filepath.Join(dumpPath, started.Format(types.FileDateFormat))
	listPath := filepath.Join(rootPath, "list", started.Format(types.FileYearFormat), started.Format(types.FileDateFormat))

	gzipDumpFile := filepath.Join(fullDumpPath, fmt.Sprintf("%s.%s", toFilePrefix(started), "json.gz"))

	if errMkdir := checkDirWritable(listPath); errMkdir != nil {
		return errMkdir
	}

	option := coll.Option{
		ChromeBin: chromeBin,
		SavePath:  dumpPath,
		Headless:  !disableHeadless,
		Logging:   enableChromeLogging,
		LogLevel:  chromeLoggingVerbosity,
	}

	c, errColl := coll.NewCollector(collectSource, collectType, option)
	if errColl != nil {
		return errColl
	}

	defer func() {
		if c != nil {
			c.Cleanup()
		}
	}()

	news := make([]types.News, 0)

	var listGetErrorCount int
	var collectedAt string
	var topNews, homeNews []types.News
	var err error

	for {
		log.Println("get top news list", listGetErrorCount, "<", listGetRetryCount)

		c.Top()
		collectedAt = nowInLocalZone().Format(types.DataDateTimeFormat)
		topNews, err = c.GetTopNewsList()

		if err == nil {
			break
		}

		log.Println("failed to get top news list for", err)
		time.Sleep(time.Second)

		if listGetErrorCount < listGetRetryCount {
			listGetErrorCount++
			continue
		}

		return err
	}

	listGetErrorCount = 0
	for {
		log.Println("get news home news list", listGetErrorCount, "<", listGetRetryCount)

		c.NewsHome()
		collectedAt = time.Now().Format(types.DataDateTimeFormat)
		homeNews, err = c.GetNewsHomeNewsList()

		if err == nil {
			break
		}

		log.Println("failed to get news home news list for", err)
		time.Sleep(time.Second)

		if listGetErrorCount < listGetRetryCount {
			listGetErrorCount++
			continue
		}

		return err
	}

	for i := range topNews {
		topNews[i].Location = types.Top
		topNews[i].CollectedAt = collectedAt
	}

	for i := range homeNews {
		homeNews[i].Location = types.Home
		homeNews[i].CollectedAt = collectedAt
	}

	news = append(news, topNews...)
	news = append(news, homeNews...)

	log.Printf("get %d numbers of news ends\n", len(news))

	for idx := range news {
		if err = c.GetNewsEnd(&news[idx]); err != nil {
			if false == endGetIgnoreError {
				return err

			}
			log.Printf("failed to get new end %s, but will contine with error %v\n", news[idx].Title, err)
		}

		if news[idx].End != nil {
			news[idx].End.HTML = ""
		}

		if idx > 10 && idx%10 == 1 {
			log.Printf("processed %d percent of news end\n", (idx*100)/len(news))
		}
	}

	tableRows := toTable(news)

	switch listOutputFormat {
	case listTypeTsv:
		dumpToFile(tableRows, listPath, toFilePrefix(started), "tsv", '\t', false)
	case listTypeMD:
		dumpToFile(tableRows, listPath, toFilePrefix(started), "md", '|', true)
	case listTypeBoth:
		dumpToFile(tableRows, listPath, toFilePrefix(started), "tsv", '\t', false)
		dumpToFile(tableRows, listPath, toFilePrefix(started), "md", '|', true)
	}

	byteArr, errGzip := toJsonGzipBytes(news)
	if errGzip != nil {
		return errGzip
	}

	if err = ioutil.WriteFile(gzipDumpFile, byteArr, os.FileMode(0644)); err != nil {
		return err
	}

	log.Println("collected news", collectSource, collectType, "to", collectDirectoryPath)

	return nil
}

func toJsonGzipBytes(news []types.News) ([]byte, error) {
	jsonBytes, errJson := json.Marshal(news)
	if errJson != nil {
		return nil, errJson
	}

	buffer := &bytes.Buffer{}
	gz, errGzip := gzip.NewWriterLevel(buffer, gzip.BestCompression)
	if errGzip != nil {
		return nil, errGzip
	}

	_, errGzip = gz.Write(jsonBytes)
	if errGzip != nil {
		return nil, errGzip
	}

	if errGzip = gz.Close(); errGzip != nil {
		return nil, errGzip
	}

	return buffer.Bytes(), nil
}

func toFilePrefix(t time.Time) string {
	return fmt.Sprintf("%s-%s", t.Format(types.FileDateFormat), t.Format(types.FileTimeFormat))
}

func toTable(news []types.News) [][]string {
	tableRows := make([][]string, 0)
	for idx, n := range news {
		emotions := make([]string, 0)
		author := ""
		publisher := n.Publisher
		numComment := uint64(0)
		title := strings.TrimSpace(n.Title)
		location := n.Location
		collectedAt := ""
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
				if e.CountString == "" {
					emotions = append(emotions, fmt.Sprintf("%s(%d)", e.Name, e.Count))
				} else {
					emotions = append(emotions, fmt.Sprintf("%s(%s)", e.Name, e.CountString))
				}
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
	return tableRows
}

func nowInLocalZone() time.Time {
	return time.Now().In(time.Local)
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

	if sep == '|' {
		for i := range rows {
			for j := range rows[i] {
				rows[i][j] = strings.ReplaceAll(rows[i][j], "|", "&vert;")
			}
		}
	}

	_ = table.Write(listHeader)
	if headerLine {
		_ = table.Write(listHeaderLine)
	}
	_ = table.WriteAll(rows)
	table.Flush()

	outputFile := filepath.Join(listPath, fmt.Sprintf("%s.%s", filePrefix, ext))

	if err := table.Error(); err != nil {
		log.Printf("error occured when writing list of type %s %v", ext, err)
	}

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

	if _, ok := availableListTypes[listOutputFormat]; false == ok {
		return fmt.Errorf("list output type %s is not supported", listOutputFormat)
	}

	return nil
}

func validateSavePathWritable() error {
	info, errStat := os.Stat(collectDirectoryPath)
	if errStat != nil {
		if false == os.IsNotExist(errStat) {
			return errStat
		}

		if info == nil {
			if err := os.MkdirAll(collectDirectoryPath, 0755); err != nil {
				return err
			}

			info, _ = os.Stat(collectDirectoryPath)
		}

		if false == info.IsDir() {
			return fmt.Errorf("save-path %s is not a directory", collectDirectoryPath)
		}

		if info.Mode().Perm()&(1<<(uint(7))) == 0 {
			return fmt.Errorf("write permission is not set to save-path %s", collectDirectoryPath)
		}
	}

	return nil
}
