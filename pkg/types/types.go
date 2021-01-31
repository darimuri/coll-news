package types

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"
)

type News struct {
	URL            string `json:"url"`
	Image          string `json:"image,omitempty"`
	Title          string `json:"title"`
	NewsPage       int    `json:"news_page"`
	Order          int    `json:"order"`
	SubOrder       int    `json:"sub_order"`
	FullHTML       string `json:"full_html"`
	FullScreenShot string `json:"full_screen_shot"`
	TabScreenShot  string `json:"tab_screen_shot"`
	Publisher      string `json:"publisher"`
	End            *End   `json:"end"`
}

type End struct {
	Category   string   `json:"category,omitempty"`
	Provider   string   `json:"provider,omitempty"`
	Title      string   `json:"title"`
	Author     string   `json:"author"`
	PostedAt   string   `json:"posted_at"`
	ModifiedAt string   `json:"modified_at,omitempty"`
	NumComment uint64   `json:"num_comment,omitempty"`
	Text       string   `json:"text"`
	HTML       string   `json:"html"`
	Images     []string `json:"images,omitempty"`
	Program    string   `json:"program,omitempty"`
	NumPlayed  uint64   `json:"num_played"`
}

func (n *News) ToString() string {
	m, err := json.Marshal(n)
	if err != nil {
		panic(err)
	}
	return string(m)
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
	d.dumpPath = path.Join(d.RootPath, d.DumpTime.Format("20060102"), d.Source)
	d.dumpPrefix = d.DumpTime.Format("150405.000000000")
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
