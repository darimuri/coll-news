package coll

import (
	"fmt"
	"log"

	"github.com/darimuri/coll-news/pkg/cache"
	"github.com/darimuri/coll-news/pkg/daum"
	dmobile "github.com/darimuri/coll-news/pkg/daum/mobile"
	dpc "github.com/darimuri/coll-news/pkg/daum/pc"
	"github.com/darimuri/coll-news/pkg/types"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/devices"
	"github.com/go-rod/rod/lib/launcher"
)

const (
	Daum    = "daum"
	Naver   = "naver"
	Sources = "daum/naver"

	Mobile = "mobile"
	PC     = "pc"
	Types  = "mobile/pc"
)

type Option struct {
	ChromeBin   string
	SavePath    string
	UserDataDir string
	LogLevel    int
	Headless    bool
	Logging     bool
}

func NewCollector(collectSource, collectType string, option Option) (types.Collector, error) {
	var c types.Collector
	var t types.TypedCollector
	var profile types.Profile
	var browser *rod.Browser

	log.Println("new collector with option", option)

	l := launcher.New()
	if option.UserDataDir != "" {
		l.UserDataDir(option.UserDataDir)
	}
	if option.ChromeBin != "" {
		l.Bin(option.ChromeBin)
	}
	if option.Logging {
		l.Set("enable-logging")
		if option.LogLevel > 0 {
			l.Set("v", fmt.Sprintf("%d", option.LogLevel))
		}
	}

	url, err := l.
		Headless(option.Headless).
		Devtools(false).
		Set("no-sandbox").
		//Set("disable-gpu").
		Set("disable-dev-shm-usage").
		Set("no-zygote").
		Set("single-process").
		Set("start-maximized").
		Launch()

	if err != nil {
		return nil, err
	}

	switch collectType {
	case PC:
		browser = rod.New()
		profile = types.PC()
	case Mobile:
		browser = rod.New().DefaultDevice(devices.IPhone6or7or8)
		profile = types.Mobile()
	default:
		return nil, fmt.Errorf("collector type %s is not supported", collectType)
	}

	err = browser.
		ControlURL(url).
		Connect()

	switch collectSource {
	case Daum:
		switch collectType {
		case Mobile:
			t = dmobile.New()
		case PC:
			t = dpc.New()
		}
		c, err = daum.NewPortal(browser, profile, t, option.SavePath, cache.NewLargeCache())
	case Naver:
	}

	if err != nil {
		return nil, err
	} else if c == nil {
		return nil, fmt.Errorf("collector source %s, type %s is not supported", collectSource, collectType)
	}

	return c, nil
}
