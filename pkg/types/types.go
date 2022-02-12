package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
)

const (
	DataDateTimeFormat = time.RFC3339
	LogDateTimeFormat  = "2006/01/02 15:04:05"
	FileYearFormat     = "2006"
	FileDateFormat     = "20060102"
	FileTimeFormat     = "150405"
	FileTimeNanoFormat = "150405.000000000"
)

type Loc string

const (
	Top  = "Top"
	Home = "Home"
)

type News struct {
	URL            string `json:"url"`
	Image          string `json:"image,omitempty"`
	Title          string `json:"title"`
	Category       string `json:"category,omitempty"`
	SeriesTitle    string `json:"series_title,omitempty"`
	NewsPage       int    `json:"news_page"`
	Order          int    `json:"order"`
	SubOrder       int    `json:"sub_order"`
	FullHTML       string `json:"full_html"`
	FullScreenShot string `json:"full_screen_shot"`
	TabScreenShot  string `json:"tab_screen_shot"`
	Publisher      string `json:"publisher"`
	Location       Loc    `json:"loc"`
	CollectedAt    string `json:"collected_at"`
	End            *End   `json:"end"`
}

type End struct {
	Category    string    `json:"category,omitempty"`
	Provider    string    `json:"provider,omitempty"`
	Title       string    `json:"title"`
	Author      string    `json:"author"`
	CollectedAt string    `json:"collected_at"`
	PostedAt    string    `json:"posted_at"`
	ModifiedAt  string    `json:"modified_at,omitempty"`
	NumComment  uint64    `json:"num_comment,omitempty"`
	Emotions    []Emotion `json:"emotions,omitempty"`
	Text        string    `json:"text"`
	HTML        string    `json:"html,omitempty"`
	Images      []string  `json:"images,omitempty"`
	Program     string    `json:"program,omitempty"`
	NumPlayed   uint64    `json:"num_played"`
}

type Emotion struct {
	Name        string
	CountString string
	Count       int64
}

func (n *News) ToString() string {
	m, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	return string(m)
}

func (n *News) SetContextData(page int, order int, subOrder int, dd DumpDirectory, tabScreenShot bool) {
	n.Order = order
	n.SubOrder = subOrder
	n.NewsPage = page
	n.FullHTML = dd.FullHTML()
	n.FullScreenShot = dd.FullScreenShot()

	if tabScreenShot {
		n.TabScreenShot = dd.TabScreenShot(page)
	}
}

type DumpDirectory struct {
	RootPath string
	Source   string
	DumpTime time.Time

	dumpPath   string
	dumpPrefix string

	fullScreenShot string
	fullHTML       string
}

func (d *DumpDirectory) Init() error {
	d.dumpPath = path.Join(d.RootPath, d.DumpTime.Format(FileDateFormat), d.Source)
	d.dumpPrefix = d.DumpTime.Format(FileTimeNanoFormat)
	d.fullScreenShot = path.Join(d.dumpPath, fmt.Sprintf("%s.00.jpg", d.dumpPrefix))
	d.fullHTML = path.Join(d.dumpPath, fmt.Sprintf("%s.00.html", d.dumpPrefix))

	info, err := os.Stat(d.dumpPath)
	if err != nil {
		if false == os.IsNotExist(err) {
			return err
		}

		if errMkdir := os.MkdirAll(d.dumpPath, os.ModePerm); errMkdir != nil {
			return errMkdir
		}
	}

	info, err = os.Stat(d.dumpPath)
	if err != nil {
		return err
	}

	if false == info.IsDir() {
		return fmt.Errorf("dump directory %s is not directory", d.dumpPath)
	}

	return nil
}

func (d *DumpDirectory) FullScreenShot() string {
	return d.fullScreenShot
}

func (d *DumpDirectory) FullHTML() string {
	return d.fullHTML
}

func (d *DumpDirectory) TabScreenShot(tabNum int) string {
	return path.Join(d.dumpPath, fmt.Sprintf("%s.tab.%02d.jpg", d.dumpPrefix, tabNum))
}

type Profile struct {
	Width  int
	Height int

	TypedCollector
}

func PC() Profile {
	return Profile{
		Width:  1920,
		Height: 1080,
	}
}

func Mobile() Profile {
	return Profile{
		Width:  640,
		Height: 1080,
	}
}
