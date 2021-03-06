package coll

import (
	"fmt"

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
	ChromeBin string
	SavePath  string
	Headless  bool
}

func NewCollector(collectSource, collectType string, option Option) (types.Collector, error) {
	var c types.Collector
	var t types.TypedCollector
	var profile types.Profile
	var browser *rod.Browser

	l := launcher.New()
	if option.ChromeBin != "" {
		l.Bin(option.ChromeBin)
	}

	url, err := l.
		Headless(option.Headless).
		Devtools(false).
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
